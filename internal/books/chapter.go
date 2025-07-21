package books

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrNoTitleFound       = errors.New("no title found in chapter string")
	ErrEmptyChapterString = errors.New("empty chapter string")
)

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
