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
	WritePlanPrompt(chapterCount int) string
	CritiquePlanPrompt() string
	UpdatePlanPrompt() string
	WriteChapterPrompt() string
	CritiqueChapterPrompt() string
	WriteNChapterPrompt(chapter int) string
	CritiqueNChapterPrompt(chapter int) string
	UpdateNChapterPrompt(chapter int) string
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
		meta = v.book.GetMeta().MergedCopy(meta)
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

func (v *Viter) UpdatePlan(chapterCount, minScore int) bool {
	if v.book == nil {
		return false
	}

	if len(v.book.GetPlan()) != chapterCount {
		plan, ok := v.writePlan(v.book.GetPlan(), chapterCount)
		if !ok {
			return false
		}
		v.book.SetPlan(plan)
	}

	if v.book.GetPlanCrit() == nil || v.book.GetPlanCrit().GetImprovements() == "" {
		crit, ok := v.critiquePlan(v.book.GetPlan())
		if !ok {
			return false
		}
		v.book.SetPlanCrit(crit)
	}

	for v.book.GetPlanCrit().GetScore() < minScore {
		plan, ok := v.updatePlan(v.book.GetPlan(), v.book.GetPlanCrit())
		if !ok {
			return false
		}
		plan = v.book.GetPlan().MergedCopy(plan)
		crit, ok := v.critiquePlan(plan)
		if !ok {
			return false
		}
		if crit.GetScore() <= v.book.GetPlanCrit().GetScore() {
			return true
		}
		v.book.SetPlan(plan)
		v.book.SetPlanCrit(crit)
	}

	return true
}

func (v *Viter) UpdateChapter(nChapter, minScore int) bool {
	if v.book == nil {
		return false
	}

	if chp, err := v.book.GetChapter(nChapter); err != nil {
		return false
	} else if chp == nil || chp.GetContent() == "" {
		newChp, ok := v.writeChapter(nChapter)
		if !ok {
			return false
		}
		v.book.SetChapter(nChapter, newChp)
	}

	if crit, err := v.book.GetChapterCritique(nChapter); err != nil {
		return false
	} else if crit == nil || crit.GetImprovements() == "" {
		chp, _ := v.book.GetChapter(nChapter)
		crit, ok := v.critiqueChapter(chp)
		if !ok {
			return false
		}
		v.book.SetChapterCritique(nChapter, crit)
	}

	for {
		prevCrit, _ := v.book.GetChapterCritique(nChapter)
		if prevCrit.GetScore() >= minScore {
			return true
		}
		prevChp, _ := v.book.GetChapter(nChapter)
		chapter, ok := v.updateChapter(prevChp, prevCrit)
		if !ok {
			return false
		}
		crit, ok := v.critiqueChapter(chapter)
		if !ok {
			return false
		}
		if crit.GetScore() <= prevCrit.GetScore() {
			return true
		}
		v.book.SetChapter(nChapter, chapter)
		v.book.SetChapterCritique(nChapter, crit)
	}
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
	if crit.GetImprovements() == "" {
		v.log.Warnw("no improvements found")
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

func (v *Viter) writePlan(prevPlan Plan, chapterCount int) (Plan, bool) {
	v.log.Infow("writing plan", "chapters", chapterCount)

	messages := make([]llms.MessageContent, 2)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.WritePlanPrompt(chapterCount)),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(v.book.GetMeta().String()),
		},
	}
	if len(prevPlan) > 0 {
		messages[1].Parts = append(messages[1].Parts, llms.TextPart(prevPlan.String()))
	}

	planData, ok := v.tg.GenerateText(messages)
	if !ok {
		return nil, false
	}

	plan, err := PlanFromString(planData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("wrote plan", "chapters", len(plan))

	return plan, true
}

func (v *Viter) critiquePlan(plan Plan) (*Critique, bool) {
	v.log.Infow("critiquing plan")

	messages := make([]llms.MessageContent, 2)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.CritiquePlanPrompt()),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(v.book.GetMeta().String()),
			llms.TextPart(plan.String()),
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
	if crit.GetImprovements() == "" {
		v.log.Warnw("no improvements found")
		return nil, false
	}

	v.log.Infow("critiqued plan", "score", crit.GetScore())

	return crit, true
}

