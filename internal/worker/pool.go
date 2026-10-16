package worker

import (
	"context"
	"log"
	"sync"

	"gosecure-daemon/internal/process"
)

// Job merepresentasikan payload event berkas yang perlu dianalisis.
type Job struct {
	FilePath  string
	EventType string
}

// WorkerPool mengelola konkurensi worker menggunakan Go channels dan WaitGroup.
type WorkerPool struct {
	workerCount int
	jobQueue    chan Job
	runner      *process.SecureRunner
	cmd         string
	wg          sync.WaitGroup
}

func NewWorkerPool(workerCount int, queueSize int, runner *process.SecureRunner, cmd string) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, queueSize), // Buffered channel untuk mencegah blocking pada producer
		runner:      runner,
		cmd:         cmd,
	}
}

// Start menginisialisasi worker goroutines sesuai kapasitas yang ditentukan.
func (p *WorkerPool) Start(ctx context.Context) {
	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
	log.Printf("[WorkerPool] Berhasil menjalankan %d worker goroutines", p.workerCount)
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	log.Printf("[Worker-%d] Siap menerima tugas...", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker-%d] Menerima sinyal pembatalan, mematikan worker...", id)
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				log.Printf("[Worker-%d] Antrean pekerjaan telah ditutup", id)
				return
			}
			log.Printf("[Worker-%d] Memproses file [%s] | Event: %s", id, job.FilePath, job.EventType)
			result, err := p.runner.AnalyzeFile(ctx, p.cmd, job.FilePath)
			if err != nil {
				log.Printf("[Worker-%d] ERROR pada %s: %v", id, job.FilePath, err)
			} else {
				log.Printf("[Worker-%d] HASIL ANALISIS [%s]: %s", id, job.FilePath, result)
			}
		}
	}
}

// Submit memasukkan pekerjaan ke antrean channel secara thread-safe.
func (p *WorkerPool) Submit(ctx context.Context, job Job) bool {
	select {
	case <-ctx.Done():
		return false
	case p.jobQueue <- job:
		return true
	default:
		log.Printf("[WorkerPool] WARNING: Antrean penuh (%d items), mengabaikan event: %s", len(p.jobQueue), job.FilePath)
		return false
	}
}

// Stop menghentikan alur worker secara bersih (Graceful Shutdown).
func (p *WorkerPool) Stop() {
	close(p.jobQueue) // Penutupan channel menandakan tidak ada job baru yang masuk
	p.wg.Wait()       // Menunggu hingga seluruh worker menyelesaikan job yang tersisa
	log.Println("[WorkerPool] Seluruh worker telah berhenti dengan bersih.")
}