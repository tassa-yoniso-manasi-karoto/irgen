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

var userWasWarned bool

type AnkiConnectRequest struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

type AnkiConnectResponse struct {
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
}

func SendAnkiConnectRequest(m *meta.Meta, action string, params interface{}) (json.RawMessage, error) {
	request := AnkiConnectRequest{
		Action:  action,
		Version: 6,
		Params:  params,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post("http://localhost:8765", jsonMIME, bytes.NewBuffer(jsonData))
	if err != nil {
		m.Log.Error().
			Err(err).
			Msg("POST request to AnkiConnect failed")
		return nil, err
	}
	defer resp.Body.Close()

	var response AnkiConnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		m.Log.Error().
			Err(err).
			Msg("couldn't decode AnkiConnect response")
		return nil, err
	}

	if response.Error != nil {
		str := fmt.Sprint("AnkiConnect: ", response.Error)
		m.Log.Error().Msg(str)
		return nil, errors.New(str)
	}

	return response.Result, nil
}

func QueryAnkiConnect(m *meta.Meta, q AnkiConnectRequest) (result string, err error) {
	response, err := SendAnkiConnectRequest(m, q.Action, nil)
	if err != nil {
		return "", err
	}

	var resultStr string
	if err := json.Unmarshal(response, &resultStr); err != nil {
		return "", err
	}
	m.Log.Debug().
		Str("Action", q.Action).
		RawJSON("response", response).
		Msg("AnkiConnect request successful")
	return resultStr, nil
}


func QueryAnkiConnectMediaDir(m *meta.Meta) bool {
	q := AnkiConnectRequest{Action: "getMediaDirPath", Version: 6}
	result, err := QueryAnkiConnect(m, q)
	if err == nil {
		m.Config.CollectionMedia = result
	} else if !userWasWarned {
		userWasWarned = true
		m.Log.Error().
			Err(err).
			Msg("Failed to connect to AnkiConnect." +
			" Please make sure Anki is running and AnkiConnect is properly installed.")
	}
	return err == nil
}



func VerifyNoteTypeFields(m *meta.Meta, modelName string, expectedFields []string) error {
	params := map[string]interface{}{
		"modelName": modelName,
	}

	response, err := SendAnkiConnectRequest(m, "modelFieldNames", params)
	if err != nil {
		return fmt.Errorf("failed to verify note type fields: %w", err)
	}

	var fields []string
	if err := json.Unmarshal(response, &fields); err != nil {
		return fmt.Errorf("failed to parse fields response: %w", err)
	}

	if !compareFields(fields, expectedFields) {
		return fmt.Errorf("note type fields mismatch. Expected: %v, Got: %v", expectedFields, fields)
	}

	return nil
}


func CreateDeck(m *meta.Meta, deckName string) error {
	params := map[string]interface{}{
		"deck": deckName,
	}

	_, err := SendAnkiConnectRequest(m, "createDeck", params)
	return err
}


func AddNote(m *meta.Meta, deckName string, modelName string, fields map[string]string, tags []string) error {
	params := map[string]interface{}{
		"note": map[string]interface{}{
			"deckName":  deckName,
			"modelName": modelName,
			"fields":    fields,
			"tags":      tags,
		},
	}

	_, err := SendAnkiConnectRequest(m, "addNote", params)
	return err
}


func compareFields(actual, expected []string) bool {
	for i, field := range actual {
		if len(expected)-1 >= i && field != expected[i] {
			return false
		}
	}
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
	if total == 0 {
		return nil
	}
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
					CurrentFile: filenames[i],
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

