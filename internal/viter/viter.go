package viter

import (
	"strings"
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
	CritiqueBookPrompt() string
	CritiqueBook2Prompt() string
	CritiqueBookNPrompt(from, to int) string
	CorrectTextPrompt() string
}

const maxTries = 3

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
		meta, ok := v.writeMeta(v.book.GetMeta(), 1)
		if !ok {
			return false
		}
		v.book.SetMeta(meta)
		i++
	}

	if v.book.GetMetaCrit() == nil || v.book.GetMetaCrit().GetImprovements() == "" {
		crit, ok := v.critiqueMeta(v.book.GetMeta(), 1)
		if !ok {
			return false
		}
		v.book.SetMetaCrit(crit)
	}

	for ; i < iterCount; i++ {
		meta, ok := v.updateMeta(v.book.GetMeta(), v.book.GetMetaCrit(), 1)
		if !ok {
			continue
		}
		meta = v.book.GetMeta().MergedCopy(meta)
		crit, ok := v.critiqueMeta(meta, 1)
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
		plan, ok := v.writePlan(v.book.GetPlan(), chapterCount, 1)
		if !ok {
			return false
		}
		v.book.SetPlan(plan)
		i++
	}

	if v.book.GetPlanCrit() == nil || v.book.GetPlanCrit().GetImprovements() == "" {
		crit, ok := v.critiquePlan(v.book.GetPlan(), 1)
		if !ok {
			return false
		}
		v.book.SetPlanCrit(crit)
	}

	for ; i < iterCount; i++ {
		plan, ok := v.updatePlan(v.book.GetPlan(), v.book.GetPlanCrit(), 1)
		if !ok {
			continue
		}
		plan = v.book.GetPlan().MergedCopy(plan)
		crit, ok := v.critiquePlan(plan, 1)
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
		newChp, ok := v.writeChapter(nChapter, 1)
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
		crit, ok := v.critiqueChapter(chp, 1)
		if !ok {
			return false
		}
		v.book.SetChapterCritique(nChapter, crit)
	}

	for ; i < iterCount; i++ {
		chp, _ := v.book.GetChapter(nChapter)
		crit, _ := v.book.GetChapterCritique(nChapter)
		chapter, ok := v.updateChapter(chp, crit, 1)
		if !ok {
			continue
		}
		crit, ok = v.critiqueChapter(chapter, 1)
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

func (v *Viter) UpdateBook(iterCount int) bool {
	if v.book == nil || v.book.GetPlan() == nil {
		return false
	}

	if len(v.book.GetBookImprovements()) == 0 {
		imps, ok := v.critiqueBook(v.book.GetChapters())
		if !ok {
			return false
		}
		v.book.SetBookImprovements(imps)
	}

	for range iterCount {
		imps := v.book.GetBookImprovements()
		chps, ok := v.updateBook(imps)
		if !ok {
			continue
		}
		imps, ok = v.critiqueBook(chps)
		if !ok {
			continue
		}
		for _, chp := range chps {
			v.book.SetChapter(chp.GetNumber(), chp)
			v.book.RemoveChapterCritique(chp.GetNumber())
		}
		v.book.SetBookImprovements(imps)
	}

	return true
}

func (v *Viter) CorrectChapter(nChapter int, maxLen int) bool {
	if v.book == nil || v.book.GetPlan() == nil {
		return false
	}

	chp, _ := v.book.GetChapter(nChapter)
	if chp.GetContent() == "" {
		v.log.Info("chapter is empty", "n", nChapter)
		return true
	}

	if !v.correctChapterText(chp, maxLen) {
		return false
	}

	v.book.SetChapter(nChapter, chp)

	return true
}

func (v *Viter) CorrectAllChapters(maxLen int) bool {
	if v.book == nil || v.book.GetPlan() == nil {
		return false
	}

	for i := 1; i <= v.book.GetPlan().Count(); i++ {
		if !v.CorrectChapter(i, maxLen) {
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

func (v *Viter) writeMeta(prevMeta *books.BookMeta, nTry int) (*books.BookMeta, bool) {
	if nTry > maxTries {
		return nil, false
	}

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
		return v.writeMeta(prevMeta, nTry+1)
	}

	v.log.Infow("wrote meta")

	return meta, true
}

func (v *Viter) critiqueMeta(meta *books.BookMeta, nTry int) (*books.Critique, bool) {
	if nTry > maxTries {
		return nil, false
	}

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
		return v.critiqueMeta(meta, nTry+1)
	}
	if crit.GetImprovements() == "" {
		v.log.Warnw("no improvements found")
		return v.critiqueMeta(meta, nTry+1)
	}

	v.log.Infow("critiqued meta", "score", crit.GetScore())

	return crit, true
}

func (v *Viter) updateMeta(prevMeta *books.BookMeta, crit *books.Critique, nTry int) (*books.BookMeta, bool) {
	if nTry > maxTries {
		return nil, false
	}

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
		return v.updateMeta(prevMeta, crit, nTry+1)
	}

	v.log.Infow("updated meta")

	return meta, true
}

func (v *Viter) writePlan(prevPlan books.Plan, chapterCount, nTry int) (books.Plan, bool) {
	if nTry > maxTries {
		return nil, false
	}

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
		return v.writePlan(prevPlan, chapterCount, nTry+1)
	}

	v.log.Infow("wrote plan", "chapters", len(plan))

	return plan, true
}

