package books

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-shiori/go-epub"
	"github.com/spf13/afero"
)

var (
	ErrInvalidChapterIndex = errors.New("invalid chapter index")
	ErrNoMetaFile          = errors.New("no meta file found")
	ErrBookNotFinished     = errors.New("book not finished")
)

type Book struct {
	fs   afero.Fs
	path string
	meta *BookMeta
	plan Plan
	chps []*Chapter

	metaCrit *Critique
	planCrit *Critique
	chpsCrit []*Critique
}

// CreateBook creates a new book with the given filesystem and path.
// it creates a file 'meta.md' and initializes the book's metadata.
func CreateBook(fs afero.Fs, path string) (*Book, error) {
	b := &Book{fs: fs, path: path, meta: &BookMeta{}}
	if err := b.Save(); err != nil {
		return nil, err
	}
	return b, nil
}

// LoadBook loads an existing book from the given filesystem and path.
// it reads the 'meta.md' file and the 'plan.md' file if it exists.
func LoadBook(fs afero.Fs, path string) (*Book, error) {
	b := &Book{fs: fs, path: path}

	// Load metadata
	metaPath := filepath.Join(path, "meta.md")
	if exists, err := afero.Exists(fs, metaPath); err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrNoMetaFile
	}

	{
		content, err := afero.ReadFile(fs, metaPath)
		if err != nil {
			return nil, err
		}
		meta, err := MetaFromString(string(content))
		if err != nil {
			return nil, err
		}
		b.meta = meta
	}

	// Load meta critique if it exists
	metaCritPath := filepath.Join(path, "meta_critique.md")
	if exists, err := afero.Exists(fs, metaCritPath); err != nil {
		return nil, err
	} else if exists {
		content, err := afero.ReadFile(fs, metaCritPath)
		if err != nil {
			return nil, err
		}
		crit, err := CritiqueFromString(string(content))
		if err != nil {
			return nil, err
		}
		b.metaCrit = crit
	}

	// Load plan if it exists
	planPath := filepath.Join(path, "plan.md")
	if exists, err := afero.Exists(fs, planPath); err != nil {
		return nil, err
	} else if exists {
		content, err := afero.ReadFile(fs, planPath)
		if err != nil {
			return nil, err
		}
		plan, err := PlanFromString(string(content))
		if err != nil {
			return nil, err
		}
		b.plan = plan
	}

	// Load plan critique if it exists
	planCritPath := filepath.Join(path, "plan_critique.md")
	if exists, err := afero.Exists(fs, planCritPath); err != nil {
		return nil, err
	} else if exists {
		content, err := afero.ReadFile(fs, planCritPath)
		if err != nil {
			return nil, err
		}
		crit, err := CritiqueFromString(string(content))
		if err != nil {
			return nil, err
		}
		b.planCrit = crit
	}

	if b.plan.Count() == 0 {
		return b, nil
	}

	b.chps = make([]*Chapter, b.plan.Count())
	b.chpsCrit = make([]*Critique, b.plan.Count())

	for i := 1; i <= b.plan.Count(); i++ {
		chpPath := filepath.Join(path, fmt.Sprintf("chapter_%d.md", i))
		if exists, err := afero.Exists(fs, chpPath); err != nil {
			return nil, err
		} else if exists {
			content, err := afero.ReadFile(fs, chpPath)
			if err != nil {
				return nil, err
			}
			chp, err := ChapterFromString(string(content))
			if err != nil {
				continue
			}
			b.chps[i-1] = chp
		}
	}

	for i := 1; i <= b.plan.Count(); i++ {
		critPath := filepath.Join(path, fmt.Sprintf("chapter_%d_critique.md", i))
		if exists, err := afero.Exists(fs, critPath); err != nil {
			return nil, err
		} else if exists {
			content, err := afero.ReadFile(fs, critPath)
			if err != nil {
				return nil, err
			}
			crit, err := CritiqueFromString(string(content))
			if err != nil {
				continue
			}
			b.chpsCrit[i-1] = crit
		}
	}

	return b, nil
}

