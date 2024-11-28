package cli

import (
	"os"
	"context"
	
	urcli "github.com/urfave/cli/v2"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/cmd/gui"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/core"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/common"
)

func Execute() {
	m := meta.New()
	if err := m.LoadConfig(); err != nil {
		m.Log.Error().Err(err).Msg("config load failed")
	}

	cli := &urcli.App{
		Name:	"irgen",
		Version: common.Version,
		Flags: []urcli.Flag{
			&urcli.StringFlag{
				Name:	"input",
				Aliases: []string{"i"},
				Usage:   "file path or URL of an HTML article",
			},
			&urcli.IntFlag{
				Name:  "max-titles",
				Value: m.Config.MaxTitles,
			},
			&urcli.IntFlag{
				Name:  "res-x-max",
				Value: m.Config.ResXMax,
			},
			&urcli.IntFlag{
				Name:  "res-y-max",
				Value: m.Config.ResYMax,
			},
		},
		Action: func(c *urcli.Context) error {
			run(c, m)
			return nil
		},
	}

	cli.Run(os.Args)
}

func run(c *urcli.Context, m *meta.Meta) {
	m.Log.Trace().Strs("os.Args", os.Args).Msg("")
	m.Log.Debug().
		Bool("startAsGUI?", c.NArg() == 0 && !c.IsSet("input")).
		Int("c.NArg()", c.NArg()).
		Bool("inputFlagNotSet", !c.IsSet("input")).
		Msg("")
	if c.NArg() == 0 && !c.IsSet("input") {
		m.GUIMode = true
		gui.Run(m)
		return
	}

	common.QueryAnkiConnect(m, common.AnkiConnectQuery{"getMediaDirPath"})
	m.Config.MaxTitles = c.Int("max-titles")
	m.Config.ResXMax = c.Int("res-x-max")
	m.Config.ResYMax = c.Int("res-y-max")

	m.Targ = c.String("input")
	if m.Targ == "" && c.Args().First() != "" {
		m.Targ = c.Args().First()
	}
	
	if success := core.Execute(context.TODO(), m); !success {
		os.Exit(1)
	}
}



func placeholder3456() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
