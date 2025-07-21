package books

import "strings"

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
	if other == nil {
		return false
	}
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

func (bm *BookMeta) MergedCopy(other *BookMeta) *BookMeta {
	newBM := &BookMeta{
		style:        bm.style,
		genres:       bm.genres,
		logline:      bm.logline,
		world:        bm.world,
		protagonists: bm.protagonists,
		antagonists:  bm.antagonists,
		minorChars:   bm.minorChars,
		plot:         bm.plot,
		title:        bm.title,
	}
	newBM.merge(other)
	return newBM
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