func (v *Viter) updatePlan(prevPlan Plan, crit *Critique) (Plan, bool) {
	v.log.Infow("updating plan")

	messages := make([]llms.MessageContent, 2)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.WritePlanPrompt(0)),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(v.book.GetMeta().String()),
			llms.TextPart(prevPlan.String()),
			llms.TextPart(crit.GetImprovements()),
			llms.TextPart(v.pp.UpdatePlanPrompt()),
		},
	}

	planData, ok := v.tg.GenerateText(messages)
	if !ok {
		return nil, false
	}

	plan, err := PlanFromString(planData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("updated plan", "chapters", len(plan))

	return plan, true
}

func (v *Viter) writeChapter(n int) (*Chapter, bool) {
	v.log.Infow("writing chapter")

	messages := make([]llms.MessageContent, 2)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.WriteChapterPrompt()),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(v.book.GetMeta().String()),
			llms.TextPart(v.book.GetPlan().String()),
		},
	}
	for i := 1; i < n; i++ {
		chp, err := v.book.GetChapter(i)
		if err != nil {
			v.log.Warnw(err.Error())
			continue
		}
		messages[1].Parts = append(messages[1].Parts, llms.TextPart(chp.String()))
	}
	messages[1].Parts = append(messages[1].Parts, llms.TextPart(v.pp.WriteNChapterPrompt(n)))
	planChp, err := v.book.GetPlanChapter(n)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}
	messages[1].Parts = append(messages[1].Parts, llms.TextPart(planChp.String()))

	chapterData, ok := v.tg.GenerateText(messages)
	if !ok {
		return nil, false
	}

	chapter, err := ChapterFromString(chapterData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}
	chapter.SetNumber(n)

	v.log.Infow("wrote chapter", "n", n, "title", chapter.GetTitle())

	return chapter, true
}

func (v *Viter) critiqueChapter(chapter *Chapter) (*Critique, bool) {
	v.log.Infow("critiquing chapter")

	messages := make([]llms.MessageContent, 2)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.CritiqueChapterPrompt()),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(v.book.GetMeta().String()),
			llms.TextPart(v.book.GetPlan().String()),
		},
	}
	for i := 1; i < chapter.GetNumber(); i++ {
		chp, err := v.book.GetChapter(i)
		if err != nil {
			v.log.Warnw(err.Error())
			continue
		}
		messages[1].Parts = append(messages[1].Parts, llms.TextPart(chp.String()))
	}
	messages[1].Parts = append(messages[1].Parts, llms.TextPart(chapter.String()))
	messages[1].Parts = append(messages[1].Parts, llms.TextPart(v.pp.CritiqueNChapterPrompt(chapter.GetNumber())))

	critData, ok := v.tg.GenerateText(messages)
	if !ok {
		return nil, false
	}

	crit, err := CritiqueFromString(critData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}
	if crit.GetImprovements() == "" {
		v.log.Warnw("no improvements found")
		return nil, false
	}

	v.log.Infow("critiqued chapter", "n", chapter.GetNumber(), "score", crit.GetScore())

	return crit, true
}

func (v *Viter) updateChapter(prevChp *Chapter, crit *Critique) (*Chapter, bool) {
	v.log.Infow("updating chapter", "n", prevChp.GetNumber())

	messages := make([]llms.MessageContent, 4)
	messages[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.WriteChapterPrompt()),
		},
	}
	messages[1] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(v.book.GetMeta().String()),
			llms.TextPart(v.book.GetPlan().String()),
		},
	}
	for i := 1; i < prevChp.GetNumber(); i++ {
		chp, err := v.book.GetChapter(i)
		if err != nil {
			v.log.Warnw(err.Error())
			continue
		}
		messages[1].Parts = append(messages[1].Parts, llms.TextPart(chp.String()))
	}
	messages[1].Parts = append(messages[1].Parts, llms.TextPart(v.pp.WriteNChapterPrompt(prevChp.GetNumber())))
	messages[2] = llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextPart(prevChp.String()),
		},
	}
	messages[3] = llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(v.pp.UpdateNChapterPrompt(prevChp.GetNumber())),
			llms.TextPart(crit.GetImprovements()),
		},
	}

	chapterData, ok := v.tg.GenerateText(messages)
	if !ok {
		return nil, false
	}

	chp, err := ChapterFromString(chapterData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}
	chp.SetNumber(prevChp.GetNumber())

	v.log.Infow("updated chapter", "n", chp.GetNumber())

	return chp, true
}