func (v *Viter) critiquePlan(plan books.Plan, nTry int) (*books.Critique, bool) {
	if nTry > maxTries {
		return nil, false
	}

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
		return v.critiquePlan(plan, nTry+1)
	}
	if crit.GetImprovements() == "" {
		v.log.Warnw("no improvements found")
		return v.critiquePlan(plan, nTry+1)
	}

	v.log.Infow("critiqued plan", "score", crit.GetScore())

	return crit, true
}

func (v *Viter) updatePlan(prevPlan books.Plan, crit *books.Critique, nTry int) (books.Plan, bool) {
	if nTry > maxTries {
		return nil, false
	}

	v.log.Infow("updating plan")

	hst := neural.NewHistory()
	hst.AddText(v.pp.WritePlanPrompt(prevPlan.Count()))
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
		return v.updatePlan(prevPlan, crit, nTry+1)
	}

	v.log.Infow("updated plan", "chapters", len(plan))

	return plan, true
}

func (v *Viter) writeChapter(n, nTry int) (*books.Chapter, bool) {
	if nTry > maxTries {
		return nil, false
	}

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
		return v.writeChapter(n, nTry+1)
	}
	chapter.SetNumber(n)

	v.log.Infow("wrote chapter", "n", n, "title", chapter.GetTitle())

	return chapter, true
}

func (v *Viter) critiqueChapter(chapter *books.Chapter, nTry int) (*books.Critique, bool) {
	if nTry > maxTries {
		return nil, false
	}

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
		return v.critiqueChapter(chapter, nTry+1)
	}
	if crit.GetImprovements() == "" {
		v.log.Warnw("no improvements found")
		return v.critiqueChapter(chapter, nTry+1)
	}

	v.log.Infow("critiqued chapter", "n", chapter.GetNumber(), "score", crit.GetScore())

	return crit, true
}

func (v *Viter) updateChapter(prevChp *books.Chapter, crit *books.Critique, nTry int) (*books.Chapter, bool) {
	if nTry > maxTries {
		return nil, false
	}

	v.log.Infow("updating chapter", "n", prevChp.GetNumber())

	prevDiff := books.DiffFromText(prevChp.GetContent())

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
	hst.AddText(prevChp.GetTitle() + "\n" + prevDiff.String())
	hst.AddText(v.pp.UpdateNChapterPrompt(prevChp.GetNumber()))
	hst.AddText(crit.GetImprovements())

	diffData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return nil, false
	}

	diff, err := books.DiffFromString(diffData)
	if err != nil {
		v.log.Warnw(err.Error())
		return v.updateChapter(prevChp, crit, nTry+1)
	}
	if diff.IsEmpty() {
		v.log.Warn("no changes found")
		return v.updateChapter(prevChp, crit, nTry+1)
	}

	prevDiff.Merge(diff)

	chp := books.NewChapter(prevChp.GetNumber(), prevChp.GetTitle(), prevDiff.Text())

	v.log.Infow("updated chapter", "n", chp.GetNumber())

	return chp, true
}

