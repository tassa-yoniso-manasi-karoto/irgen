package cmd

import (
	"os"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"path/filepath"
	
	"github.com/rs/zerolog"
)

type prefType struct {
	DestDir, CollectionMedia string
	MaxTitles, ResXMax, ResYMax int
	KeepListEntire,CommentsToCutpattern   bool
	// MustMoveTxtAddendumToSrc bool
	Fn []string
	FnScope []int
	InFile,Title,Caption string
	ObjectsRegister map[string][]string
}

var pref prefType

func init() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	e, err := os.Executable()
	if err != nil {
		logger.Fatal().Msg(fmt.Sprint(err))
	} else {
		CurrentDir = path.Dir(e)
	}
	data, err := os.ReadFile(CurrentDir + "/config.json")
	if err != nil {
		logger.Info().Msg(fmt.Sprint(err))
	}
	if errjs := json.Unmarshal([]byte(data), &pref); errjs != nil && !errors.Is(err, os.ErrNotExist){
		check(err)
	}
	if pref.DestDir == "" && filepath.IsAbs(CurrentDir) {
		pref.DestDir = CurrentDir
	}
	pref.DestDir = safe(pref.DestDir)
	pref.CollectionMedia = safe(pref.CollectionMedia)
}

// whether path of dir is given with a final "/" should be irrelevant
func safe(path string) string {
	return strings.TrimSuffix(path, string(os.PathSeparator)) + string(os.PathSeparator)
}
