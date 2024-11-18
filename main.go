package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.eu.org/envloader"

	_ "github.com/lemon-mint/coord/provider/aistudio"
	_ "github.com/lemon-mint/coord/provider/anthropic"
	_ "github.com/lemon-mint/coord/provider/openai"
	_ "github.com/lemon-mint/coord/provider/vertexai"
)

//go:generate sqlc generate

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	envloader.LoadEnvFile(".env")
	config, err := LoadConfig("config.jsonnet")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	_ = config
}
