package viter

import (
	"github.com/spf13/afero"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/zap"
)

type PromptProvider interface {
	WriteMetaPrompt() string
	CritiqueMetaPrompt() string
	UpdateMetaPrompt() string
}

type TextGenerator interface {
	GenerateText(messages []llms.MessageContent) (string, bool)
}

type Viter struct {
	tg   TextGenerator
	pp   PromptProvider
	book *Book
	log  *zap.SugaredLogger
	cfg  Config
}

func NewViter(cfg Config, pp PromptProvider, tg TextGenerator) (*Viter, bool) {
	v := &Viter{
		tg:  tg,
		pp:  pp,
		log: zap.L().Sugar().Named("viter"),
		cfg: cfg,
	}

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

func (v *Viter) LoadBook(fs afero.Fs, path string) bool {
	book, err := LoadBook(fs, path)
	if err != nil {
		v.log.Error(err)
		return false
	}

	v.book = book

	v.log.Infow("loaded book", "path", path)

	return true
}

func (v *Viter) UpdateMeta(minScore int) bool {
	if v.book == nil {
		return false
	}

	if !v.book.GetMeta().IsFilled() {
		meta, ok := v.writeMeta(v.book.GetMeta())
		if !ok {
			return false
		}
		v.book.SetMeta(meta)
	}

	if v.book.GetMetaCrit() == nil || v.book.GetMetaCrit().GetImprovements() == "" {
		crit, ok := v.critiqueMeta(v.book.GetMeta())
		if !ok {
			return false
		}
		v.book.SetMetaCrit(crit)
	}

	for v.book.GetMetaCrit().GetScore() < minScore {
		meta, ok := v.updateMeta(v.book.GetMeta(), v.book.GetMetaCrit())
		if !ok {
			return false
		}
		crit, ok := v.critiqueMeta(meta)
		if !ok {
			return false
		}
		if crit.GetScore() <= v.book.GetMetaCrit().GetScore() {
			return true
		}
		v.book.SetMeta(meta)
		v.book.SetMetaCrit(crit)
	}

	return true
}

func (v *Viter) writeMeta(prevMeta *BookMeta) (*BookMeta, bool) {
	v.log.Infow("writing meta")

	messages := make([]llms.MessageContent, 2)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.WriteMetaPrompt()),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(prevMeta.String()),
		},
	}

	metaData, ok := v.tg.GenerateText(messages)
	if !ok {
		return nil, false
	}

	meta, err := MetaFromString(metaData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("wrote meta")

	return meta, true
}

func (v *Viter) critiqueMeta(meta *BookMeta) (*Critique, bool) {
	v.log.Infow("critiquing meta")

	messages := make([]llms.MessageContent, 2)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.CritiqueMetaPrompt()),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(meta.String()),
		},
	}

	critData, ok := v.tg.GenerateText(messages)
	if !ok {
		return nil, false
	}

	crit, err := CritiqueFromString(critData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("critiqued meta", "score", crit.GetScore())

	return crit, true
}

func (v *Viter) updateMeta(prevMeta *BookMeta, crit *Critique) (*BookMeta, bool) {
	v.log.Infow("updating meta")

	messages := make([]llms.MessageContent, 2)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.WriteMetaPrompt()),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(prevMeta.String()),
			llms.TextPart(crit.GetImprovements()),
			llms.TextPart(v.pp.UpdateMetaPrompt()),
		},
	}

	metaData, ok := v.tg.GenerateText(messages)
	if !ok {
		return nil, false
	}

	meta, err := MetaFromString(metaData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("updated meta")

	return meta, true
}
