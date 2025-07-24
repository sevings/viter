package books

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/go-epub"
	"github.com/iris-contrib/blackfriday"
	"github.com/spf13/afero"
)

var (
	ErrInvalidChapterIndex = errors.New("invalid chapter index")
	ErrNoMetaFile          = errors.New("no meta file found")
	ErrBookNotFinished     = errors.New("book not finished")
	ErrFileNotFound        = errors.New("file not found")
)

func trimTitle(s string) string {
	return strings.Trim(s, "#*\"«» \n\r\t")
}

type Book struct {
	fs   afero.Fs
	path string
	meta *BookMeta
	plan Plan
	chps []*Chapter

	metaCrit *Critique
	planCrit *Critique
	chpsCrit []*Critique

	arch bool
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
	if content, err := b.readFile("meta.md"); err != nil && err != ErrFileNotFound {
		return nil, err
	} else if err == ErrFileNotFound {
		return nil, ErrNoMetaFile
	} else {
		meta, err := MetaFromString(string(content))
		if err != nil {
			return nil, err
		}
		b.meta = meta
	}

	// Load meta critique if it exists
	if content, err := b.readFile("meta_critique.md"); err != nil && err != ErrFileNotFound {
		return nil, err
	} else if err == nil {
		crit, err := CritiqueFromString(string(content))
		if err != nil {
			return nil, err
		}
		b.metaCrit = crit
	}

	// Load plan if it exists
	if content, err := b.readFile("plan.md"); err != nil && err != ErrFileNotFound {
		return nil, err
	} else if err == nil {
		plan, err := PlanFromString(string(content))
		if err != nil {
			return nil, err
		}
		b.plan = plan
	}

	// Load plan critique if it exists
	if content, err := b.readFile("plan_critique.md"); err != nil && err != ErrFileNotFound {
		return nil, err
	} else if err == nil {
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
		chpName := fmt.Sprintf("chapter_%d.md", i)
		if content, err := b.readFile(chpName); err != nil && err != ErrFileNotFound {
			return nil, err
		} else if err == nil {
			chp, err := ChapterFromString(string(content))
			if err != nil {
				continue
			}
			b.chps[i-1] = chp
		}
	}

	for i := 1; i <= b.plan.Count(); i++ {
		critName := fmt.Sprintf("chapter_%d_critique.md", i)
		if content, err := b.readFile(critName); err != nil && err != ErrFileNotFound {
			return nil, err
		} else if err == nil {
			crit, err := CritiqueFromString(string(content))
			if err != nil {
				continue
			}
			b.chpsCrit[i-1] = crit
		}
	}

	return b, nil
}

func (b *Book) EnableArchiving() {
	archPath := b.getFilePath("archived")
	if err := b.fs.MkdirAll(archPath, 0755); err != nil {
		return
	}

	b.arch = true
}

func (b *Book) DisableArchiving() {
	b.arch = false
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
	return b.writeFile("meta.md", []byte(b.meta.String()))
}

func (b *Book) savePlan() error {
	return b.writeFile("plan.md", []byte(b.plan.String()))
}

func (b *Book) saveMetaCrit() error {
	return b.writeFile("meta_critique.md", []byte(b.metaCrit.String()))
}

func (b *Book) savePlanCrit() error {
	return b.writeFile("plan_critique.md", []byte(b.planCrit.String()))
}

func (b *Book) writeFile(name string, content []byte) error {
	filePath := b.getFilePath(name)

	if b.arch {
		archName := strings.ReplaceAll(name, ".", time.Now().Format("_2006-01-02_15-04-05."))
		archPath := filepath.Join(b.path, "archived", archName)
		b.fs.Rename(filePath, archPath)
	}

	return afero.WriteFile(b.fs, filePath, content, 0644)
}

func (b *Book) readFile(name string) ([]byte, error) {
	filePath := b.getFilePath(name)
	if exists, err := afero.Exists(b.fs, filePath); err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrFileNotFound
	}
	return afero.ReadFile(b.fs, filePath)
}

func (b *Book) getFilePath(name string) string {
	return filepath.Join(b.path, name)
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

	return b.writeFile(fmt.Sprintf("chapter_%d.md", i), []byte(chapter.String()))
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

	return b.writeFile(fmt.Sprintf("chapter_%d_critique.md", i), []byte(crit.String()))
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

func (b *Book) Markdown() (string, error) {
	if b.plan == nil || b.plan.Count() < len(b.chps) {
		return "", ErrBookNotFinished
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

	return text.String(), nil
}

func (b *Book) ExportMarkdown() error {
	md, err := b.Markdown()
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s.md", b.getTitle())
	return b.writeFile(fileName, []byte(md))
}

func (b *Book) ExportHTML() error {
	md, err := b.Markdown()
	if err != nil {
		return err
	}

	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.Smartypants |
			blackfriday.SmartypantsAngledQuotes |
			blackfriday.SmartypantsDashes |
			blackfriday.TOC,
	})
	html := blackfriday.Run([]byte(md), blackfriday.WithRenderer(renderer))

	title := b.getTitle()
	html = fmt.Appendf(nil, "<html><meta charset=\"utf-8\"><head><title>\n%s\n</title></head>\n<body>\n%s\n</body></html>", title, string(html))
	fileName := fmt.Sprintf("%s.html", title)
	return b.writeFile(fileName, html)
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
		content := fmt.Sprintf("### %s\n\n%s\n\n", chpTitle, chp.GetContent())

		renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			Flags: blackfriday.Smartypants |
				blackfriday.SmartypantsAngledQuotes |
				blackfriday.SmartypantsDashes,
		})
		html := blackfriday.Run([]byte(content), blackfriday.WithRenderer(renderer))

		_, err = ep.AddSection(string(html), chpTitle, "", "")
		if err != nil {
			return err
		}
	}

	epubPath := b.getFilePath(fmt.Sprintf("%s.epub", title))
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
