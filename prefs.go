package main

import (
	"os"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"github.com/rs/zerolog"
)

type prefType struct {
	DestDir, Collection string
	LenStack, ResXMax, ResYMax int
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
	if pref.DestDir == "" {
		pref.DestDir = CurrentDir
	}
	if pref.Collection == "" {
		logger.Fatal().Msg("Images can't be imported because the path to collection has not been provided. Aborting...")
	}
}