package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/fsnotify/fsnotify"

	"gosecure-daemon/internal/worker"
)

type FileWatcher struct {
	watchDir string
	watcher  *fsnotify.Watcher
	pool     *worker.WorkerPool
	logger   *slog.Logger
}

func NewFileWatcher(watchDir string, pool *worker.WorkerPool, logger *slog.Logger) (*FileWatcher, error) {
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
		logger:   logger,
	}, nil
}

func (w *FileWatcher) StartWatch(ctx context.Context) error {
	if err := w.watcher.Add(w.watchDir); err != nil {
		return fmt.Errorf("gagal mendaftarkan direktori: %w", err)
	}

	w.logger.Info("Memulai pengawasan direktori", slog.String("watch_dir", w.watchDir))

	go func() {
		defer w.watcher.Close()
		for {
			select {
			case <-ctx.Done():
				w.logger.Info("Menghentikan pengawasan direktori...")
				return
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

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
				w.logger.Error("Error pada event watcher", slog.String("error", err.Error()))
			}
		}
	}()

	return nil
}