func (v *Viter) critiqueBook(chps []*books.Chapter) (books.Plan, bool) {
	v.log.Info("critiquing book")

	hst := neural.NewHistory()
	hst.AddText(v.pp.CritiqueBookPrompt())
	hst.AddMessage()
	for _, chp := range chps {
		hst.AddText(chp.String())
	}
	hst.AddText(v.pp.CritiqueBook2Prompt())

	review, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return books.Plan{}, false
	}

	hst.AddMessage()
	hst.AddText(review)

	res := books.Plan{}

	for i := 1; i <= v.book.GetPlan().Count(); i += 5 {
		from := i
		to := min(from+4, v.book.GetPlan().Count())
		v.log.Infow("critiquing chapters", "from", from, "to", to)

		hst.AddMessage()
		hst.AddText(v.pp.CritiqueBookNPrompt(from, to))

		var imps books.Plan
		var err error
		for range maxTries {
			impData, ok := v.tg.GenerateText(hst.Messages())
			if !ok {
				continue
			}

			hst.AddMessage()
			hst.AddText(impData)

			imps, err = books.PlanFromString(impData)
			if err == nil {
				break
			}
			v.log.Warnw(err.Error())
		}
		if err != nil {
			continue
		}

		res = res.MergedCopy(imps)
	}

	v.log.Infow("critiqued book")

	return res, true
}

func (v *Viter) updateBook(imps books.Plan) ([]*books.Chapter, bool) {
	v.log.Infow("updating book", "chapters", imps.Count())

	chps := make([]*books.Chapter, 0, imps.Count())

	for i := 1; i <= imps.Count(); i++ {
		imp := imps[i-1]
		crit := &books.Critique{}
		crit.SetImprovements(imp.GetContent())

		prevChp, err := v.book.GetChapter(imp.GetNumber())
		if err != nil {
			v.log.Warnw(err.Error())
			continue
		}

		chp, ok := v.updateChapter(prevChp, crit, 1)
		if !ok {
			continue
		}

		chps = append(chps, chp)
	}

	return chps, true
}

func (v *Viter) correctChapterText(chp *books.Chapter, maxLen int) bool {
	v.log.Infow("correcting chapter text", "n", chp.GetNumber(), "len", len(chp.GetContent()), "max", maxLen)

	if len(chp.GetContent()) <= maxLen {
		res, ok := v.correctText(chp.GetContent())
		if !ok {
			return false
		}
		chp.SetContent(res)
		v.log.Infow("corrected chapter text", "n", chp.GetNumber())
		return true
	}

	correctedContent := ""
	{
		result := ""
		part := ""
		ps := strings.SplitSeq(chp.GetContent(), "\n")
		for p := range ps {
			if p == "" {
				continue
			}

			if len(part) == 0 {
				part = p
			} else if len(part)+len(p)+1 > maxLen {
				ok := false
				part, ok = v.correctText(part)
				if !ok {
					return false
				}
				result += part + "\n\n"
				part = p
			} else {
				part += "\n\n" + p
			}
		}

		if len(part) > 0 {
			ok := false
			part, ok = v.correctText(part)
			if !ok {
				return false
			}
			result += part
		}

		correctedContent = result
	}

	{
		result := ""
		part := ""
		ps := strings.Split(correctedContent, "\n")
		for i := len(ps) - 1; i >= 0; i-- {
			p := ps[i]
			if p == "" {
				continue
			}

			if len(part) == 0 {
				part = p
			} else if len(part)+len(p)+1 > maxLen {
				ok := false
				part, ok = v.correctText(part)
				if !ok {
					return false
				}
				result = part + "\n\n" + result
				part = p
			} else {
				part = p + "\n\n" + part
			}
		}

		if len(part) > 0 {
			ok := false
			part, ok = v.correctText(part)
			if !ok {
				return false
			}
			result = part + "\n\n" + result
		}

		correctedContent = result
	}

	correctedContent = strings.TrimSpace(correctedContent)
	chp.SetContent(correctedContent)

	v.log.Infow("corrected chapter text", "n", chp.GetNumber(), "len", len(correctedContent))

	return true
}

func (v *Viter) correctText(text string) (string, bool) {
	v.log.Infow("correcting text", "len", len(text))

	hst := neural.NewHistory()
	hst.AddText(v.pp.CorrectTextPrompt())
	hst.AddMessage()
	hst.AddText(text)

	correctedData, ok := v.tg.GenerateText(hst.Messages())
	if !ok {
		return "", false
	}

	v.log.Infow("corrected text", "len", len(correctedData))

	return correctedData, true
}
