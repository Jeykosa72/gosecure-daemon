package process_test

import (
	"context"
	"testing"
	"time"

	"gosecure-daemon/internal/process"
)

func TestSecureRunner_TimeoutHandling(t *testing.T) {
	// Batas waktu sangat singkat (100ms)
	runner := process.NewSecureRunner(100 * time.Millisecond)

	// Eksekusi perintah 'sleep 1' yang dipastikan melebihi timeout
	_, err := runner.AnalyzeFile(context.Background(), "sleep", "1")
	if err == nil {
		t.Fatal("Ekspektasi error timeout, namun perintah berhasil dijalankan")
	}
}