package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gosecure-daemon/internal/config"
	"gosecure-daemon/internal/logger"
	"gosecure-daemon/internal/monitor"
	"gosecure-daemon/internal/process"
	"gosecure-daemon/internal/worker"
)

func main() {
	// Inisialisasi Logger Terstruktur (slog)
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	log := logger.SetupLogger(logLevel)
	log.Info("Memulai GoSecure-Daemon", slog.String("version", "1.0.0"))

	// Memuat Konfigurasi
	cfg := config.LoadDefaultConfig()
	if envDir := os.Getenv("WATCH_DIR"); envDir != "" {
		cfg.WatchDir = envDir
	}

	// Setup Context dengan Signal Handling untuk Graceful Shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Inisialisasi Komponen Modul dengan Dependency Injection Logger
	runner := process.NewSecureRunner(cfg.ExecTimeout)
	pool := worker.NewWorkerPool(cfg.WorkerCount, cfg.QueueSize, runner, cfg.SecurityCmd, log)
	watcher, err := monitor.NewFileWatcher(cfg.WatchDir, pool, log)
	if err != nil {
		log.Error("Gagal menginisialisasi FileWatcher", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Jalankan Worker Pool & Watcher secara Asinkron
	pool.Start(ctx)
	if err := watcher.StartWatch(ctx); err != nil {
		log.Error("Gagal menjalankan watcher", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("Daemon berjalan penuh", slog.String("watch_dir", cfg.WatchDir))

	// Blok eksekusi hingga menerima sinyal penghentian OS
	<-ctx.Done()
	log.Info("Sinyal OS diterima, memulai Graceful Shutdown...")

	// Alur Shutdown Terkontrol dengan Batas Waktu 5 Detik
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		pool.Stop()
		close(done)
	}()

	select {
	case <-done:
		log.Info("Graceful shutdown selesai. Daemon berhenti bersih.")
	case <-shutdownCtx.Done():
		log.Warn("Shutdown melebihi batas 5s, menghentikan daemon secara paksa.")
	}
}
