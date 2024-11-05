package main

import (
	"context"

	"github.com/lemon-mint/coord"
	"github.com/lemon-mint/coord/llm"
	"github.com/lemon-mint/coord/pconf"
	"github.com/lemon-mint/coord/provider"

	_ "github.com/lemon-mint/coord/provider/aistudio"
	_ "github.com/lemon-mint/coord/provider/anthropic"
	_ "github.com/lemon-mint/coord/provider/openai"
	_ "github.com/lemon-mint/coord/provider/vertexai"

	"encoding/json"
	"os"
	"strings"

	"github.com/google/go-jsonnet"

	"gopkg.eu.org/envloader"
)

func LoadConfig(file string) (*Config, error) {
	envloader.LoadEnvFile(".env")
	vm := jsonnet.MakeVM()

	var envMap = make(map[string]string, len(os.Environ()))
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if ok {
			envMap[k] = v
		}
	}

	envData, err := json.Marshal(envMap)
	if err != nil {
		return nil, err
	}
	vm.ExtVar("env_data", string(envData))

	jsondata, err := vm.EvaluateFile(file)
	if err != nil {
		return nil, err
	}

	var c Config
	err = json.Unmarshal([]byte(jsondata), &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

type ModelConfigs struct {
	ChunkGenerator ModelConfig `json:"chunk_generator"`
}

type Config struct {
	ModelConfigs ModelConfigs `json:"model_configs"`
	Providers    []Providers  `json:"providers"`
}

type Parameters struct {
	Temperature float32 `json:"temperature"`
	TopP        float32 `json:"top_p"`
	TopK        int     `json:"top_k"`
	MaxTokens   int     `json:"max_tokens"`
}

type ModelConfig struct {
	Model      string     `json:"model"`
	Parameters Parameters `json:"parameters"`
	Provider   string     `json:"provider"`
}

type Providers struct {
	APIKey    string `json:"api_key"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Baseurl   string `json:"baseurl,omitempty"`
	Location  string `json:"location,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

func Connect(m Providers) (c provider.LLMClient, err error) {
	switch m.Type {
	case "aistudio":
		c, err = coord.NewLLMClient(context.Background(), "aistudio", pconf.WithAPIKey(m.APIKey))
	case "anthropic":
		c, err = coord.NewLLMClient(context.Background(), "anthropic", pconf.WithAPIKey(m.APIKey))
	case "openai":
		if m.Baseurl != "" {
			c, err = coord.NewLLMClient(context.Background(), "openai", pconf.WithAPIKey(m.APIKey), pconf.WithBaseURL(m.Baseurl))
			break
		}
		c, err = coord.NewLLMClient(context.Background(), "openai", pconf.WithAPIKey(m.APIKey))
	case "vertexai":
		c, err = coord.NewLLMClient(context.Background(), "vertexai", pconf.WithLocation(m.Location), pconf.WithProjectID(m.ProjectID))
	}
	return
}

func GetModel(c provider.LLMClient, name string, params Parameters) (m llm.Model, err error) {
	config := new(llm.Config)
	config.Temperature = &params.Temperature
	config.SafetyFilterThreshold = llm.BlockDefault
	if params.TopP != 0 {
		config.TopP = &params.TopP
	}
	if params.TopK != 0 {
		config.TopK = &params.TopK
	}
	if params.MaxTokens != 0 {
		config.MaxOutputTokens = &params.MaxTokens
	}

	return c.NewLLM(name, config)
}
