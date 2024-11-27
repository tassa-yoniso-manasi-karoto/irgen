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
	"bytes"
	"encoding/json"
	
	"github.com/schollz/progressbar/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/k0kubun/pp"
	"github.com/gookit/color"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
)

const (
	maxRetries = 2
	retryDelay = 1 * time.Second
	jsonMIME = "application/json"
)



type AnkiConnectQuery struct {
	Action string `json:"action"`
}

type AnkiConnectRespGetMediaDirPath struct {
	Result,	Error string
}

var ankiConnectQueryTemplate = `{
    "action": "%s",
    "version": 6
}`

func QueryAnkiConnect(m *meta.Meta, q AnkiConnectQuery) (ok bool) {
	jsonStr := []byte(fmt.Sprintf(ankiConnectQueryTemplate, q.Action))
	resp, err := http.Post("http://localhost:8765", jsonMIME, bytes.NewBuffer(jsonStr))
	if err != nil {
		m.Log.Error().
			Err(err).
			Msg("POST request to AnkiConnect failed")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		m.Log.Error().
			Err(err).
			Msg("couldn't read response of AnkiConnect after getMediaDirPath")
		return
	}
	var response AnkiConnectRespGetMediaDirPath
	json.Unmarshal(body, &response)
	if response.Error != "" {
		pp.Println(response)
		m.Log.Error().
			//Str("RawAnkiConnectErr", response.Error).
			Err(errors.New(response.Error)).
			Msg("AnkiConnect returned an error when requesting getMediaDirPath")
		return
	}
	m.Config.CollectionMedia = response.Result
	m.Log.Debug().Str("MediaDir", response.Result).Msg("")
	m.Log.Info().Msg("AnkiConnect detected")
	return true
}


type DownloadProgress struct {
	Current		int	`json:"current"`
	Total		int	`json:"total"`
	Progress	float64	`json:"progress"`
	Speed		string	`json:"speed"`
	CurrentFile	string	`json:"currentFile"`
	Operation	string	`json:"operation"`
}

func calculateAverageSpeed(bytesDownloaded int64, startTime time.Time) string {
	elapsed := time.Since(startTime).Seconds()
	bytesPerSecond := float64(bytesDownloaded) / elapsed
	
	if bytesPerSecond >= 1024*1024 {
		return fmt.Sprintf("%.1f MB/s", bytesPerSecond/1024/1024)
	}
	return fmt.Sprintf("%.1f KB/s", bytesPerSecond/1024)
}


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
					Operation:		"Downloading",
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

func StringCapLen(s string, max int) string{
	trimmed := false
	for len(s) > max {
		s = s[:len(s)-1]
		trimmed = true
	}
	if trimmed {
		s += "…"
	}
	return s
}






func placeholder6zuwertzuikuztrewi9876() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}

