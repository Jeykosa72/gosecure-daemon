package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gosecure-daemon/internal/config"
	"gosecure-daemon/internal/monitor"
	"gosecure-daemon/internal/process"
	"gosecure-daemon/internal/worker"
)

func main() {
	log.Println("=====================================================")
	log.Println(" Starting GoSecure-Daemon (Secure System Monitor) ")
	log.Println("=====================================================")

	// 1. Memuat Konfigurasi Utama
	cfg := config.LoadDefaultConfig()

	// 2. Inisialisasi Context dengan OS Signal Handling (Graceful Shutdown)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 3. Menginisialisasi Komponen Modul
	runner := process.NewSecureRunner(cfg.ExecTimeout)
	pool := worker.NewWorkerPool(cfg.WorkerCount, cfg.QueueSize, runner, cfg.SecurityCmd)

	watcher, err := monitor.NewFileWatcher(cfg.WatchDir, pool)
	if err != nil {
		log.Fatalf("[FATAL] Gagal menginisialisasi FileWatcher: %v", err)
	}

	// 4. Memulai Worker Pool & Watcher secara Asinkron
	pool.Start(ctx)
	if err := watcher.StartWatch(ctx); err != nil {
		log.Fatalf("[FATAL] Gagal menjalankan pengawasan direktori: %v", err)
	}

	log.Println("[Main] Daemon berjalan stabil. Tekan Ctrl+C atau kirim SIGTERM untuk menghentikan.")

	// 5. Blocking sampai sinyal OS diterima
	<-ctx.Done()

	log.Println("\n[Main] Sinyal penghentian diterima! Memulai alur Graceful Shutdown...")

	// 6. Alur Terkontrol Shutdown dengan Hard Limit Timeout 5 Detik
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		pool.Stop() // Menunggu worker menyelesaikan job aktif hingga selesai
		close(done)
	}()

	select {
	case <-done:
		log.Println("[Main] Graceful shutdown selesai dengan sempurna. Daemon keluar bersih.")
	case <-shutdownCtx.Done():
		log.Println("[Main] WARNING: Shutdown melebihi batas 5s, menghentikan aplikasi secara paksa.")
	}
}