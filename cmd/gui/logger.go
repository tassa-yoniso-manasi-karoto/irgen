package gui

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type LogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time	string `json:"time"`
}

type LogWriter struct {
	ctx context.Context
}

func setupLogger(ctx context.Context) zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:			&LogWriter{ctx: ctx},
		NoColor:		true,
		FormatTimestamp: 	func(i interface{}) string { return "" },
	}).With().Logger()
}


func detectLogLevel(msg string) (level, cleanMsg string) {
	levels := map[string]string{
		"DBG": "DEBUG",
		"INF": "INFO",
		"WRN": "WARN",
		"ERR": "ERROR",
		"FTL": "FATAL",
		"PNC": "PANIC",
		"TRC": "TRACE",
	}
	
	// Check if message starts with a level prefix
	if len(msg) >= 3 {
		prefix := msg[:3]
		if level, exists := levels[prefix]; exists {
			// Remove the level prefix and any leading spaces
			cleanMsg := msg[3:]
			if len(cleanMsg) > 0 && cleanMsg[0] == ' ' {
				cleanMsg = cleanMsg[1:]
			}
			return level, cleanMsg
		}
	}
	
	return "---", msg
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	rawMsg := string(p)
	level, cleanMsg := detectLogLevel(rawMsg)
	
	msg := LogMessage{
		Level:   level,
		Message: cleanMsg,
		Time:	time.Now().Format(time.TimeOnly),
	}
	
	runtime.EventsEmit(w.ctx, "log", msg)
	return len(p), nil
}