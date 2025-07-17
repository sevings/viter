package viter

import (
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Ai AiConfig
}

type AiConfig struct {
	Provider string
	BaseUrl  string `koanf:"base_url"`
	ApiKey   string `koanf:"api_key"`
	Model    string `koanf:"model"`
	Temp     float64
	TopK     int     `koanf:"tok_k"`
	RepPen   float64 `koanf:"rep_pen"`
	MaxTok   int     `koanf:"max_tok"`
	Stop     []string
}

func LoadConfig(path string) (Config, error) {
	var kConf = koanf.New("/")

	var cfg Config

	err := kConf.Load(file.Provider(path), toml.Parser())
	if err != nil {
		return cfg, err
	}

	err = kConf.Unmarshal("", &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
