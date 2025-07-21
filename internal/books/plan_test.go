package books_test

import (
	"testing"
	"viter/internal/books"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestPlanFromString(t *testing.T) {
	planStr := `## Chapter 1: The Beginning
Our hero starts their journey

## Chapter 2: The Challenge
The first major obstacle appears

## Chapter 3: The Resolution
Everything comes together in the end`

	plan, err := books.PlanFromString(planStr)
	require.NoError(t, err)
	require.Len(t, plan, 3)

	require.Equal(t, "Chapter 1: The Beginning", plan[0].GetTitle())
	require.Equal(t, "Our hero starts their journey", plan[0].GetContent())

	require.Equal(t, "Chapter 2: The Challenge", plan[1].GetTitle())
	require.Equal(t, "The first major obstacle appears", plan[1].GetContent())

	require.Equal(t, "Chapter 3: The Resolution", plan[2].GetTitle())
	require.Equal(t, "Everything comes together in the end", plan[2].GetContent())
}

func TestPlanFromStringEmpty(t *testing.T) {
	plan, err := books.PlanFromString("")
	require.NoError(t, err)
	require.Empty(t, plan)
}

func TestPlanString(t *testing.T) {
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}

	result := plan.String()
	expected := "## 1. Chapter 1\nFirst chapter\n\n## 2. Chapter 2\nSecond chapter"
	require.Equal(t, expected, result)
}

func TestPlanStringEmpty(t *testing.T) {
	plan := books.Plan{}
	result := plan.String()
	require.Empty(t, result)
}

func TestRoundTripPlan(t *testing.T) {
	original := books.Plan{
		books.NewChapter(0, "Prologue", "The story begins"),
		books.NewChapter(1, "Chapter 1", "The adventure starts"),
		books.NewChapter(0, "Epilogue", "The story ends"),
	}

	// Convert to string and back
	str := original.String()
	parsed, err := books.PlanFromString(str)
	require.NoError(t, err)

	// Verify all data is preserved
	require.Len(t, parsed, 3)
	for i, chapter := range original {
		require.Equal(t, chapter.GetTitle(), parsed[i].GetTitle())
		require.Equal(t, chapter.GetContent(), parsed[i].GetContent())
	}
}

func TestPlanMerge(t *testing.T) {
	// Create original plan
	original := books.Plan{
		books.NewChapter(1, "Chapter 1", "Original content 1"),
		books.NewChapter(2, "Chapter 2", "Original content 2"),
		books.NewChapter(0, "Prologue", "Original prologue"),
	}

	// Create new plan with overlapping and new chapters
	updates := books.Plan{
		books.NewChapter(2, "Chapter 2 Updated", "Updated content 2"),
		books.NewChapter(3, "Chapter 3", "New content 3"),
		books.NewChapter(0, "Epilogue", "New epilogue"),
	}

	// Test merge through SetPlan
	fs := afero.NewMemMapFs()
	book, err := books.CreateBook(fs, "/test")
	require.NoError(t, err)

	// Set original plan
	err = book.SetPlan(original)
	require.NoError(t, err)

	// Merge updates
	err = book.SetPlan(updates)
	require.NoError(t, err)

	// Verify merged result
	result := book.GetPlan()

	// Should have 5 chapters: original 1, updated 2, original prologue, new 3, new epilogue
	require.Len(t, result, 5)

	// Find chapters by number for verification
	chaptersByNumber := make(map[int]*books.Chapter)
	var zeroChapters []*books.Chapter

	for _, ch := range result {
		if ch.GetNumber() == 0 {
			zeroChapters = append(zeroChapters, ch)
		} else {
			chaptersByNumber[ch.GetNumber()] = ch
		}
	}

	// Verify numbered chapters
	require.Equal(t, "Chapter 1", chaptersByNumber[1].GetTitle())
	require.Equal(t, "Original content 1", chaptersByNumber[1].GetContent())

	require.Equal(t, "Chapter 2 Updated", chaptersByNumber[2].GetTitle())
	require.Equal(t, "Updated content 2", chaptersByNumber[2].GetContent())

	require.Equal(t, "Chapter 3", chaptersByNumber[3].GetTitle())
	require.Equal(t, "New content 3", chaptersByNumber[3].GetContent())

	// Verify chapters with number 0 (should have both prologue and epilogue)
	require.Len(t, zeroChapters, 2)
	titles := make([]string, len(zeroChapters))
	for i, ch := range zeroChapters {
		titles[i] = ch.GetTitle()
	}
	require.Contains(t, titles, "Prologue")
	require.Contains(t, titles, "Epilogue")
}

