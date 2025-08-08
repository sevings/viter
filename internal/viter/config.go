package viter

import (
	"viter/internal/neural"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Models Models
	Ai     []neural.AiConfig
}

type Models struct {
	Write    string
	Update   string
	Critique string
	Correct  string
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
