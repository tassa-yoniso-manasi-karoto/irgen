package main

import (
	"os"
	"fmt"
	"context"
	"embed"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/cmd"
)

//go:embed all:frontend/dist
var assets embed.FS

type ProcessParams struct {
	URL		string   `json:"url"`
	NumberOfTitle  int	`json:"numberOfTitle"`
	MaxXResolution int	`json:"maxXResolution"`
	MaxYResolution int	`json:"maxYResolution"`
}


// LogMessage represents a structured log message
type LogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Create custom writer for zerolog
	/*logWriter := zerolog.ConsoleWriter{
		Out:        &LogWriter{app: a},
		TimeFormat: "15:04:05",
	}*/

	// Configure zerolog
	//log := zerolog.New(logWriter).With().Timestamp().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{NoColor: true, Out:&LogWriter{ctx: ctx}}).With().Timestamp().Logger()
	//log.Output(&LogWriter{ctx: ctx})

}

// LogWriter implements io.Writer interface for zerolog
type LogWriter struct {
	ctx context.Context
}

// Write implements io.Writer
func (w *LogWriter) Write(p []byte) (n int, err error) {
	msg := LogMessage{
		Message: string(p),
		Time:    "", // zerolog's ConsoleWriter already includes time
	}
	// Emit the log message as an event
	runtime.EventsEmit(w.ctx, "log", msg)
	return len(p), nil
}


///#############################

func (a *App) ProcessURL(params ProcessParams) string {
	//time.Sleep(3 * time.Second)
	cmd.GUI(params.URL)
	return "SOMETHING INFORMING ABOUT THE RESULTS"
	// Add your URL processing logic here with the additional parameters
	/*return "Processed URL: " + params.URL + 
		   "\nNumber of Title: " + string(params.NumberOfTitle) +
		   "\nMax X Resolution: " + string(params.MaxXResolution) +
		   "\nMax Y Resolution: " + string(params.MaxYResolution)*/
}


func main() {
	if len(os.Args) > 1 {
		cmd.CLI()
		return
	}
	app := NewApp()
	err := wails.Run(&options.App{
		Title:	 "IRGen",
		Width:  750,
		Height: 720,
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