func TestPlanMergeEmpty(t *testing.T) {
	fs := afero.NewMemMapFs()
	book, err := books.CreateBook(fs, "/test")
	require.NoError(t, err)

	// Set original plan
	original := books.Plan{
		books.NewChapter(1, "Chapter 1", "Content 1"),
	}
	err = book.SetPlan(original)
	require.NoError(t, err)

	// Merge with empty plan
	empty := books.Plan{}
	err = book.SetPlan(empty)
	require.NoError(t, err)

	// Should still have original plan
	result := book.GetPlan()
	require.Len(t, result, 1)
	require.Equal(t, "Chapter 1", result[0].GetTitle())
}

func TestPlanFromStringWithNumbers(t *testing.T) {
	// Test plan string with numbered chapters
	planStr := `## 1. Chapter 1: The Beginning
Our hero starts their journey

## 2. Chapter 2: The Challenge
The first major obstacle appears

## 3. Chapter 3: The Resolution
Everything comes together in the end`

	plan, err := books.PlanFromString(planStr)
	require.NoError(t, err)
	require.Len(t, plan, 3)

	// Verify numbers and titles are parsed correctly
	require.Equal(t, 1, plan[0].GetNumber())
	require.Equal(t, "Chapter 1: The Beginning", plan[0].GetTitle())
	require.Equal(t, "Our hero starts their journey", plan[0].GetContent())

	require.Equal(t, 2, plan[1].GetNumber())
	require.Equal(t, "Chapter 2: The Challenge", plan[1].GetTitle())
	require.Equal(t, "The first major obstacle appears", plan[1].GetContent())

	require.Equal(t, 3, plan[2].GetNumber())
	require.Equal(t, "Chapter 3: The Resolution", plan[2].GetTitle())
	require.Equal(t, "Everything comes together in the end", plan[2].GetContent())
}

func TestPlanFromStringMixedNumbers(t *testing.T) {
	// Test plan string with mixed numbered and non-numbered chapters
	planStr := `## Prologue
The story begins

## 1. Chapter 1: The Discovery
The first chapter

## Interlude
A break in the action

## 2. Chapter 2: The Conflict
The second chapter`

	plan, err := books.PlanFromString(planStr)
	require.NoError(t, err)
	require.Len(t, plan, 4)

	// Verify mixed numbering
	require.Equal(t, 0, plan[0].GetNumber())
	require.Equal(t, "Prologue", plan[0].GetTitle())

	require.Equal(t, 1, plan[1].GetNumber())
	require.Equal(t, "Chapter 1: The Discovery", plan[1].GetTitle())

	require.Equal(t, 0, plan[2].GetNumber())
	require.Equal(t, "Interlude", plan[2].GetTitle())

	require.Equal(t, 2, plan[3].GetNumber())
	require.Equal(t, "Chapter 2: The Conflict", plan[3].GetTitle())
}

