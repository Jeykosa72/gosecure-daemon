package worker_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gosecure-daemon/internal/process"
	"gosecure-daemon/internal/worker"
)

func TestWorkerPool_ConcurrentSubmit(t *testing.T) {
	// Buat direktori dan file sementara untuk pengujian
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Gagal membuat file temp: %v", err)
	}

	// Inisialisasi runner, logger, dan worker pool
	runner := process.NewSecureRunner(2 * time.Second)
	testLogger := slog.Default()
	pool := worker.NewWorkerPool(4, 50, runner, "echo", testLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)

	// Simulasikan submit job serentak dari 5 goroutine berbeda (menguji data race)
	done := make(chan struct{})
	for i := 0; i < 5; i++ {
		go func(routineID int) {
			for j := 0; j < 10; j++ {
				job := worker.Job{
					FilePath:  tmpFile,
					EventType: fmt.Sprintf("WRITE_%d_%d", routineID, j),
				}
				pool.Submit(ctx, job)
			}
			done <- struct{}{}
		}(i)
	}

	// Menunggu seluruh goroutine pengirim selesai
	for i := 0; i < 5; i++ {
		<-done
	}

	// Beri jeda singkat lalu hentikan worker pool secara bersih
	time.Sleep(300 * time.Millisecond)
	pool.Stop()
}