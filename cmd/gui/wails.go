package gui

import (
	"embed"

	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
	
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func Run(m *meta.Meta) {
	app := NewApp(m)
	wails.Run(&options.App{
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