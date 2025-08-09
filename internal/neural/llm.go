package neural

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/llms/openai"
	"go.uber.org/zap"
)

type LLM struct {
	llm  llms.Model
	opts []llms.CallOption
	fs   afero.Fs
	path string
	rpm  int
	reqs []time.Time
	log  *zap.SugaredLogger
}

func NewLLM(cfg AiConfig) (*LLM, bool) {
	llm := &LLM{
		rpm:  cfg.Rpm,
		reqs: make([]time.Time, 0),
		log:  zap.L().Sugar().Named("llm"),
	}

	var err error
	switch cfg.Provider {
	case "openai":
		opts := make([]openai.Option, 0)
		if cfg.BaseUrl != "" {
			opts = append(opts, openai.WithBaseURL(cfg.BaseUrl))
		}
		if cfg.ApiKey != "" {
			opts = append(opts, openai.WithToken(cfg.ApiKey))
		}
		if cfg.Model != "" {
			opts = append(opts, openai.WithModel(cfg.Model))
		}
		llm.llm, err = openai.New(opts...)
	case "mistral":
		opts := make([]mistral.Option, 0)
		if cfg.BaseUrl != "" {
			opts = append(opts, mistral.WithEndpoint(cfg.BaseUrl))
		}
		if cfg.ApiKey != "" {
			opts = append(opts, mistral.WithAPIKey(cfg.ApiKey))
		}
		if cfg.Model != "" {
			opts = append(opts, mistral.WithModel(cfg.Model))
		}
		llm.llm, err = mistral.New(opts...)
	case "googleai":
		opts := make([]googleai.Option, 0)
		opts = append(opts, googleai.WithHarmThreshold(googleai.HarmBlockNone))
		if cfg.ApiKey != "" {
			opts = append(opts, googleai.WithAPIKey(cfg.ApiKey))
		}
		if cfg.Model != "" {
			opts = append(opts, googleai.WithDefaultModel(cfg.Model))
		}
		llm.llm, err = googleai.New(context.Background(), opts...)
	case "anthropic":
		opts := make([]anthropic.Option, 0)
		if cfg.ApiKey != "" {
			opts = append(opts, anthropic.WithToken(cfg.ApiKey))
		}
		if cfg.Model != "" {
			opts = append(opts, anthropic.WithModel(cfg.Model))
		}
		llm.llm, err = anthropic.New(opts...)
	default:
		err = fmt.Errorf("unknown AI provider: %s", cfg.Provider)
	}
	if err != nil {
		llm.log.Error(err)
		return nil, false
	}

	llm.opts = append(llm.opts, llms.WithRepetitionPenalty(cfg.RepPen))
	llm.opts = append(llm.opts, llms.WithTemperature(cfg.Temp))
	llm.opts = append(llm.opts, llms.WithTopK(cfg.TopK))
	llm.opts = append(llm.opts, llms.WithMaxTokens(cfg.MaxTok))
	llm.opts = append(llm.opts, llms.WithStopWords(cfg.Stop))

	return llm, true
}

func (llm *LLM) EnableResponseLogging(fs afero.Fs, path string) {
	llm.fs = fs
	llm.path = filepath.Join(path, "ai_resp.txt")
}

func (llm *LLM) GenerateText(messages []llms.MessageContent) (string, bool) {
	resp, ok := llm.generate(messages, 0)
	if !ok {
		return "", false
	}

	llm.logResponse(resp)

	if len(resp) > 7 && resp[:7] == "<think>" {
		i := strings.Index(resp, "</think>")
		if i > 0 {
			resp = resp[i+8:]
			resp = strings.TrimSpace(resp)
		}
	}

	return resp, ok
}

func (llm *LLM) generate(messages []llms.MessageContent, nTry int) (string, bool) {
	if nTry > 5 {
		return "", false
	}

	if llm.rpm > 0 {
		for len(llm.reqs) > 0 && time.Since(llm.reqs[0]) > time.Minute {
			llm.reqs = llm.reqs[1:]
		}
		for len(llm.reqs) >= llm.rpm {
			sec := time.Until(llm.reqs[0].Add(time.Minute)).Seconds()
			sec = math.Ceil(sec)
			llm.log.Infow("sleeping", "sec", sec)
			time.Sleep(time.Duration(sec) * time.Second)
			llm.reqs = llm.reqs[1:]
		}
	}

	resp, err := llm.llm.GenerateContent(context.Background(), messages, llm.opts...)
	if llm.rpm > 0 {
		llm.reqs = append(llm.reqs, time.Now())
	}
	if err == nil {
		if len(resp.Choices) == 0 {
			return llm.generate(messages, nTry+1)
		}
		text := resp.Choices[0].Content
		if text == "" {
			return llm.generate(messages, nTry+1)
		}

		return text, true
	}

	if strings.Contains(err.Error(), "Service Unavailable") || strings.Contains(err.Error(), "Error 50") {
		sec := (nTry + 1) * 3
		llm.log.Infow("sleeping", "sec", sec)
		time.Sleep(time.Duration(sec) * time.Second)

		return llm.generate(messages, nTry+1)
	}

	if strings.Contains(err.Error(), "try again later") {
		sec := (nTry + 1) * 3
		if llm.rpm > 0 && len(llm.reqs) > 0 {
			secf := time.Since(llm.reqs[0]).Seconds()
			secf = math.Ceil(secf)
			sec = int(secf)
		}

		llm.log.Infow("sleeping", "sec", sec)
		time.Sleep(time.Duration(sec) * time.Second)

		return llm.generate(messages, nTry+1)
	}

	idx := strings.Index(err.Error(), "Please try again in")
	if idx <= 0 {
		llm.log.Warnw(err.Error())
		return "", false
	}

	str := err.Error()[idx:]
	str = regexp.MustCompile(`\d+`).FindString(str)
	sec, err := strconv.Atoi(str)
	if err != nil {
		llm.log.Warnw(err.Error())
		return "", false
	}

	sec++
	dur := time.Duration(sec) * time.Second
	llm.log.Infow("sleeping", "sec", sec)
	time.Sleep(dur)

	return llm.generate(messages, nTry+1)
}

func (llm *LLM) logResponse(resp string) {
	if resp == "" || llm.fs == nil {
		return
	}

	const delim = `

	========================================
	@--------------------------------------@
	========================================

`

	f, err := llm.fs.OpenFile(llm.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		llm.log.Warnw(err.Error())
		return
	}
	n, err := f.Write([]byte(resp))
	if err == nil && n < len(resp) {
		err = io.ErrShortWrite
	}
	n, err = f.Write([]byte(delim))
	if err == nil && n < len(delim) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		llm.log.Warnw(err.Error())
	}
}