// Save saves the book's metadata and plan (if it's not empty) to the filesystem.
func (b *Book) Save() error {
	// Ensure directory exists
	if err := b.fs.MkdirAll(b.path, 0755); err != nil {
		return err
	}

	if err := b.saveMeta(); err != nil {
		return err
	}

	if len(b.plan) > 0 {
		if err := b.savePlan(); err != nil {
			return err
		}
	}

	if b.metaCrit != nil {
		if err := b.saveMetaCrit(); err != nil {
			return err
		}
	}

	if b.planCrit != nil {
		if err := b.savePlanCrit(); err != nil {
			return err
		}
	}

	return nil
}

func (b *Book) saveMeta() error {
	metaPath := filepath.Join(b.path, "meta.md")
	if err := afero.WriteFile(b.fs, metaPath, []byte(b.meta.String()), 0644); err != nil {
		return err
	}
	return nil
}

func (b *Book) savePlan() error {
	planPath := filepath.Join(b.path, "plan.md")
	if err := afero.WriteFile(b.fs, planPath, []byte(b.plan.String()), 0644); err != nil {
		return err
	}
	return nil
}

func (b *Book) saveMetaCrit() error {
	metaCritPath := filepath.Join(b.path, "meta_critique.md")
	if err := afero.WriteFile(b.fs, metaCritPath, []byte(b.metaCrit.String()), 0644); err != nil {
		return err
	}
	return nil
}

func (b *Book) savePlanCrit() error {
	planCritPath := filepath.Join(b.path, "plan_critique.md")
	if err := afero.WriteFile(b.fs, planCritPath, []byte(b.planCrit.String()), 0644); err != nil {
		return err
	}
	return nil
}

// SetMeta sets the book's metadata and saves it to the filesystem.
func (b *Book) SetMeta(meta *BookMeta) error {
	if upd := b.meta.merge(meta); upd {
		return b.saveMeta()
	}

	return nil
}

func (b *Book) GetMeta() *BookMeta {
	return b.meta
}

// SetPlan sets the book's plan and saves it to the filesystem.
func (b *Book) SetPlan(plan Plan) error {
	b.plan = b.plan.merge(plan)
	return b.savePlan()
}

func (b *Book) GetPlan() Plan {
	return b.plan
}

func (b *Book) GetPlanChapter(i int) (*Chapter, error) {
	if i <= 0 || i > len(b.plan) {
		return &Chapter{}, ErrInvalidChapterIndex
	}
	return b.plan[i-1], nil
}

func (b *Book) GetChapter(i int) (*Chapter, error) {
	if i <= 0 || i > len(b.plan) {
		return &Chapter{}, ErrInvalidChapterIndex
	}

	for i > len(b.chps) {
		b.chps = append(b.chps, &Chapter{})
	}
	return b.chps[i-1], nil
}

// SetChapter saves the given chapter to the filesystem.
// It creates a new file 'chapter_<index>.md' if it doesn't exist, or overwrites it if it does.
func (b *Book) SetChapter(i int, chapter *Chapter) error {
	if i <= 0 || i > len(b.plan) {
		return ErrInvalidChapterIndex
	}

	for i > len(b.chps) {
		b.chps = append(b.chps, &Chapter{})
	}
	b.chps[i-1] = chapter

	filename := fmt.Sprintf("chapter_%d.md", i)
	chapterPath := filepath.Join(b.path, filename)
	return afero.WriteFile(b.fs, chapterPath, []byte(chapter.String()), 0644)
}

func (b *Book) GetChapterCritique(i int) (*Critique, error) {
	if i <= 0 || i > len(b.plan) {
		return nil, ErrInvalidChapterIndex
	}

	for i > len(b.chpsCrit) {
		b.chpsCrit = append(b.chpsCrit, &Critique{})
	}
	return b.chpsCrit[i-1], nil
}

func (b *Book) SetChapterCritique(i int, crit *Critique) error {
	if i <= 0 || i > len(b.plan) {
		return ErrInvalidChapterIndex
	}

	for i > len(b.chpsCrit) {
		b.chpsCrit = append(b.chpsCrit, &Critique{})
	}
	b.chpsCrit[i-1] = crit

	filename := fmt.Sprintf("chapter_%d_critique.md", i)
	critPath := filepath.Join(b.path, filename)
	return afero.WriteFile(b.fs, critPath, []byte(crit.String()), 0644)
}

func (b *Book) GetFs() afero.Fs {
	return b.fs
}

func (b *Book) GetFilePath() string {
	return b.path
}

