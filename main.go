package main

import (
"context"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/cmd"
)

var version = "0.8.2-prerelease"

//go:embed all:frontend/dist
var assets embed.FS

// TODO mv to /internal
type Config struct {
	MaxTitles int `json:"maxTitles"`
	ResXMax   int `json:"resXMax"`
	ResYMax   int `json:"resYMax"`
}

var defaultConfig = Config{
	MaxTitles: 3,
	ResXMax:   1920,
	ResYMax:   1080,
}

var appConfig = defaultConfig

type ProcessParams struct {
	URL			string `json:"url"`
	NumberOfTitle  int	`json:"numberOfTitle"`
	MaxXResolution int	`json:"maxXResolution"`
	MaxYResolution int	`json:"maxYResolution"`
}

type LogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time	string `json:"time"`
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
		NoColor:		 true,
		Out:			&LogWriter{ctx: ctx},
		// suppress timestamp
		FormatTimestamp: func(i interface{}) string { return "" },
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


func (a *App) Process(params ProcessParams) string {
	appConfig.MaxTitles = params.NumberOfTitle
	appConfig.ResXMax = params.MaxXResolution
	appConfig.ResYMax = params.MaxYResolution

	cmd.Execute(params.URL)
	return ""
}

func loadConfig() error {
	k := koanf.New(".")
	
	if err := k.Load(file.Provider("config.json"), json.Parser()); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error loading config: %v", err)
		}
	} else {
		if err := k.Unmarshal("", &appConfig); err != nil {
			return fmt.Errorf("error unmarshaling config: %v", err)
		}
	}
	return nil
}

func runGUI() error {
	app := NewApp()
	return wails.Run(&options.App{
		Title:	 "IRGen",
		Width:	 750,
		Height:	635,
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
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if err := loadConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to load configuration")
	}

	app := &cli.App{
		Name:	"irgen",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:	"input",
				Aliases: []string{"i"},
				Usage:   "file path or URL of an HTML article",
			},
			&cli.IntFlag{
				Name:	"max-titles",
				Value:   appConfig.MaxTitles,
				Usage:   "maximum number of titles",
			},
			&cli.IntFlag{
				Name:	"res-x-max",
				Value:   appConfig.ResXMax,
				Usage:   "maximum X resolution",
			},
			&cli.IntFlag{
				Name:	"res-y-max",
				Value:   appConfig.ResYMax,
				Usage:   "maximum Y resolution",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 && !c.IsSet("input") {
				return runGUI()
			}

			// Update configuration from CLI flags
			appConfig.MaxTitles = c.Int("max-titles")
			appConfig.ResXMax = c.Int("res-x-max")
			appConfig.ResYMax = c.Int("res-y-max")

			input := c.String("input")
			if input == "" && c.Args().First() != "" {
				input = c.Args().First()
			}
			
			cmd.Execute(input)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("Application error")
	}
}