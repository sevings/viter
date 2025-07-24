package viter

import (
	"viter/internal/books"
	"viter/internal/neural"

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
	UpdateChapterPrompt() string
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
	book *books.Book
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
	book, err := books.CreateBook(fs, path)
	if err != nil {
		v.log.Error(err)
		return false
	}

	v.book = book

	v.log.Infow("created book", "path", path)

	return true
}

func (v *Viter) LoadBook(fs afero.Fs, path string) bool {
	book, err := books.LoadBook(fs, path)
	if err != nil {
		v.log.Error(err)
		return false
	}

	v.book = book

	v.log.Infow("loaded book", "path", path)

	return true
}

func (v *Viter) EnableArchiving() {
	v.book.EnableArchiving()
}

func (v *Viter) SetPlot(plot string) {
	v.book.GetMeta().SetPlot(plot)
}

func (v *Viter) GetBook() *books.Book {
	return v.book
}

func (v *Viter) UpdateMeta(iterCount int) bool {
	if v.book == nil {
		return false
	}

	i := 0
	if !v.book.GetMeta().IsFilled() {
		meta, ok := v.writeMeta(v.book.GetMeta())
		if !ok {
			return false
		}
		v.book.SetMeta(meta)
		i++
	}

	if v.book.GetMetaCrit() == nil || v.book.GetMetaCrit().GetImprovements() == "" {
		crit, ok := v.critiqueMeta(v.book.GetMeta())
		if !ok {
			return false
		}
		v.book.SetMetaCrit(crit)
	}

	for ; i < iterCount; i++ {
		meta, ok := v.updateMeta(v.book.GetMeta(), v.book.GetMetaCrit())
		if !ok {
			continue
		}
		meta = v.book.GetMeta().MergedCopy(meta)
		crit, ok := v.critiqueMeta(meta)
		if !ok {
			continue
		}
		v.book.SetMeta(meta)
		v.book.SetMetaCrit(crit)
	}

	return true
}

func (v *Viter) UpdatePlan(chapterCount, iterCount int) bool {
	if v.book == nil {
		return false
	}

	i := 0
	if len(v.book.GetPlan()) != chapterCount {
		plan, ok := v.writePlan(v.book.GetPlan(), chapterCount)
		if !ok {
			return false
		}
		v.book.SetPlan(plan)
		i++
	}

	if v.book.GetPlanCrit() == nil || v.book.GetPlanCrit().GetImprovements() == "" {
		crit, ok := v.critiquePlan(v.book.GetPlan())
		if !ok {
			return false
		}
		v.book.SetPlanCrit(crit)
	}

	for ; i < iterCount; i++ {
		plan, ok := v.updatePlan(v.book.GetPlan(), v.book.GetPlanCrit())
		if !ok {
			continue
		}
		plan = v.book.GetPlan().MergedCopy(plan)
		crit, ok := v.critiquePlan(plan)
		if !ok {
			continue
		}
		v.book.SetPlan(plan)
		v.book.SetPlanCrit(crit)
	}

	return true
}

