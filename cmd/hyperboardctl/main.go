package main

import (
	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/dharmab/hyperboard/internal/cli/notes"
	"github.com/dharmab/hyperboard/internal/cli/posts"
	"github.com/dharmab/hyperboard/internal/cli/replace"
	"github.com/dharmab/hyperboard/internal/cli/tagcategories"
	"github.com/dharmab/hyperboard/internal/cli/tags"
	"github.com/rs/zerolog/log"
)

func main() {
	app := cli.NewApp()
	posts.Register(app)
	notes.Register(app)
	tags.Register(app)
	tagcategories.Register(app)
	replace.Register(app)
	if err := app.RootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run command")
	}
}
