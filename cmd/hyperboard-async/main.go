package main

import (
	"github.com/dharmab/hyperboard/internal/async"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := async.NewCommand().Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start async controller")
	}
}
