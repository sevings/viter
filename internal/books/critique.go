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

	sep := "\n##"
	if !strings.Contains(s, sep) {
		sep = "\n**"
	}

	// Split by sections
	sections := strings.SplitSeq(s, sep)
	for section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		lines := strings.SplitN(section, "\n", 2)
		if len(lines) < 2 {
			continue
		}

		header := trimTitle(lines[0])
		content := strings.TrimSpace(lines[1])

		switch header {
		case "Strengths":
			c.Strengths = content
		case "Improvements":
			c.Improvements = content
		case "Impressions":
			c.Impressions = content
		case "Score":
			if score, err := strconv.Atoi(content); err == nil {
				c.Score = score
			}
		}
	}

	return c, nil
}

func (c *Critique) String() string {
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

func (c *Critique) GetStrengths() string {
	return c.Strengths
}

func (c *Critique) GetImprovements() string {
	return c.Improvements
}

func (c *Critique) GetImpressions() string {
	return c.Impressions
}

func (c *Critique) GetScore() int {
	return c.Score
}

func (c *Critique) SetStrengths(strengths string) {
	c.Strengths = strengths
}

func (c *Critique) SetImprovements(improvements string) {
	c.Improvements = improvements
}

func (c *Critique) SetImpressions(impressions string) {
	c.Impressions = impressions
}

func (c *Critique) SetScore(score int) {
	c.Score = score
}
