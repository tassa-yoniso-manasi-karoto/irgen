package gui

import (
	"embed"

	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
	
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func Run(m *meta.Meta) {
	app := NewApp(m)
	wails.Run(&options.App{
		Title:	 "IRGen",
		Width:	 750,
		Height:	660,
		MinWidth:  750,
		MinHeight: 300,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 255}, // #1e1e1e

		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			BackdropType: windows.None,
			Theme:	windows.Dark,
			CustomTheme: &windows.ThemeSettings{
				DarkModeTitleBar:   windows.RGB(30, 30, 30),  // Match background
				DarkModeTitleText:  windows.RGB(212, 212, 212),
				DarkModeBorder:	 windows.RGB(30, 30, 30),
			},
		},
		Mac: &mac.Options{
			Appearance:  mac.NSAppearanceNameDarkAqua,
		},
		//Debug: options.Debug{ OpenInspectorOnStartup: false, },
	})
}