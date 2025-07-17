package viter

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var (
	ErrInvalidChapterIndex = errors.New("invalid chapter index")
	ErrEmptyChapterString  = errors.New("empty chapter string")
	ErrNoTitleFound        = errors.New("no title found in chapter string")
)

type Book struct {
	fs   afero.Fs
	path string
	meta BookMeta
	plan Plan
}

// CreateBook creates a new book with the given filesystem and path.
// it creates a file 'meta.md' and initializes the book's metadata.
func CreateBook(fs afero.Fs, path string) (*Book, error) {
	b := &Book{fs: fs, path: path, meta: BookMeta{}}
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
	} else if exists {
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

// SetMeta sets the book's metadata and saves it to the filesystem.
func (b *Book) SetMeta(meta BookMeta) error {
	b.meta = meta
	return b.saveMeta()
}

func (b *Book) GetMeta() BookMeta {
	return b.meta
}

// SetPlan sets the book's plan and saves it to the filesystem.
func (b *Book) SetPlan(plan Plan) error {
	b.plan = plan
	return b.savePlan()
}

func (b *Book) GetPlan() Plan {
	return b.plan
}

func (b *Book) GetPlanChapter(i int) (Chapter, error) {
	if i < 0 || i >= len(b.plan) {
		return Chapter{}, ErrInvalidChapterIndex
	}
	return b.plan[i], nil
}

// SaveChapter saves the given chapter to the filesystem.
// It creates a new file 'chapter_<index>.md' if it doesn't exist, or overwrites it if it does.
func (b *Book) SaveChapter(i int, chapter Chapter) error {
	if i < 0 || i >= len(b.plan) {
		return ErrInvalidChapterIndex
	}

	filename := fmt.Sprintf("chapter_%d.md", i)
	chapterPath := filepath.Join(b.path, filename)
	return afero.WriteFile(b.fs, chapterPath, []byte(chapter.String()), 0644)
}

type BookMeta struct {
	style      string
	genres     []string
	world      string
	mainChars  []Character
	minorChars []Character
	plot       string
}

type Character struct {
	name string
	desc string
}

func NewCharacter(name, desc string) Character {
	return Character{name: name, desc: desc}
}

func (c Character) GetName() string {
	return c.name
}

func (c Character) GetDesc() string {
	return c.desc
}

// MetaFromString creates a new BookMeta from a string.
// Format:
// ```
// ## Style
// {style}
// ## Genres
// {genres comma separated}
// ## World
// {world}
// ## Main Characters
// ### {name}
// {description}
// ## Minor Characters
// ### {name}
// {description}
// ## Plot
// {plot}
// ```
func MetaFromString(s string) (BookMeta, error) {
	meta := BookMeta{}

	// Split by sections
	sections := strings.SplitSeq(s, "\n## ")
	for section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		// Add back the ## if it was removed by split
		if !strings.HasPrefix(section, "## ") {
			section = "## " + section
		}

		lines := strings.SplitN(section, "\n", 2)
		if len(lines) < 2 {
			continue
		}

		header := strings.TrimSpace(lines[0])
		content := strings.TrimSpace(lines[1])

		switch header {
		case "## Style":
			meta.style = content
		case "## Genres":
			genres := strings.Split(content, ",")
			for i, genre := range genres {
				genres[i] = strings.TrimSpace(genre)
			}
			meta.genres = genres
		case "## World":
			meta.world = content
		case "## Main Characters":
			meta.mainChars = parseCharacters(content)
		case "## Minor Characters":
			meta.minorChars = parseCharacters(content)
		case "## Plot":
			meta.plot = content
		}
	}

	return meta, nil
}

func parseCharacters(s string) []Character {
	var characters []Character

	// Split by ### to find character sections
	parts := strings.SplitSeq(s, "\n### ")
	for part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Add back the ### if it was removed by split
		if !strings.HasPrefix(part, "### ") {
			part = "### " + part
		}

		lines := strings.SplitN(part, "\n", 2)
		if len(lines) < 2 {
			continue
		}

		name := strings.TrimSpace(strings.TrimPrefix(lines[0], "### "))
		desc := strings.TrimSpace(lines[1])

		if name != "" {
			characters = append(characters, Character{
				name: name,
				desc: desc,
			})
		}
	}

	return characters
}

