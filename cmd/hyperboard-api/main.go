package main

import (
	"github.com/dharmab/hyperboard/internal/api"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := api.NewCommand().Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start API server")
	}
}
