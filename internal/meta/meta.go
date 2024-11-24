package meta

import (
	//"encoding/json"
	"fmt"
	"time"
	"os"
	"strings"
	
	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog"
)

type Config struct {
	MaxTitles int `json:"maxTitles"`
	ResXMax   int `json:"resXMax"`
	ResYMax   int `json:"resYMax"`
}

type Meta struct {
	Targ	string
	Log	zerolog.Logger
	Koanf  *koanf.Koanf
	Config Config
}

func New() *Meta {
	if strings.Contains(os.Args[0], "-dev-") {		
		zerolog.SetGlobalLevel(zerolog.TraceLevel)	
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	return &Meta{
		Log:   zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly}).With().Timestamp().Logger(),
		Koanf: koanf.New("."),
		Config: Config{
			MaxTitles: 3,
			ResXMax:   1920,
			ResYMax:   1080,
		},
	}
}

func (m *Meta) LoadConfig() error {
	if err := m.Koanf.Load(file.Provider("config.json"), json.Parser()); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error loading config: %v", err)
		}
		return nil
	}

	return m.Koanf.Unmarshal("", &m.Config)
}

type ConfigManager struct {
	k      *koanf.Koanf
	config *Config
	logger zerolog.Logger
}

func NewConfigManager() *ConfigManager {
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "config-manager").
		Logger()

	return &ConfigManager{
		k:      koanf.New("."),
		config: &Config{},
		logger: logger,
	}
}


