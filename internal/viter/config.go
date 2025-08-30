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
	WriteMeta       string `koanf:"write_meta"`
	CritiqueMeta    string `koanf:"critique_meta"`
	UpdateMeta      string `koanf:"update_meta"`
	WritePlan       string `koanf:"write_plan"`
	CritiquePlan    string `koanf:"critique_plan"`
	UpdatePlan      string `koanf:"update_plan"`
	WriteChapter    string `koanf:"write_chapter"`
	CritiqueChapter string `koanf:"critique_chapter"`
	UpdateChapter   string `koanf:"update_chapter"`
	CritiqueBook    string `koanf:"critique_book"`
	CorrectText     string `koanf:"correct_text"`

	// Legacy fields for backward compatibility
	Write    string `koanf:"write"`
	Update   string `koanf:"update"`
	Critique string `koanf:"critique"`
	Correct  string `koanf:"correct"`
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
