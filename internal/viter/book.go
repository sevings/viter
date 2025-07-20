package viter

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

var (
	ErrInvalidChapterIndex = errors.New("invalid chapter index")
	ErrEmptyChapterString  = errors.New("empty chapter string")
	ErrNoTitleFound        = errors.New("no title found in chapter string")
	ErrNoMetaFile          = errors.New("no meta file found")
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

type BookMeta struct {
	style        string
	genres       []string
	logline      string
	world        string
	protagonists []Character
	antagonists  []Character
	minorChars   []Character
	plot         string
	title        string
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
// ## Logline
// {logline}
// ## World
// {world}
// ## Protagonists
// ### {name}
// {description}
// ## Antagonists
// ### {name}
// {description}
// ## Minor Characters
// ### {name}
// {description}
// ## Plot
// {plot}
// ## Title
// {title}
// ```
func MetaFromString(s string) (*BookMeta, error) {
	meta := &BookMeta{}

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
		case "## Logline":
			meta.logline = content
		case "## World":
			meta.world = content
		case "## Protagonists":
			meta.protagonists = parseCharacters(content)
		case "## Antagonists":
			meta.antagonists = parseCharacters(content)
		case "## Minor Characters":
			meta.minorChars = parseCharacters(content)
		case "## Plot":
			meta.plot = content
		case "## Title":
			meta.title = content
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

	sb.WriteString("## Logline\n")
	sb.WriteString(bm.logline)
	sb.WriteString("\n\n")

	sb.WriteString("## World\n")
	sb.WriteString(bm.world)
	sb.WriteString("\n\n")

	sb.WriteString("## Protagonists\n")
	for _, char := range bm.protagonists {
		sb.WriteString("### ")
		sb.WriteString(char.name)
		sb.WriteString("\n")
		sb.WriteString(char.desc)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## Antagonists\n")
	for _, char := range bm.antagonists {
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
	sb.WriteString("\n\n")

	sb.WriteString("## Title\n")
	sb.WriteString(bm.title)

	return sb.String()
}

func (bm *BookMeta) merge(other *BookMeta) bool {
	changed := false
	if other.style != "" {
		bm.style = other.style
		changed = true
	}
	if len(other.genres) > 0 {
		bm.genres = other.genres
		changed = true
	}
	if other.logline != "" {
		bm.logline = other.logline
		changed = true
	}
	if other.world != "" {
		bm.world = other.world
		changed = true
	}
	if len(other.protagonists) > 0 {
		bm.protagonists = other.protagonists
		changed = true
	}
	if len(other.antagonists) > 0 {
		bm.antagonists = other.antagonists
		changed = true
	}
	if len(other.minorChars) > 0 {
		bm.minorChars = other.minorChars
		changed = true
	}
	if other.plot != "" {
		bm.plot = other.plot
		changed = true
	}
	if other.title != "" {
		bm.title = other.title
		changed = true
	}
	return changed
}

func (bm *BookMeta) GetStyle() string {
	return bm.style
}

func (bm *BookMeta) GetGenres() []string {
	return bm.genres
}

func (bm *BookMeta) GetLogline() string {
	return bm.logline
}

func (bm *BookMeta) GetWorld() string {
	return bm.world
}

func (bm *BookMeta) GetProtagonists() []Character {
	return bm.protagonists
}

func (bm *BookMeta) GetAntagonists() []Character {
	return bm.antagonists
}

func (bm *BookMeta) GetMinorCharacters() []Character {
	return bm.minorChars
}

func (bm *BookMeta) GetPlot() string {
	return bm.plot
}

func (bm *BookMeta) GetTitle() string {
	return bm.title
}

func (bm *BookMeta) SetStyle(style string) {
	bm.style = style
}

func (bm *BookMeta) SetGenres(genres []string) {
	bm.genres = genres
}

func (bm *BookMeta) SetLogline(logline string) {
	bm.logline = logline
}

func (bm *BookMeta) SetWorld(world string) {
	bm.world = world
}

func (bm *BookMeta) SetProtagonists(chars []Character) {
	bm.protagonists = chars
}

func (bm *BookMeta) SetAntagonists(antagonists []Character) {
	bm.antagonists = antagonists
}

func (bm *BookMeta) SetMinorCharacters(chars []Character) {
	bm.minorChars = chars
}

func (bm *BookMeta) SetPlot(plot string) {
	bm.plot = plot
}

func (bm *BookMeta) SetTitle(title string) {
	bm.title = title
}

func (bm *BookMeta) IsFilled() bool {
	return bm.style != "" &&
		len(bm.genres) > 0 &&
		bm.logline != "" &&
		bm.world != "" &&
		len(bm.protagonists) > 0 &&
		len(bm.antagonists) > 0 &&
		len(bm.minorChars) > 0 &&
		bm.plot != "" &&
		bm.title != ""
}

type Chapter struct {
	number  int
	title   string
	content string
}

func NewChapter(number int, title, content string) *Chapter {
	return &Chapter{number: number, title: title, content: content}
}

// ChapterFromString creates a new Chapter from a string.
// Format:
// ```
// ## {number}. {title}
// {content}
// ```
func ChapterFromString(s string) (*Chapter, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return &Chapter{}, ErrEmptyChapterString
	}

	lines := strings.Split(s, "\n")
	if len(lines) < 1 {
		return &Chapter{}, ErrEmptyChapterString
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
		return &Chapter{}, ErrNoTitleFound
	}

	titleText := strings.TrimSpace(lines[titleLine][3:]) // Remove "## "

	// Parse chapter number from title if it follows the format "number. title"
	var number int
	var title string

	if dotIndex := strings.Index(titleText, ". "); dotIndex > 0 {
		// Try to parse the number part
		if num, err := strconv.Atoi(titleText[:dotIndex]); err == nil {
			number = num
			title = titleText[dotIndex+2:] // Skip ". "
		} else {
			// If parsing fails, use the whole string as title
			title = titleText
		}
	} else {
		// No number format found, use whole string as title
		title = titleText
	}

	// Content is everything after the title line
	var content strings.Builder
	for i := titleLine + 1; i < len(lines); i++ {
		if i > titleLine+1 {
			content.WriteString("\n")
		}
		content.WriteString(lines[i])
	}

	return &Chapter{
		number:  number,
		title:   title,
		content: strings.TrimSpace(content.String()),
	}, nil
}

func (c *Chapter) String() string {
	var sb strings.Builder
	sb.WriteString("## ")
	if c.number > 0 {
		sb.WriteString(strconv.Itoa(c.number))
		sb.WriteString(". ")
	}
	sb.WriteString(c.title)
	sb.WriteString("\n")
	sb.WriteString(c.content)
	return sb.String()
}

func (c *Chapter) GetTitle() string {
	return c.title
}

func (c *Chapter) GetContent() string {
	return c.content
}

func (c *Chapter) GetNumber() int {
	return c.number
}

func (c *Chapter) SetNumber(number int) {
	c.number = number
}

func (c *Chapter) SetTitle(title string) {
	c.title = title
}

func (c *Chapter) SetContent(content string) {
	c.content = content
}

type Plan []*Chapter

// PlanFromString creates a new Plan from a string.
// Format:
// ```
// ## {number}. {title}
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

// merge combines this plan with another, with the other plan taking precedence
// for chapters that exist in both plans (based on chapter number).
// Chapters with number 0 are always appended.
func (p Plan) merge(other Plan) Plan {
	if len(other) == 0 {
		return p
	}

	// Create a map of existing chapters by number for efficient lookup
	existing := make(map[int]int) // number -> index
	for i, chapter := range p {
		if chapter.number > 0 {
			existing[chapter.number] = i
		}
	}

	result := make(Plan, len(p))
	copy(result, p)

	// Process chapters from other plan
	for _, otherChapter := range other {
		if otherChapter.number == 0 {
			// Chapters with number 0 are always appended
			result = append(result, otherChapter)
		} else if existingIndex, exists := existing[otherChapter.number]; exists {
			// Replace existing chapter with same number
			result[existingIndex] = otherChapter
		} else {
			// Add new chapter and update the map
			result = append(result, otherChapter)
			existing[otherChapter.number] = len(result) - 1
		}
	}

	return result
}

func (p Plan) Count() int {
	return len(p)
}

type Critique struct {
	Strengths    string
	Improvements string
	Impressions  string
	Score        int
}

// CritiqueFromString parses a string into a Critique struct.
// Format:
// ```
// ## Strengths
// {content}
// ## Improvements
// {content}
// ## Impressions
// {content}
// ## Score
// {number}
// ```
func CritiqueFromString(s string) (*Critique, error) {
	c := &Critique{}

	if s == "" {
		return c, nil
	}

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

		// Handle "## Score" format with number on same or next line
		if strings.HasPrefix(section, "## Score") {
			// Check if score is on same line after space
			if len(section) > 8 && section[8] == ' ' {
				scoreStr := strings.TrimSpace(section[9:])
				if score, err := strconv.Atoi(scoreStr); err == nil {
					c.Score = score
				}
			} else {
				// Score might be on next line, so don't continue - let it fall through
			}
		}

		lines := strings.SplitN(section, "\n", 2)
		if len(lines) < 2 {
			continue
		}

		header := strings.TrimSpace(lines[0])
		content := strings.TrimSpace(lines[1])

		switch header {
		case "## Strengths":
			c.Strengths = content
		case "## Improvements":
			c.Improvements = content
		case "## Impressions":
			c.Impressions = content
		case "## Score":
			// Handle "## Score\n6" format (score on next line)
			if score, err := strconv.Atoi(content); err == nil {
				c.Score = score
			}
		}
	}

	return c, nil
}

func (c Critique) String() string {
	var sb strings.Builder
	sb.WriteString("## Strengths\n")
	sb.WriteString(c.Strengths)
	sb.WriteString("\n\n## Improvements\n")
	sb.WriteString(c.Improvements)
	sb.WriteString("\n\n## Impressions\n")
	sb.WriteString(c.Impressions)
	sb.WriteString("\n\n## Score\n")
	sb.WriteString(strconv.Itoa(c.Score))
	return sb.String()
}

func (c Critique) GetStrengths() string {
	return c.Strengths
}

func (c Critique) GetImprovements() string {
	return c.Improvements
}

func (c Critique) GetImpressions() string {
	return c.Impressions
}

func (c Critique) GetScore() int {
	return c.Score
}
