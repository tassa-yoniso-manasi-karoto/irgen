package main

import (
	"os"
	"fmt"
	"context"
	"embed"
	"flag"
	"time"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/cmd"
)

//var version = "0.9.0-alpha"
var version = "0.8.1-prerelease"

//go:embed all:frontend/dist
var assets embed.FS

type ProcessParams struct {
	URL		string   `json:"url"`
	NumberOfTitle  int	`json:"numberOfTitle"`
	MaxXResolution int	`json:"maxXResolution"`
	MaxYResolution int	`json:"maxYResolution"`
}

type LogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) GetVersion() string {
	return version
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Logger = log.Output(zerolog.ConsoleWriter{
		NoColor: true,
		Out: &LogWriter{ctx: ctx},
		FormatTimestamp: func(i interface{}) string {return ""},
	}).With().Logger()

}

// LogWriter implements io.Writer interface for zerolog
type LogWriter struct {
	ctx context.Context
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

// Write implements io.Writer
func (w *LogWriter) Write(p []byte) (n int, err error) {
    rawMsg := string(p)
    level, cleanMsg := detectLogLevel(rawMsg)
    
    msg := LogMessage{
        Level:   level,
        Message: cleanMsg,
        Time:    time.Now().Format(time.TimeOnly),
    }
    
    runtime.EventsEmit(w.ctx, "log", msg)
    return len(p), nil
}

func (a *App) Process(params ProcessParams) string {
	cmd.Execute(params.URL)
	return ""
}


func main() {
	if len(os.Args) > 1 {
		inFile := flag.String("i", "", "file path or URL of an HTML article\n")
		wantVersion := flag.Bool("version", false, "print program version and exit")
		flag.Parse()
		if *wantVersion {
			fmt.Println(version)
			return
		}
		cmd.Execute(*inFile)
		return
	}
	app := NewApp()
	err := wails.Run(&options.App{
		Title:	 "IRGen",
		Width:  750,
		Height: 635,
		MinWidth:  750,
		MinHeight: 300,
		MaxWidth:  750,
		MaxHeight: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		fmt.Println(err)
	}
}