func TestRoundTripNumberedChaptersAndPlanMerge(t *testing.T) {
	fs := afero.NewMemMapFs()
	book, err := books.CreateBook(fs, "/test")
	require.NoError(t, err)

	// Create original plan with numbered chapters
	original := books.Plan{
		books.NewChapter(1, "Chapter 1: The Start", "Beginning of the story"),
		books.NewChapter(2, "Chapter 2: The Journey", "Middle of the story"),
		books.NewChapter(0, "Prologue", "Before it all began"),
	}

	err = book.SetPlan(original)
	require.NoError(t, err)

	// Convert plan to string and back
	planStr := original.String()
	parsedPlan, err := books.PlanFromString(planStr)
	require.NoError(t, err)

	// Verify round-trip preserves numbers
	require.Len(t, parsedPlan, 3)

	// Find chapters by number
	chaptersByNumber := make(map[int]*books.Chapter)
	var zeroChapters []*books.Chapter

	for _, ch := range parsedPlan {
		if ch.GetNumber() == 0 {
			zeroChapters = append(zeroChapters, ch)
		} else {
			chaptersByNumber[ch.GetNumber()] = ch
		}
	}

	require.Equal(t, 1, chaptersByNumber[1].GetNumber())
	require.Equal(t, "Chapter 1: The Start", chaptersByNumber[1].GetTitle())
	require.Equal(t, "Beginning of the story", chaptersByNumber[1].GetContent())

	require.Equal(t, 2, chaptersByNumber[2].GetNumber())
	require.Equal(t, "Chapter 2: The Journey", chaptersByNumber[2].GetTitle())
	require.Equal(t, "Middle of the story", chaptersByNumber[2].GetContent())

	require.Len(t, zeroChapters, 1)
	require.Equal(t, "Prologue", zeroChapters[0].GetTitle())
	require.Equal(t, "Before it all began", zeroChapters[0].GetContent())

	// Now test merge functionality with updates
	updates := books.Plan{
		books.NewChapter(2, "Chapter 2: The Updated Journey", "Updated middle story"),
		books.NewChapter(3, "Chapter 3: The End", "End of the story"),
		books.NewChapter(0, "Epilogue", "After it all ended"),
	}

	// Merge updates
	err = book.SetPlan(updates)
	require.NoError(t, err)

	// Verify merged result maintains all chapters
	merged := book.GetPlan()
	require.Len(t, merged, 5) // 1, updated 2, prologue, new 3, epilogue

	// Reset maps for merged plan
	chaptersByNumber = make(map[int]*books.Chapter)
	zeroChapters = []*books.Chapter{}

	for _, ch := range merged {
		if ch.GetNumber() == 0 {
			zeroChapters = append(zeroChapters, ch)
		} else {
			chaptersByNumber[ch.GetNumber()] = ch
		}
	}

	// Verify original chapter 1 is preserved
	require.Equal(t, "Chapter 1: The Start", chaptersByNumber[1].GetTitle())
	require.Equal(t, "Beginning of the story", chaptersByNumber[1].GetContent())

	// Verify chapter 2 was updated
	require.Equal(t, "Chapter 2: The Updated Journey", chaptersByNumber[2].GetTitle())
	require.Equal(t, "Updated middle story", chaptersByNumber[2].GetContent())

	// Verify new chapter 3 was added
	require.Equal(t, "Chapter 3: The End", chaptersByNumber[3].GetTitle())
	require.Equal(t, "End of the story", chaptersByNumber[3].GetContent())

	// Verify both zero-numbered chapters exist
	require.Len(t, zeroChapters, 2)
	titles := make([]string, len(zeroChapters))
	for i, ch := range zeroChapters {
		titles[i] = ch.GetTitle()
	}
	require.Contains(t, titles, "Prologue")
	require.Contains(t, titles, "Epilogue")

	// Final round-trip test - convert merged plan to string and back
	finalStr := merged.String()
	finalParsed, err := books.PlanFromString(finalStr)
	require.NoError(t, err)
	require.Len(t, finalParsed, 5)

	// Verify the final parsed version has all the right data
	finalByNumber := make(map[int]*books.Chapter)
	var finalZeroChapters []*books.Chapter

	for _, ch := range finalParsed {
		if ch.GetNumber() == 0 {
			finalZeroChapters = append(finalZeroChapters, ch)
		} else {
			finalByNumber[ch.GetNumber()] = ch
		}
	}

	require.Equal(t, "Chapter 1: The Start", finalByNumber[1].GetTitle())
	require.Equal(t, "Chapter 2: The Updated Journey", finalByNumber[2].GetTitle())
	require.Equal(t, "Chapter 3: The End", finalByNumber[3].GetTitle())
	require.Len(t, finalZeroChapters, 2)
}

