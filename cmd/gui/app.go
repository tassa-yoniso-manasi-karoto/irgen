package gui

import (
	"context"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/core"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/common"
	
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx  context.Context
	m    *meta.Meta
}

func NewApp(m *meta.Meta) *App {
	return &App{m: m}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.m.Log = setupLogger(ctx)
}

func (a *App) GetVersion() string {
	return common.Version
}

type ProcessParams struct {
	URL            string `json:"url"`
	NumberOfTitle  int    `json:"numberOfTitle"`
	MaxXResolution int    `json:"maxXResolution"`
	MaxYResolution int    `json:"maxYResolution"`
}

func (a *App) Process(params ProcessParams) string {
	a.m.Log.Debug().
		Str("Targ", params.URL).
		Int("MaxTitles", params.NumberOfTitle).
		Int("MaxXResolution", params.MaxXResolution).
		Int("MaxYResolution", params.MaxYResolution).
		Msg("Parameters provided by GUI")
	a.m.Targ = params.URL
	a.m.Config.MaxTitles = params.NumberOfTitle
	a.m.Config.ResXMax = params.MaxXResolution
	a.m.Config.ResYMax = params.MaxYResolution
	core.Execute(a.ctx, a.m)
	return ""
}


func (a *App) QueryAnkiConnect4MediaDir(q common.AnkiConnectQuery) bool {
	return common.QueryAnkiConnect(a.m, q)
}

func (a *App) OpenFileDialog() (string, error) {
    file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
        Title: "Select HTML File",
        Filters: []runtime.FileFilter{
            {
                DisplayName: "HTML Files (*.html;*.htm)",
                Pattern:     "*.html;*.htm",
            },
            {
                DisplayName: "All Files (*.*)",
                Pattern:     "*.*",
            },
        },
    })
    
    if err != nil {
        a.m.Log.Error().Err(err).Msg("Failed to open file dialog")
        return "", err
    }
    
    return file, nil
}
