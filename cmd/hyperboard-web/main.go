package main

import (
	"github.com/dharmab/hyperboard/internal/web"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := web.NewCommand().Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start web server")
	}
}
