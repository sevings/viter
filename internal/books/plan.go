package books

import "strings"

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

func (p Plan) MergedCopy(other Plan) Plan {
	return p.merge(other)
}

func (p Plan) Count() int {
	return len(p)
}
