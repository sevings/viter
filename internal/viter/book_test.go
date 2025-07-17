package viter_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"viter/internal/viter"
)

func TestCreateBook(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)
	require.NotNil(t, book)

	// Check that meta.md file was created
	exists, err := afero.Exists(fs, filepath.Join(path, "meta.md"))
	require.NoError(t, err)
	require.True(t, exists)
}

func TestLoadBook(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create directory and files
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)

	// Create meta.md
	metaContent := `## Style
Fantasy Adventure

## Genres
Fantasy, Adventure, Coming of Age

## World
A magical realm where dragons still roam

## Main Characters
### Aria
A young mage discovering her powers

### Gareth
A skilled warrior with a mysterious past

## Minor Characters
### Elder Thorne
The wise village elder

## Plot
A young mage must save her village from an ancient evil`

	err = afero.WriteFile(fs, filepath.Join(path, "meta.md"), []byte(metaContent), 0644)
	require.NoError(t, err)

	// Create plan.md
	planContent := `## Chapter 1: The Awakening
Aria discovers her magical abilities when her village is attacked

## Chapter 2: The Journey Begins
Aria and Gareth set out to find the source of the evil`

	err = afero.WriteFile(fs, filepath.Join(path, "plan.md"), []byte(planContent), 0644)
	require.NoError(t, err)

	// Load the book
	book, err := viter.LoadBook(fs, path)
	require.NoError(t, err)
	require.NotNil(t, book)

	// Verify metadata was loaded
	meta := book.GetMeta()
	require.Equal(t, "Fantasy Adventure", meta.GetStyle())
	require.Equal(t, []string{"Fantasy", "Adventure", "Coming of Age"}, meta.GetGenres())
	require.Equal(t, "A magical realm where dragons still roam", meta.GetWorld())
	require.Len(t, meta.GetMainCharacters(), 2)
	require.Equal(t, "Aria", meta.GetMainCharacters()[0].GetName())
	require.Equal(t, "A young mage discovering her powers", meta.GetMainCharacters()[0].GetDesc())

	// Verify plan was loaded
	plan := book.GetPlan()
	require.Len(t, plan, 2)

	chapter1, err := book.GetPlanChapter(0)
	require.NoError(t, err)
	require.Equal(t, "Chapter 1: The Awakening", chapter1.GetTitle())
	require.Equal(t, "Aria discovers her magical abilities when her village is attacked", chapter1.GetContent())
}

func TestLoadBookNonExistent(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/nonexistent/book"

	book, err := viter.LoadBook(fs, path)
	require.NoError(t, err)
	require.NotNil(t, book)

	// Should have empty metadata and plan
	meta := book.GetMeta()
	require.Empty(t, meta.GetStyle())
	require.Empty(t, meta.GetGenres())

	plan := book.GetPlan()
	require.Empty(t, plan)
}

func TestBookSave(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set metadata
	meta := viter.BookMeta{}
	meta.SetStyle("Science Fiction")
	meta.SetGenres([]string{"Sci-Fi", "Thriller"})
	meta.SetWorld("A dystopian future")
	meta.SetMainCharacters([]viter.Character{
		viter.NewCharacter("Alex", "A rebel hacker"),
	})
	meta.SetPlot("The fight against a totalitarian regime")

	err = book.SetMeta(meta)
	require.NoError(t, err)

	// Set plan
	plan := viter.Plan{
		viter.NewChapter("Chapter 1: The Resistance", "Alex joins the underground"),
		viter.NewChapter("Chapter 2: The Mission", "The first strike against the system"),
	}

	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Save the book
	err = book.Save()
	require.NoError(t, err)

	// Verify files were created
	metaExists, err := afero.Exists(fs, filepath.Join(path, "meta.md"))
	require.NoError(t, err)
	require.True(t, metaExists)

	planExists, err := afero.Exists(fs, filepath.Join(path, "plan.md"))
	require.NoError(t, err)
	require.True(t, planExists)

	// Verify content
	metaContent, err := afero.ReadFile(fs, filepath.Join(path, "meta.md"))
	require.NoError(t, err)
	require.Contains(t, string(metaContent), "Science Fiction")
	require.Contains(t, string(metaContent), "Sci-Fi, Thriller")

	planContent, err := afero.ReadFile(fs, filepath.Join(path, "plan.md"))
	require.NoError(t, err)
	require.Contains(t, string(planContent), "Chapter 1: The Resistance")
	require.Contains(t, string(planContent), "Alex joins the underground")
}

