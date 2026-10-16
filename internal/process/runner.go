package process

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// SecureRunner mengeksekusi perintah analisis keamanan dengan batas waktu ketat.
type SecureRunner struct {
	timeout time.Duration
}

func NewSecureRunner(timeout time.Duration) *SecureRunner {
	return &SecureRunner{timeout: timeout}
}

// AnalyzeFile mengeksekusi perintah eksternal terhadap file target secara terisolasi.
func (r *SecureRunner) AnalyzeFile(parentCtx context.Context, command string, filePath string) (string, error) {
	// Membuat context child dengan timeout eksplisit
	ctx, cancel := context.WithTimeout(parentCtx, r.timeout)
	defer cancel() // Melepas resource timer segera setelah fungsi selesai

	// Mematikan subprocess secara otomatis jika context timeout/canceled
	cmd := exec.CommandContext(ctx, command, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("proses terhenti: timeout %v terlampaui untuk file %s", r.timeout, filePath)
		}
		return "", fmt.Errorf("eksekusi gagal [%s %s]: %w (output: %s)", command, filePath, err, string(output))
	}

	return string(output), nil
}