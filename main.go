package main

import (
	"gopkg.eu.org/envloader"

	_ "github.com/lemon-mint/coord/provider/aistudio"
	_ "github.com/lemon-mint/coord/provider/anthropic"
	_ "github.com/lemon-mint/coord/provider/openai"
	_ "github.com/lemon-mint/coord/provider/vertexai"
)

func main() {
	envloader.LoadEnvFile(".env")
	coord.
}
