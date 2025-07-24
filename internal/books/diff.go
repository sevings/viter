package books

import (
	"fmt"
	"maps"
	"regexp"
	"strconv"
	"strings"
)

type Diff struct {
	parts map[int]string
}

func DiffFromText(text string) *Diff {
	lines := strings.Split(text, "\n")
	parts := make(map[int]string)
	for i, line := range lines {
		parts[i+1] = line
	}
	return &Diff{parts: parts}
}

func (d *Diff) Text() string {
	if len(d.parts) == 0 {
		return ""
	}

	// Find the maximum index to determine the size
	maxIndex := 0
	for i := range d.parts {
		if i > maxIndex {
			maxIndex = i
		}
	}

	lines := make([]string, maxIndex)
	for i := 1; i <= maxIndex; i++ {
		lines[i-1] = d.parts[i]
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] != "" && lines[i-1] != "" {
			lines[i-1] += "\n"
		}
	}
	lines[len(lines)-1] += "\n"

	return strings.Join(lines, "\n")
}

func DiffFromString(text string) (*Diff, error) {
	parts := make(map[int]string)

	// Use regex to find all <n>content</n> patterns
	re := regexp.MustCompile(`<(\d+)>(.*?)</\d+>`)
	matches := re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) == 3 {
			index, err := strconv.Atoi(match[1])
			if err != nil {
				// Skip invalid indices instead of returning error
				continue
			}
			parts[index] = match[2]
		}
	}

	return &Diff{parts: parts}, nil
}

func (d *Diff) String() string {
	if len(d.parts) == 0 {
		return "<1></1>"
	}

	// Find the maximum index to determine the size
	maxIndex := 0
	for i := range d.parts {
		if i > maxIndex {
			maxIndex = i
		}
	}

	lines := make([]string, maxIndex)
	for i := 1; i <= maxIndex; i++ {
		if line, exists := d.parts[i]; exists {
			lines[i-1] = fmt.Sprintf("<%d>%s</%d>", i, line, i)
		} else {
			lines[i-1] = fmt.Sprintf("<%d></%d>", i, i)
		}
	}

	return strings.Join(lines, "\n")
}

func (d *Diff) Merge(other *Diff) {
	if other == nil {
		return
	}

	maps.Copy(d.parts, other.parts)
}