func TestPlanMergedCopy(t *testing.T) {
	ch1 := books.NewChapter(1, "Chapter 1", "Content 1")
	ch2 := books.NewChapter(2, "Chapter 2", "Content 2")
	ch0 := books.NewChapter(0, "Unnumbered", "No number")

	original := books.Plan{ch1, ch2, ch0}

	newCh2 := books.NewChapter(2, "New Chapter 2", "New Content 2")
	ch3 := books.NewChapter(3, "Chapter 3", "Content 3")
	anotherCh0 := books.NewChapter(0, "Another Unnumbered", "Also no number")

	other := books.Plan{newCh2, ch3, anotherCh0}

	merged := original.MergedCopy(other)

	// Verify original unchanged
	require.Len(t, original, 3)
	require.Equal(t, "Chapter 2", original[1].GetTitle())

	// Verify merged result
	require.Len(t, merged, 5) // 3 original + 2 new (chapter 2 replaced, chapter 3 added, unnumbered added)

	// Chapter 1 should remain unchanged
	require.Equal(t, 1, merged[0].GetNumber())
	require.Equal(t, "Chapter 1", merged[0].GetTitle())
	require.Equal(t, "Content 1", merged[0].GetContent())

	// Chapter 2 should be replaced
	chap2Found := false
	for _, ch := range merged {
		if ch.GetNumber() == 2 {
			require.Equal(t, "New Chapter 2", ch.GetTitle())
			require.Equal(t, "New Content 2", ch.GetContent())
			chap2Found = true
			break
		}
	}
	require.True(t, chap2Found, "Chapter 2 should be found and replaced")

	// Chapter 3 should be added
	chap3Found := false
	for _, ch := range merged {
		if ch.GetNumber() == 3 {
			require.Equal(t, "Chapter 3", ch.GetTitle())
			require.Equal(t, "Content 3", ch.GetContent())
			chap3Found = true
			break
		}
	}
	require.True(t, chap3Found, "Chapter 3 should be added")

	// Both unnumbered chapters should be present
	unnumberedCount := 0
	for _, ch := range merged {
		if ch.GetNumber() == 0 {
			unnumberedCount++
		}
	}
	require.Equal(t, 2, unnumberedCount, "Should have 2 unnumbered chapters")
}

func TestPlanMergedCopyEmpty(t *testing.T) {
	ch1 := books.NewChapter(1, "Chapter 1", "Content 1")
	original := books.Plan{ch1}

	empty := books.Plan{}
	merged := original.MergedCopy(empty)

	// Should be identical to original
	require.Len(t, merged, 1)
	require.Equal(t, original[0].GetNumber(), merged[0].GetNumber())
	require.Equal(t, original[0].GetTitle(), merged[0].GetTitle())
	require.Equal(t, original[0].GetContent(), merged[0].GetContent())
}

func TestPlanMergedCopyOnlyUnnumbered(t *testing.T) {
	ch1 := books.NewChapter(1, "Chapter 1", "Content 1")
	original := books.Plan{ch1}

	unCh1 := books.NewChapter(0, "Unnumbered 1", "No number 1")
	unCh2 := books.NewChapter(0, "Unnumbered 2", "No number 2")
	other := books.Plan{unCh1, unCh2}

	merged := original.MergedCopy(other)

	require.Len(t, merged, 3) // 1 original + 2 unnumbered
	require.Equal(t, 1, merged[0].GetNumber())
	require.Equal(t, 0, merged[1].GetNumber())
	require.Equal(t, 0, merged[2].GetNumber())
	require.Equal(t, "Unnumbered 1", merged[1].GetTitle())
	require.Equal(t, "Unnumbered 2", merged[2].GetTitle())
}

func TestPlanMergedCopyAllReplaced(t *testing.T) {
	oldCh1 := books.NewChapter(1, "Old Chapter 1", "Old Content 1")
	oldCh2 := books.NewChapter(2, "Old Chapter 2", "Old Content 2")
	original := books.Plan{oldCh1, oldCh2}

	newCh1 := books.NewChapter(1, "New Chapter 1", "New Content 1")
	newCh2 := books.NewChapter(2, "New Chapter 2", "New Content 2")
	other := books.Plan{newCh1, newCh2}

	merged := original.MergedCopy(other)

	require.Len(t, merged, 2)
	require.Equal(t, "New Chapter 1", merged[0].GetTitle())
	require.Equal(t, "New Chapter 2", merged[1].GetTitle())
}

func TestPlanMergedCopyEmptyOriginal(t *testing.T) {
	// Test merging into an empty plan
	empty := books.Plan{}

	ch1 := books.NewChapter(1, "Chapter 1", "Content 1")
	ch2 := books.NewChapter(0, "Unnumbered", "No number")
	other := books.Plan{ch1, ch2}

	merged := empty.MergedCopy(other)

	// Should contain all chapters from other
	require.Len(t, merged, 2)
	require.Equal(t, 1, merged[0].GetNumber())
	require.Equal(t, "Chapter 1", merged[0].GetTitle())
	require.Equal(t, "Content 1", merged[0].GetContent())
	require.Equal(t, 0, merged[1].GetNumber())
	require.Equal(t, "Unnumbered", merged[1].GetTitle())
	require.Equal(t, "No number", merged[1].GetContent())
}