func (bm *BookMeta) String() string {
	var sb strings.Builder

	sb.WriteString("## Style\n")
	sb.WriteString(bm.style)
	sb.WriteString("\n\n")

	sb.WriteString("## Genres\n")
	sb.WriteString(strings.Join(bm.genres, ", "))
	sb.WriteString("\n\n")

	sb.WriteString("## World\n")
	sb.WriteString(bm.world)
	sb.WriteString("\n\n")

	sb.WriteString("## Main Characters\n")
	for _, char := range bm.mainChars {
		sb.WriteString("### ")
		sb.WriteString(char.name)
		sb.WriteString("\n")
		sb.WriteString(char.desc)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Minor Characters\n")
	for _, char := range bm.minorChars {
		sb.WriteString("### ")
		sb.WriteString(char.name)
		sb.WriteString("\n")
		sb.WriteString(char.desc)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Plot\n")
	sb.WriteString(bm.plot)

	return sb.String()
}

func (bm *BookMeta) GetStyle() string {
	return bm.style
}

func (bm *BookMeta) GetGenres() []string {
	return bm.genres
}

func (bm *BookMeta) GetWorld() string {
	return bm.world
}

func (bm *BookMeta) GetMainCharacters() []Character {
	return bm.mainChars
}

func (bm *BookMeta) GetMinorCharacters() []Character {
	return bm.minorChars
}

func (bm *BookMeta) GetPlot() string {
	return bm.plot
}

func (bm *BookMeta) SetStyle(style string) {
	bm.style = style
}

func (bm *BookMeta) SetGenres(genres []string) {
	bm.genres = genres
}

func (bm *BookMeta) SetWorld(world string) {
	bm.world = world
}

func (bm *BookMeta) SetMainCharacters(chars []Character) {
	bm.mainChars = chars
}

func (bm *BookMeta) SetMinorCharacters(chars []Character) {
	bm.minorChars = chars
}

func (bm *BookMeta) SetPlot(plot string) {
	bm.plot = plot
}

type Chapter struct {
	title   string
	content string
}

func NewChapter(title, content string) Chapter {
	return Chapter{title: title, content: content}
}

// ChapterFromString creates a new Chapter from a string.
// Format:
// ```
// ## {title}
// {content}
// ```
func ChapterFromString(s string) (Chapter, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Chapter{}, ErrEmptyChapterString
	}

	lines := strings.Split(s, "\n")
	if len(lines) < 1 {
		return Chapter{}, ErrEmptyChapterString
	}

	// Find first line starting with ##
	titleLine := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "## ") {
			titleLine = i
			break
		}
	}

	if titleLine == -1 {
		return Chapter{}, ErrNoTitleFound
	}

	title := strings.TrimSpace(lines[titleLine][3:]) // Remove "## "

	// Content is everything after the title line
	var content strings.Builder
	for i := titleLine + 1; i < len(lines); i++ {
		if i > titleLine+1 {
			content.WriteString("\n")
		}
		content.WriteString(lines[i])
	}

	return Chapter{
		title:   title,
		content: strings.TrimSpace(content.String()),
	}, nil
}

func (c *Chapter) String() string {
	var sb strings.Builder
	sb.WriteString("## ")
	sb.WriteString(c.title)
	sb.WriteString("\n")
	sb.WriteString(c.content)
	return sb.String()
}

func (c Chapter) GetTitle() string {
	return c.title
}

func (c Chapter) GetContent() string {
	return c.content
}

type Plan []Chapter

// PlanFromString creates a new Plan from a string.
// Format:
// ```
// ## {title}
// {content}
// ```
func PlanFromString(s string) (Plan, error) {
	var plan Plan

	if s == "" {
		return Plan{}, nil
	}

	parts := strings.SplitSeq(s, "\n## ")
	for part := range parts {
		if !strings.HasPrefix(part, "## ") {
			part = "## " + part
		}
		chapter, err := ChapterFromString(part)
		if err != nil {
			return nil, err
		}
		plan = append(plan, chapter)
	}

	return plan, nil
}

func (p Plan) String() string {
	var sb strings.Builder
	for i, chapter := range p {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(chapter.String())
	}
	return sb.String()
}
