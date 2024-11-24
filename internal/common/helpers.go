package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"context"
	"time"
	"sync/atomic"
	"errors"

	"github.com/schollz/progressbar/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
)

const (
	maxRetries = 1
	retryDelay = 1 * time.Second
)

type DownloadProgress struct {
	Current	 int	 `json:"current"`	  // Current number of files downloaded
	Total	   int	 `json:"total"`		// Total number of files to download
	Progress	float64 `json:"progress"`	 // Percentage complete
	Speed	   string  `json:"speed"`		// Current download speed
	CurrentFile string  `json:"currentFile"`  // Name of file being downloaded
}

// Function to calculate average speed
func calculateAverageSpeed(bytesDownloaded int64, startTime time.Time) string {
	elapsed := time.Since(startTime).Seconds()
	bytesPerSecond := float64(bytesDownloaded) / elapsed
	
	if bytesPerSecond >= 1024*1024 {
		return fmt.Sprintf("%.1f MB/s", bytesPerSecond/1024/1024)
	}
	return fmt.Sprintf("%.1f KB/s", bytesPerSecond/1024)
}

// retryableDownload attempts to download a file with retries
func retryableDownload(ctx context.Context, filepath, URL string, totalBytes *int64, m *meta.Meta) (int64, error) {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		written, err := downloadSingleFile(ctx, filepath, URL, totalBytes)
		if err == nil {
			return written, nil
		}
		
		lastErr = err
		m.Log.Warn().
			Err(err).
			Int("attempt", attempt).
			Str("url", URL).
			Msg("Download failed, retrying...")
		
		// Check if context is cancelled before retrying
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(retryDelay * time.Duration(attempt)):
			// Exponential backoff
		}
	}
	
	return 0, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func DownloadFiles(ctx context.Context, m *meta.Meta, URLs, filenames []string) error {
	if len(URLs) != len(filenames) {
		return errors.New("URLs and filenames slices must have the same length")
	}

	total := len(URLs)
	current := 0
	failed := 0
	startTime := time.Now()
	var totalBytesDownloaded int64 = 0

	var bar *progressbar.ProgressBar
	if !m.GUIMode {
		bar = progressbar.Default(int64(total))
	}

	m.Log.Trace().Int("total", len(URLs)).Msg("URLs of img to download")
	
	for i, URL := range URLs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fullPath := m.Config.CollectionMedia + filenames[i]
			written, err := retryableDownload(ctx, fullPath, URL, &totalBytesDownloaded, m)
			
			// Always increment current and update progress, even on failure
			current++
			progress := float64(current) / float64(total) * 100
			
			if err != nil {
				failed++
				m.Log.Error().
					Err(err).
					Str("url", URL).
					Str("filepath", fullPath).
					Msg("Failed to download file after all retries")
			}

			if !m.GUIMode {
				bar.Add(1)
			} else {
				m.Log.Trace().
					Str("filename", filenames[i]).
					Int("idx", i).
					Int64("bytes", written).
					Bool("success", err == nil).
					Msg("Download attempt completed")

				runtime.EventsEmit(ctx, "download-progress", DownloadProgress{
					Current:	 current,
					Total:	   total,
					Progress:	progress,
					Speed:	   calculateAverageSpeed(totalBytesDownloaded, startTime),
					CurrentFile: filenames[i], // Show just filename instead of full path
				})
			}
		}
	}

	if failed > 0 {
		return fmt.Errorf("completed %d/%d downloads with %d failures", current, total, failed)
	}
	return nil
}

func downloadSingleFile(ctx context.Context, filepath string, URL string, totalBytes *int64) (int64, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		// Try to remove partially downloaded file
		os.Remove(filepath)
		return 0, fmt.Errorf("failed to save file: %w", err)
	}
	
	atomic.AddInt64(totalBytes, written)
	return written, nil
}