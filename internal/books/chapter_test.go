package books_test

import (
	"testing"
	"viter/internal/books"

	"github.com/stretchr/testify/require"
)

func TestChapterFromString(t *testing.T) {
	chapterStr := `## Chapter 1: The Discovery
Sarah stared at the crime scene, her coffee growing cold in her hands. The victim lay sprawled across the alley, but something was wrong. There was no blood, despite the obvious wounds.

"This doesn't make sense," she muttered to herself.`

	chapter, err := books.ChapterFromString(chapterStr)
	require.NoError(t, err)
	require.Equal(t, "Chapter 1: The Discovery", chapter.GetTitle())
	require.Contains(t, chapter.GetContent(), "Sarah stared at the crime scene")
	require.Contains(t, chapter.GetContent(), "This doesn't make sense")
}

func TestChapterFromStringEmpty(t *testing.T) {
	_, err := books.ChapterFromString("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty chapter string")
}

func TestChapterFromStringNoTitle(t *testing.T) {
	chapterStr := `This is just content without a title`
	_, err := books.ChapterFromString(chapterStr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no title found")
}

func TestChapterString(t *testing.T) {
	chapter := books.NewChapter(5, "Chapter 5: The Revelation", "Everything became clear in that moment. The pieces of the puzzle finally fit together.")

	result := chapter.String()
	require.Equal(t, "## 5. Chapter 5: The Revelation\nEverything became clear in that moment. The pieces of the puzzle finally fit together.", result)
}

func TestRoundTripChapter(t *testing.T) {
	original := books.NewChapter(42, "Chapter 42: The Answer", "The answer to life, the universe, and everything is 42.")

	// Convert to string and back
	str := original.String()
	parsed, err := books.ChapterFromString(str)
	require.NoError(t, err)

	// Verify all data is preserved
	require.Equal(t, original.GetTitle(), parsed.GetTitle())
	require.Equal(t, original.GetContent(), parsed.GetContent())
	require.Equal(t, original.GetNumber(), parsed.GetNumber())
}

func TestChapterGetNumber(t *testing.T) {
	chapter := books.NewChapter(42, "Test Chapter", "Test content")
	require.Equal(t, 42, chapter.GetNumber())

	// Test chapter with number 0
	chapterZero := books.NewChapter(0, "Prologue", "Beginning")
	require.Equal(t, 0, chapterZero.GetNumber())
}

func TestChapterFromStringWithNumber(t *testing.T) {
	// Test parsing chapter with number
	chapterStr := `## 5. The Final Confrontation
The hero faces the ultimate challenge.`

	chapter, err := books.ChapterFromString(chapterStr)
	require.NoError(t, err)
	require.Equal(t, 5, chapter.GetNumber())
	require.Equal(t, "The Final Confrontation", chapter.GetTitle())
	require.Equal(t, "The hero faces the ultimate challenge.", chapter.GetContent())
}

func TestChapterFromStringWithoutNumber(t *testing.T) {
	// Test parsing chapter without number format
	chapterStr := `## Epilogue
The story concludes.`

	chapter, err := books.ChapterFromString(chapterStr)
	require.NoError(t, err)
	require.Equal(t, 0, chapter.GetNumber())
	require.Equal(t, "Epilogue", chapter.GetTitle())
	require.Equal(t, "The story concludes.", chapter.GetContent())
}

func TestChapterFromStringInvalidNumber(t *testing.T) {
	// Test parsing chapter with invalid number format
	chapterStr := `## abc. Invalid Number
This should not parse the number.`

	chapter, err := books.ChapterFromString(chapterStr)
	require.NoError(t, err)
	require.Equal(t, 0, chapter.GetNumber())
	require.Equal(t, "abc. Invalid Number", chapter.GetTitle())
	require.Equal(t, "This should not parse the number.", chapter.GetContent())
}
