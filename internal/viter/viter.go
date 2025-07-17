package viter

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/llms/openai"
	"go.uber.org/zap"
)

type Viter struct {
	llm  llms.Model
	opts []llms.CallOption
	book *Book
	log  *zap.SugaredLogger
	cfg  Config
}

func NewViter(cfg Config) (*Viter, bool) {
	v := &Viter{
		log: zap.L().Sugar().Named("viter"),
		cfg: cfg,
	}

	var err error
	switch cfg.Ai.Provider {
	case "openai":
		opts := make([]openai.Option, 0)
		if cfg.Ai.BaseUrl != "" {
			opts = append(opts, openai.WithBaseURL(cfg.Ai.BaseUrl))
		}
		if cfg.Ai.ApiKey != "" {
			opts = append(opts, openai.WithToken(cfg.Ai.ApiKey))
		}
		if cfg.Ai.Model != "" {
			opts = append(opts, openai.WithModel(cfg.Ai.Model))
		}
		v.llm, err = openai.New(opts...)
	case "mistral":
		opts := make([]mistral.Option, 0)
		if cfg.Ai.BaseUrl != "" {
			opts = append(opts, mistral.WithEndpoint(cfg.Ai.BaseUrl))
		}
		if cfg.Ai.ApiKey != "" {
			opts = append(opts, mistral.WithAPIKey(cfg.Ai.ApiKey))
		}
		if cfg.Ai.Model != "" {
			opts = append(opts, mistral.WithModel(cfg.Ai.Model))
		}
		v.llm, err = mistral.New(opts...)
	case "googleai":
		opts := make([]googleai.Option, 0)
		opts = append(opts, googleai.WithHarmThreshold(googleai.HarmBlockNone))
		if cfg.Ai.ApiKey != "" {
			opts = append(opts, googleai.WithAPIKey(cfg.Ai.ApiKey))
		}
		if cfg.Ai.Model != "" {
			opts = append(opts, googleai.WithDefaultModel(cfg.Ai.Model))
		}
		v.llm, err = googleai.New(context.Background(), opts...)
	default:
		err = fmt.Errorf("unknown AI provider: %s", cfg.Ai.Provider)
	}
	if err != nil {
		v.log.Error(err)
		return nil, false
	}

	v.opts = append(v.opts, llms.WithRepetitionPenalty(cfg.Ai.RepPen))
	v.opts = append(v.opts, llms.WithTemperature(cfg.Ai.Temp))
	v.opts = append(v.opts, llms.WithTopK(cfg.Ai.TopK))
	v.opts = append(v.opts, llms.WithMaxTokens(cfg.Ai.MaxTok))
	v.opts = append(v.opts, llms.WithStopWords(cfg.Ai.Stop))

	return v, true
}

func (v *Viter) CreateBook(fs afero.Fs, path string) bool {
	book, err := CreateBook(fs, path)
	if err != nil {
		v.log.Error(err)
		return false
	}

	v.book = book

	v.log.Infow("created book", "path", path)

	return true
}