func TestSaveChapter(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan first
	plan := viter.Plan{
		viter.NewChapter("Chapter 1", "First chapter"),
		viter.NewChapter("Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Save a chapter
	chapter := viter.NewChapter("Chapter 1: The Beginning", "It was a dark and stormy night...")
	err = book.SaveChapter(0, chapter)
	require.NoError(t, err)

	// Verify file was created
	chapterPath := filepath.Join(path, "chapter_0.md")
	exists, err := afero.Exists(fs, chapterPath)
	require.NoError(t, err)
	require.True(t, exists)

	// Verify content
	content, err := afero.ReadFile(fs, chapterPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "## Chapter 1: The Beginning")
	require.Contains(t, string(content), "It was a dark and stormy night...")
}

func TestSaveChapterInvalidIndex(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := viter.Plan{
		viter.NewChapter("Chapter 1", "First chapter"),
		viter.NewChapter("Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	chapter := viter.NewChapter("Invalid Chapter", "This shouldn't work")

	// Test negative index
	err = book.SaveChapter(-1, chapter)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	// Test index too high
	err = book.SaveChapter(2, chapter)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)
}

func TestGetPlanChapter(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan
	plan := viter.Plan{
		viter.NewChapter("Chapter 1", "First chapter content"),
		viter.NewChapter("Chapter 2", "Second chapter content"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Test valid indices
	chapter0, err := book.GetPlanChapter(0)
	require.NoError(t, err)
	require.Equal(t, "Chapter 1", chapter0.GetTitle())
	require.Equal(t, "First chapter content", chapter0.GetContent())

	chapter1, err := book.GetPlanChapter(1)
	require.NoError(t, err)
	require.Equal(t, "Chapter 2", chapter1.GetTitle())
	require.Equal(t, "Second chapter content", chapter1.GetContent())

	// Test invalid indices
	_, err = book.GetPlanChapter(-1)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	_, err = book.GetPlanChapter(2)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)
}

func TestMetaFromString(t *testing.T) {
	metaStr := `## Style
Epic Fantasy

## Genres
Fantasy, Adventure, Magic

## World
A world where magic flows through ancient ley lines

## Main Characters
### Lyra
A young sorceress with untapped potential

### Thane
A gruff dwarf warrior with a heart of gold

## Minor Characters
### Wizard Aldric
The mentor figure who guides our heroes

### Queen Morwyn
The ruler of the northern kingdom

## Plot
An ancient evil stirs, threatening to destroy the delicate balance of magic in the world`

	meta, err := viter.MetaFromString(metaStr)
	require.NoError(t, err)

	require.Equal(t, "Epic Fantasy", meta.GetStyle())
	require.Equal(t, []string{"Fantasy", "Adventure", "Magic"}, meta.GetGenres())
	require.Equal(t, "A world where magic flows through ancient ley lines", meta.GetWorld())

	mainChars := meta.GetMainCharacters()
	require.Len(t, mainChars, 2)
	require.Equal(t, "Lyra", mainChars[0].GetName())
	require.Equal(t, "A young sorceress with untapped potential", mainChars[0].GetDesc())
	require.Equal(t, "Thane", mainChars[1].GetName())
	require.Equal(t, "A gruff dwarf warrior with a heart of gold", mainChars[1].GetDesc())

	minorChars := meta.GetMinorCharacters()
	require.Len(t, minorChars, 2)
	require.Equal(t, "Wizard Aldric", minorChars[0].GetName())
	require.Equal(t, "The mentor figure who guides our heroes", minorChars[0].GetDesc())

	require.Equal(t, "An ancient evil stirs, threatening to destroy the delicate balance of magic in the world", meta.GetPlot())
}

func TestMetaString(t *testing.T) {
	meta := viter.BookMeta{}
	meta.SetStyle("Urban Fantasy")
	meta.SetGenres([]string{"Fantasy", "Mystery", "Urban"})
	meta.SetWorld("Modern city with hidden supernatural elements")
	meta.SetMainCharacters([]viter.Character{
		viter.NewCharacter("Detective Sarah", "A cop who discovers the supernatural"),
		viter.NewCharacter("Marcus", "A vampire trying to solve his own murder"),
	})
	meta.SetMinorCharacters([]viter.Character{
		viter.NewCharacter("Chief Williams", "Sarah's skeptical boss"),
	})
	meta.SetPlot("A detective and vampire must work together to solve supernatural crimes")

	result := meta.String()
	require.Contains(t, result, "## Style")
	require.Contains(t, result, "Urban Fantasy")
	require.Contains(t, result, "## Genres")
	require.Contains(t, result, "Fantasy, Mystery, Urban")
	require.Contains(t, result, "## World")
	require.Contains(t, result, "Modern city with hidden supernatural elements")
	require.Contains(t, result, "## Main Characters")
	require.Contains(t, result, "### Detective Sarah")
	require.Contains(t, result, "A cop who discovers the supernatural")
	require.Contains(t, result, "## Minor Characters")
	require.Contains(t, result, "### Chief Williams")
	require.Contains(t, result, "## Plot")
	require.Contains(t, result, "A detective and vampire must work together to solve supernatural crimes")
}

func TestChapterFromString(t *testing.T) {
	chapterStr := `## Chapter 1: The Discovery
Sarah stared at the crime scene, her coffee growing cold in her hands. The victim lay sprawled across the alley, but something was wrong. There was no blood, despite the obvious wounds.

"This doesn't make sense," she muttered to herself.`

	chapter, err := viter.ChapterFromString(chapterStr)
	require.NoError(t, err)
	require.Equal(t, "Chapter 1: The Discovery", chapter.GetTitle())
	require.Contains(t, chapter.GetContent(), "Sarah stared at the crime scene")
	require.Contains(t, chapter.GetContent(), "This doesn't make sense")
}

func TestChapterFromStringEmpty(t *testing.T) {
	_, err := viter.ChapterFromString("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty chapter string")
}

func TestChapterFromStringNoTitle(t *testing.T) {
	chapterStr := `This is just content without a title`
	_, err := viter.ChapterFromString(chapterStr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no title found")
}

func TestChapterString(t *testing.T) {
	chapter := viter.NewChapter("Chapter 5: The Revelation", "Everything became clear in that moment. The pieces of the puzzle finally fit together.")

	result := chapter.String()
	require.Equal(t, "## Chapter 5: The Revelation\nEverything became clear in that moment. The pieces of the puzzle finally fit together.", result)
}

func TestPlanFromString(t *testing.T) {
	planStr := `## Chapter 1: The Beginning
Our hero starts their journey

## Chapter 2: The Challenge
The first major obstacle appears

## Chapter 3: The Resolution
Everything comes together in the end`

	plan, err := viter.PlanFromString(planStr)
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
	plan, err := viter.PlanFromString("")
	require.NoError(t, err)
	require.Empty(t, plan)
}

func TestPlanString(t *testing.T) {
	plan := viter.Plan{
		viter.NewChapter("Chapter 1", "First chapter"),
		viter.NewChapter("Chapter 2", "Second chapter"),
	}

	result := plan.String()
	expected := "## Chapter 1\nFirst chapter\n\n## Chapter 2\nSecond chapter"
	require.Equal(t, expected, result)
}

func TestPlanStringEmpty(t *testing.T) {
	plan := viter.Plan{}
	result := plan.String()
	require.Empty(t, result)
}

func TestRoundTripMetadata(t *testing.T) {
	// Create original metadata
	original := viter.BookMeta{}
	original.SetStyle("Space Opera")
	original.SetGenres([]string{"Science Fiction", "Adventure"})
	original.SetWorld("A galaxy far, far away")
	original.SetMainCharacters([]viter.Character{
		viter.NewCharacter("Captain Nova", "A fearless space explorer"),
	})
	original.SetPlot("The quest to save the galaxy")

	// Convert to string and back
	str := original.String()
	parsed, err := viter.MetaFromString(str)
	require.NoError(t, err)

	// Verify all data is preserved
	require.Equal(t, original.GetStyle(), parsed.GetStyle())
	require.Equal(t, original.GetGenres(), parsed.GetGenres())
	require.Equal(t, original.GetWorld(), parsed.GetWorld())
	require.Equal(t, original.GetPlot(), parsed.GetPlot())
	require.Len(t, parsed.GetMainCharacters(), 1)
	require.Equal(t, "Captain Nova", parsed.GetMainCharacters()[0].GetName())
}

func TestRoundTripChapter(t *testing.T) {
	original := viter.NewChapter("Chapter 42: The Answer", "The answer to life, the universe, and everything is 42.")

	// Convert to string and back
	str := original.String()
	parsed, err := viter.ChapterFromString(str)
	require.NoError(t, err)

	// Verify all data is preserved
	require.Equal(t, original.GetTitle(), parsed.GetTitle())
	require.Equal(t, original.GetContent(), parsed.GetContent())
}

func TestRoundTripPlan(t *testing.T) {
	original := viter.Plan{
		viter.NewChapter("Prologue", "The story begins"),
		viter.NewChapter("Chapter 1", "The adventure starts"),
		viter.NewChapter("Epilogue", "The story ends"),
	}

	// Convert to string and back
	str := original.String()
	parsed, err := viter.PlanFromString(str)
	require.NoError(t, err)

	// Verify all data is preserved
	require.Len(t, parsed, 3)
	for i, chapter := range original {
		require.Equal(t, chapter.GetTitle(), parsed[i].GetTitle())
		require.Equal(t, chapter.GetContent(), parsed[i].GetContent())
	}
}

func TestIntegrationSaveAndLoad(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/integration"

	// Create and configure a book
	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up complete metadata
	meta := viter.BookMeta{}
	meta.SetStyle("Horror")
	meta.SetGenres([]string{"Horror", "Thriller", "Supernatural"})
	meta.SetWorld("A small town with dark secrets")
	meta.SetMainCharacters([]viter.Character{
		viter.NewCharacter("Dr. Emma Carter", "A psychiatrist who uncovers the truth"),
		viter.NewCharacter("Father Miguel", "A priest battling ancient evils"),
	})
	meta.SetMinorCharacters([]viter.Character{
		viter.NewCharacter("Sheriff Brooks", "The local law enforcement"),
	})
	meta.SetPlot("A town's buried secrets come back to haunt the living")

	err = book.SetMeta(meta)
	require.NoError(t, err)

	// Set up a plan
	plan := viter.Plan{
		viter.NewChapter("Chapter 1: Arrival", "Dr. Carter arrives in the small town"),
		viter.NewChapter("Chapter 2: Strange Occurrences", "Mysterious events begin to unfold"),
		viter.NewChapter("Chapter 3: The Truth", "The dark history is revealed"),
	}

	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Save everything
	err = book.Save()
	require.NoError(t, err)

	// Save individual chapters
	fullChapter1 := viter.NewChapter("Chapter 1: Arrival - The Beginning", "Dr. Emma Carter stepped off the bus into the dusty main street of Millbrook. The town seemed ordinary enough, but something in the air made her skin crawl.")
	err = book.SaveChapter(0, fullChapter1)
	require.NoError(t, err)

	// Load the book fresh
	loadedBook, err := viter.LoadBook(fs, path)
	require.NoError(t, err)

	// Verify metadata was preserved
	loadedMeta := loadedBook.GetMeta()
	require.Equal(t, "Horror", loadedMeta.GetStyle())
	require.Equal(t, []string{"Horror", "Thriller", "Supernatural"}, loadedMeta.GetGenres())
	require.Len(t, loadedMeta.GetMainCharacters(), 2)
	require.Equal(t, "Dr. Emma Carter", loadedMeta.GetMainCharacters()[0].GetName())

	// Verify plan was preserved
	loadedPlan := loadedBook.GetPlan()
	require.Len(t, loadedPlan, 3)
	require.Equal(t, "Chapter 1: Arrival", loadedPlan[0].GetTitle())
	require.Equal(t, "Dr. Carter arrives in the small town", loadedPlan[0].GetContent())

	// Verify individual chapter file was created
	chapterExists, err := afero.Exists(fs, filepath.Join(path, "chapter_0.md"))
	require.NoError(t, err)
	require.True(t, chapterExists)

	// Verify chapter content
	chapterContent, err := afero.ReadFile(fs, filepath.Join(path, "chapter_0.md"))
	require.NoError(t, err)
	require.Contains(t, string(chapterContent), "Chapter 1: Arrival - The Beginning")
	require.Contains(t, string(chapterContent), "Dr. Emma Carter stepped off the bus")
}
