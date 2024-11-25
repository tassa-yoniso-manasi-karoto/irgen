package gui

import (
	"context"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/core"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/common"
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
	core.Execute(a.m)
	return ""
}