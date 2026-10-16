package config

import (
	"time"
)

// Config menyimpan seluruh parameter operasional daemon.
type Config struct {
	WatchDir    string        // Direktori sensitif yang dipantau
	WorkerCount int           // Jumlah worker goroutine sejajar
	QueueSize   int           // Kapasitas buffered channel antrean
	ExecTimeout time.Duration // Batas waktu maksimal eksekusi proses analisis
	SecurityCmd string        // Binary keamanan yang dieksekusi (contoh: sha256sum)
}

// LoadDefaultConfig mengembalikan konfigurasi default tingkat produksi.
func LoadDefaultConfig() *Config {
	return &Config{
		WatchDir:    "/tmp/gosecure-target",
		WorkerCount: 4,
		QueueSize:   100,
		ExecTimeout: 3 * time.Second,
		SecurityCmd: "sha256sum",
	}
}