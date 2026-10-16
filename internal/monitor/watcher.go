package monitor

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"

	"gosecure-daemon/internal/worker"
)

// FileWatcher memantau sistem berkas secara real-time menggunakan fsnotify/inotify.
type FileWatcher struct {
	watchDir string
	watcher  *fsnotify.Watcher
	pool     *worker.WorkerPool
}

func NewFileWatcher(watchDir string, pool *worker.WorkerPool) (*FileWatcher, error) {
	// Menerapkan Least Privilege: Pastikan direktori ada dengan izin ketat 0750
	if _, err := os.Stat(watchDir); os.IsNotExist(err) {
		if err := os.MkdirAll(watchDir, 0750); err != nil {
			return nil, fmt.Errorf("gagal membuat direktori target: %w", err)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi fsnotify: %w", err)
	}

	return &FileWatcher{
		watchDir: watchDir,
		watcher:  watcher,
		pool:     pool,
	}, nil
}

// StartWatch menjalankan event loop pengawasan berkas dalam goroutine terpisah.
func (w *FileWatcher) StartWatch(ctx context.Context) error {
	if err := w.watcher.Add(w.watchDir); err != nil {
		return fmt.Errorf("gagal mendaftarkan direktori ke watcher: %w", err)
	}
	log.Printf("[Monitor] Memulai pengawasan direktori: %s", w.watchDir)

	go func() {
		defer w.watcher.Close()
		for {
			select {
			case <-ctx.Done():
				log.Println("[Monitor] Sinyal shutdown diterima, menghentikan watcher...")
				return
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				// Menyaring hanya event pembuatan (Create) dan modifikasi (Write)
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					job := worker.Job{
						FilePath:  event.Name,
						EventType: event.Op.String(),
					}
					w.pool.Submit(ctx, job)
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[Monitor] ERROR watcher event: %v", err)
			}
		}
	}()

	return nil
}