package meta

import (
	//"encoding/json"
	"fmt"
	"time"
	"os"
	"strings"
	"strconv"
	
	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog"
)

type Config struct {
	CollectionMedia string `json:"collectionMedia"`
	DestDir string `json:"destDir"`
	Functions []string
	FunctionsScopes []int
	MaxTitles int `json:"maxTitles"`
	ResXMax   int `json:"resXMax"`
	ResYMax   int `json:"resYMax"`
}

type Meta struct {
	Targ	string
	Log	zerolog.Logger
	Koanf  *koanf.Koanf
	Config Config
	DevMode bool
}

var ConsoleWriter = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly}).With().Timestamp().Logger()

func New() *Meta {
	return &Meta{
		Log:   ConsoleWriter,
		Koanf: koanf.New("."),
		Config: Config{
			Functions: []string{"FromSuperior", "FromSuperior", "FromSuperiorAndDescendants", "FromSuperiorAndDescendants"},
			FunctionsScopes: []int{1, 2, 3, 10},
			MaxTitles: 3,
			ResXMax:   1920,
			ResYMax:   1080,
		},
	}
}

func (m *Meta) LoadConfig() error {
	if strings.Contains(os.Args[0], "-dev-") {		
		m.DevMode = true
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		m.Log.Trace().Msg("Running in development with trace logging")
	}
	if err := m.Koanf.Load(file.Provider("config.json"), json.Parser()); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error loading config: %v", err)
		}
		return nil
	}
	err := m.Koanf.Unmarshal("", &m.Config)
	// clear defaults if config.json exist
	m.Config.Functions = []string{}
	m.Config.FunctionsScopes = []int{}
	
	for _, field := range strings.Fields(m.Koanf.String("functions")) {
		parts := strings.Split(field, "=")
		if len(parts) != 2 {
			continue
		}
		scope, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		m.Config.Functions = append(m.Config.Functions, parts[0])
		m.Config.FunctionsScopes = append(m.Config.FunctionsScopes, scope)
	}
	m.LogConfig("config state at LoadConfig()")
	// TODO rm ↓?
	/*if m.Config.DestDir == "" {
		exe, err := os.Executable()
		if err != nil {
			logger.Fatal().Error(err).Msg("failed to identify directory from which the binary is run")
		}
		CurrentDir = path.Dir(exe)
	}*/
	return err
}



func (m *Meta) LogConfig(s string)  {
	msg := "config state:"
	if s != "" {
		msg = s
	}
	m.Log.Debug().
		Strs("Functions", m.Config.Functions).
		Ints("FunctionsScopes", m.Config.FunctionsScopes).
		Int("MaxTitles", m.Config.MaxTitles).
		Int("ResXMax", m.Config.ResXMax).
		Int("ResYMax", m.Config.ResYMax).
		Msg(msg)
}


// TODO rm ↓?
// whether path of dir is given with a final "/" should be irrelevant
func safe(path string) string {
	return strings.TrimSuffix(path, string(os.PathSeparator)) + string(os.PathSeparator)
}


