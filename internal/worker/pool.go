package worker

import (
	"context"
	"log/slog"
	"sync"

	"gosecure-daemon/internal/process"
)

type Job struct {
	FilePath  string
	EventType string
}

type WorkerPool struct {
	workerCount int
	jobQueue    chan Job
	runner      *process.SecureRunner
	cmd         string
	wg          sync.WaitGroup
	logger      *slog.Logger
}

func NewWorkerPool(workerCount int, queueSize int, runner *process.SecureRunner, cmd string, logger *slog.Logger) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, queueSize),
		runner:      runner,
		cmd:         cmd,
		logger:      logger,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
	p.logger.Info("WorkerPool berhasil dimulai", slog.Int("worker_count", p.workerCount))
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	p.logger.Debug("Worker siap menerima tugas", slog.Int("worker_id", id))

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("Worker menghentikan operasi", slog.Int("worker_id", id))
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				return
			}
			p.logger.Info("Memproses file event",
				slog.Int("worker_id", id),
				slog.String("file_path", job.FilePath),
				slog.String("event_type", job.EventType),
			)

			result, err := p.runner.AnalyzeFile(ctx, p.cmd, job.FilePath)
			if err != nil {
				p.logger.Error("Analisis file gagal",
					slog.Int("worker_id", id),
					slog.String("file_path", job.FilePath),
					slog.String("error", err.Error()),
				)
			} else {
				p.logger.Info("Analisis file berhasil",
					slog.Int("worker_id", id),
					slog.String("file_path", job.FilePath),
					slog.String("output", result),
				)
			}
		}
	}
}

func (p *WorkerPool) Submit(ctx context.Context, job Job) bool {
	select {
	case <-ctx.Done():
		return false
	case p.jobQueue <- job:
		return true
	default:
		p.logger.Warn("Antrean worker penuh, mengabaikan event",
			slog.Int("queue_len", len(p.jobQueue)),
			slog.String("file_path", job.FilePath),
		)
		return false
	}
}

func (p *WorkerPool) Stop() {
	close(p.jobQueue)
	p.wg.Wait()
	p.logger.Info("Seluruh worker telah selesai dieksekusi")
}