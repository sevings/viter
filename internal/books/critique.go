package books

import (
	"strconv"
	"strings"
)

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