func (v *Viter) UpdateChapter(nChapter, iterCount int) bool {
	if v.book == nil {
		return false
	}

	i := 0
	if chp, err := v.book.GetChapter(nChapter); err != nil {
		return false
	} else if chp == nil || chp.GetContent() == "" {
		newChp, ok := v.writeChapter(nChapter)
		if !ok {
			return false
		}
		v.book.SetChapter(nChapter, newChp)
		i++
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

	for ; i < iterCount; i++ {
		chp, _ := v.book.GetChapter(nChapter)
		crit, _ := v.book.GetChapterCritique(nChapter)
		chapter, ok := v.updateChapter(chp, crit)
		if !ok {
			continue
		}
		crit, ok = v.critiqueChapter(chapter)
		if !ok {
			continue
		}
		v.book.SetChapter(nChapter, chapter)
		v.book.SetChapterCritique(nChapter, crit)
	}

	return true
}

func (v *Viter) UpdateAllChapters(iterCount int) bool {
	if v.book == nil || v.book.GetPlan() == nil {
		return false
	}

	for i := 1; i <= v.book.GetPlan().Count(); i++ {
		if !v.UpdateChapter(i, iterCount) {
			return false
		}
	}

	return true
}

func (v *Viter) ExportMarkdown() bool {
	if v.book == nil {
		return false
	}

	err := v.book.ExportMarkdown()
	if err != nil {
		v.log.Errorw(err.Error())
		return false
	}

	v.log.Info("exported markdown")

	return true
}

func (v *Viter) ExportHTML() bool {
	if v.book == nil {
		return false
	}

	err := v.book.ExportHTML()
	if err != nil {
		v.log.Errorw(err.Error())
		return false
	}

	v.log.Info("exported html")

	return true
}

func (v *Viter) ExportEPUB() bool {
	if v.book == nil {
		return false
	}

	err := v.book.ExportEPUB()
	if err != nil {
		v.log.Errorw(err.Error())
		return false
	}

	v.log.Info("exported epub")

	return true
}

func (v *Viter) writeMeta(prevMeta *books.BookMeta) (*books.BookMeta, bool) {
	v.log.Infow("writing meta")

	hst := neural.NewHistory()
	hst.AddText(v.pp.WriteMetaPrompt())
	hst.AddMessage()
	hst.AddText(prevMeta.String())

	metaData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	meta, err := books.MetaFromString(metaData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("wrote meta")

	return meta, true
}

func (v *Viter) critiqueMeta(meta *books.BookMeta) (*books.Critique, bool) {
	v.log.Infow("critiquing meta")

	hst := neural.NewHistory()
	hst.AddText(v.pp.CritiqueMetaPrompt())
	hst.AddMessage()
	hst.AddText(meta.String())

	critData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	crit, err := books.CritiqueFromString(critData)
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

func (v *Viter) updateMeta(prevMeta *books.BookMeta, crit *books.Critique) (*books.BookMeta, bool) {
	v.log.Infow("updating meta")

	hst := neural.NewHistory()
	hst.AddText(v.pp.WriteMetaPrompt())
	hst.AddMessage()
	hst.AddText(prevMeta.String())
	hst.AddText(crit.GetImprovements())
	hst.AddText(v.pp.UpdateMetaPrompt())

	metaData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	meta, err := books.MetaFromString(metaData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("updated meta")

	return meta, true
}

func (v *Viter) writePlan(prevPlan books.Plan, chapterCount int) (books.Plan, bool) {
	v.log.Infow("writing plan", "chapters", chapterCount)

	hst := neural.NewHistory()
	hst.AddText(v.pp.WritePlanPrompt(chapterCount))
	hst.AddMessage()
	hst.AddText(v.book.GetMeta().String())
	if len(prevPlan) > 0 {
		hst.AddText(prevPlan.String())
	}

	planData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	plan, err := books.PlanFromString(planData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("wrote plan", "chapters", len(plan))

	return plan, true
}

func (v *Viter) critiquePlan(plan books.Plan) (*books.Critique, bool) {
	v.log.Infow("critiquing plan")

	hst := neural.NewHistory()
	hst.AddText(v.pp.CritiquePlanPrompt())
	hst.AddMessage()
	hst.AddText(v.book.GetMeta().String())
	hst.AddText(plan.String())

	critData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	crit, err := books.CritiqueFromString(critData)
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

func (v *Viter) updatePlan(prevPlan books.Plan, crit *books.Critique) (books.Plan, bool) {
	v.log.Infow("updating plan")

	hst := neural.NewHistory()
	hst.AddText(v.pp.WritePlanPrompt(0))
	hst.AddMessage()
	hst.AddText(v.book.GetMeta().String())
	hst.AddText(prevPlan.String())
	hst.AddText(crit.GetImprovements())
	hst.AddText(v.pp.UpdatePlanPrompt())

	planData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	plan, err := books.PlanFromString(planData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}

	v.log.Infow("updated plan", "chapters", len(plan))

	return plan, true
}

func (v *Viter) writeChapter(n int) (*books.Chapter, bool) {
	v.log.Infow("writing chapter")

	hst := neural.NewHistory()
	hst.AddText(v.pp.WriteChapterPrompt())
	hst.AddMessage()
	hst.AddText(v.book.GetMeta().String())
	hst.AddText(v.book.GetPlan().String())
	for i := 1; i < n; i++ {
		chp, err := v.book.GetChapter(i)
		if err != nil {
			v.log.Warnw(err.Error())
			continue
		}
		hst.AddText(chp.String())
	}
	hst.AddText(v.pp.WriteNChapterPrompt(n))

	planChp, err := v.book.GetPlanChapter(n)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}
	hst.AddText(planChp.String())

	chapterData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	chapter, err := books.ChapterFromString(chapterData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}
	chapter.SetNumber(n)

	v.log.Infow("wrote chapter", "n", n, "title", chapter.GetTitle())

	return chapter, true
}

func (v *Viter) critiqueChapter(chapter *books.Chapter) (*books.Critique, bool) {
	v.log.Infow("critiquing chapter")

	hst := neural.NewHistory()
	hst.AddText(v.pp.CritiqueChapterPrompt())
	hst.AddMessage()
	hst.AddText(v.book.GetMeta().String())
	hst.AddText(v.book.GetPlan().String())
	for i := 1; i < chapter.GetNumber(); i++ {
		chp, err := v.book.GetChapter(i)
		if err != nil {
			v.log.Warnw(err.Error())
			continue
		}
		hst.AddText(chp.String())
	}
	hst.AddText(chapter.String())
	hst.AddText(v.pp.CritiqueNChapterPrompt(chapter.GetNumber()))

	critData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	crit, err := books.CritiqueFromString(critData)
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

func (v *Viter) updateChapter(prevChp *books.Chapter, crit *books.Critique) (*books.Chapter, bool) {
	v.log.Infow("updating chapter", "n", prevChp.GetNumber())

	prevDiff := books.DiffFromText(prevChp.String())

	hst := neural.NewHistory()
	hst.AddText(v.pp.UpdateChapterPrompt())
	hst.AddMessage()
	hst.AddText(v.book.GetMeta().String())
	hst.AddText(v.book.GetPlan().String())
	for i := 1; i < prevChp.GetNumber(); i++ {
		chp, err := v.book.GetChapter(i)
		if err != nil {
			v.log.Warnw(err.Error())
			continue
		}
		hst.AddText(chp.String())
	}
	hst.AddText(prevDiff.String())
	hst.AddText(v.pp.UpdateNChapterPrompt(prevChp.GetNumber()))
	hst.AddText(crit.GetImprovements())

	diffData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	diff, err := books.DiffFromString(diffData)
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}
	prevDiff.Merge(diff)

	chp, err := books.ChapterFromString(prevDiff.Text())
	if err != nil {
		v.log.Warnw(err.Error())
		return nil, false
	}
	chp.SetNumber(prevChp.GetNumber())

	v.log.Infow("updated chapter", "n", chp.GetNumber())

	return chp, true
}