// SetMetaCrit sets the book's meta critique and saves it to the filesystem.
func (b *Book) SetMetaCrit(metaCrit *Critique) error {
	b.metaCrit = metaCrit
	return b.saveMetaCrit()
}

func (b *Book) GetMetaCrit() *Critique {
	return b.metaCrit
}

// SetPlanCrit sets the book's plan critique and saves it to the filesystem.
func (b *Book) SetPlanCrit(planCrit *Critique) error {
	b.planCrit = planCrit
	return b.savePlanCrit()
}

func (b *Book) GetPlanCrit() *Critique {
	return b.planCrit
}

func (b *Book) ExportMarkdown() error {
	if b.plan == nil || b.plan.Count() < len(b.chps) {
		return ErrBookNotFinished
	}

	title := b.getTitle()

	var text strings.Builder
	text.WriteString("# ")
	text.WriteString(title)
	text.WriteString("\n\n")

	for _, chp := range b.chps {
		text.WriteString("### ")
		text.WriteString(strconv.Itoa(chp.GetNumber()))
		text.WriteString(". ")
		text.WriteString(chp.GetTitle())
		text.WriteString("\n\n")
		text.WriteString(chp.GetContent())
		text.WriteString("\n\n")
	}

	mdPath := filepath.Join(b.path, fmt.Sprintf("%s.md", title))
	return afero.WriteFile(b.fs, mdPath, []byte(text.String()), 0644)
}

func (b *Book) ExportHTML() error {
	if b.plan == nil || b.plan.Count() < len(b.chps) {
		return ErrBookNotFinished
	}

	title := b.getTitle()

	var text strings.Builder
	text.WriteString("<meta charset=\"utf-8\">")
	text.WriteString("<title>")
	text.WriteString(title)
	text.WriteString("</title>")

	text.WriteString("<h1>")
	text.WriteString(title)
	text.WriteString("</h1>")

	for _, chp := range b.chps {
		text.WriteString("<h3>")
		text.WriteString(strconv.Itoa(chp.GetNumber()))
		text.WriteString(". ")
		text.WriteString(chp.GetTitle())
		text.WriteString("</h3>")
		ps := strings.SplitSeq(chp.GetContent(), "\n")
		for p := range ps {
			if len(p) == 0 {
				continue
			}
			text.WriteString("<p>")
			text.WriteString(p)
			text.WriteString("</p>")
		}
		text.WriteString("<hr>")
	}

	htmlPath := filepath.Join(b.path, fmt.Sprintf("%s.html", title))
	return afero.WriteFile(b.fs, htmlPath, []byte(text.String()), 0644)
}

func (b *Book) ExportEPUB() error {
	if b.plan == nil || b.plan.Count() < len(b.chps) {
		return ErrBookNotFinished
	}

	title := b.getTitle()

	ep, err := epub.NewEpub(title)
	if err != nil {
		return err
	}

	ep.SetAuthor("Viter — AI novel generator")
	ep.SetTitle(title)
	ep.SetDescription(strings.Join(b.meta.GetGenres(), ", "))

	for _, chp := range b.chps {
		chpTitle := fmt.Sprintf("%d. %s", chp.GetNumber(), chp.GetTitle())
		var content strings.Builder
		content.WriteString("<h3>")
		content.WriteString(strconv.Itoa(chp.GetNumber()))
		content.WriteString(". ")
		content.WriteString(chp.GetTitle())
		content.WriteString("</h3>")
		ps := strings.SplitSeq(chp.GetContent(), "\n")
		for p := range ps {
			if len(p) == 0 {
				continue
			}
			content.WriteString("<p>")
			content.WriteString(p)
			content.WriteString("</p>")
		}
		_, err = ep.AddSection(content.String(), chpTitle, "", "")
		if err != nil {
			return err
		}
	}

	epubPath := filepath.Join(b.path, fmt.Sprintf("%s.epub", title))
	f, err := b.fs.OpenFile(epubPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = ep.WriteTo(f)
	return err
}

func (b *Book) getTitle() string {
	title := b.meta.GetTitle()
	if strings.Contains(title, "/") {
		parts := strings.Split(title, "/")
		title = parts[0]
		for i := 1; i < len(parts); i++ {
			if len(parts[i]) > len(title) {
				title = parts[i]
			}
		}
		title = strings.TrimSpace(title)
	}
	return title
}
