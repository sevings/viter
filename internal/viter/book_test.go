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

## Logline
A young mage must discover her true power to save her village from an ancient evil.

## World
A magical realm where dragons still roam

## Protagonists
### Aria
A young mage discovering her powers

### Gareth
A skilled warrior with a mysterious past

## Antagonists
### Malachar
An ancient sorcerer seeking to drain all magic from the world

## Minor Characters
### Elder Thorne
The wise village elder

## Plot
A young mage must save her village from an ancient evil

## Title
The Mage's Quest`

	err = afero.WriteFile(fs, filepath.Join(path, "meta.md"), []byte(metaContent), 0644)
	require.NoError(t, err)

	// Create plan.md
	planContent := `## Chapter 1: The Awakening
Aria discovers her magical abilities when her village is attacked

## Chapter 2: The Journey Begins
Aria and Gareth set out to find the source of the evil`

	err = afero.WriteFile(fs, filepath.Join(path, "plan.md"), []byte(planContent), 0644)
	require.NoError(t, err)

	// Create meta_critique.md
	metaCritContent := `## Strengths
Strong fantasy world-building

## Improvements
Character development could be enhanced

## Impressions
The book's metadata shows a solid foundation for a fantasy adventure story.

## Score
8`
	err = afero.WriteFile(fs, filepath.Join(path, "meta_critique.md"), []byte(metaCritContent), 0644)
	require.NoError(t, err)

	// Create plan_critique.md
	planCritContent := `## Strengths
Good story structure

## Improvements
More detailed chapter outlines

## Impressions
The chapter plan provides a good structure for the story arc.

## Score
7`
	err = afero.WriteFile(fs, filepath.Join(path, "plan_critique.md"), []byte(planCritContent), 0644)
	require.NoError(t, err)

	// Load the book
	book, err := viter.LoadBook(fs, path)
	require.NoError(t, err)
	require.NotNil(t, book)

	// Verify metadata was loaded
	meta := book.GetMeta()
	require.Equal(t, "The Mage's Quest", meta.GetTitle())
	require.Equal(t, "Fantasy Adventure", meta.GetStyle())
	require.Equal(t, []string{"Fantasy", "Adventure", "Coming of Age"}, meta.GetGenres())
	require.Equal(t, "A young mage must discover her true power to save her village from an ancient evil.", meta.GetLogline())
	require.Equal(t, "A magical realm where dragons still roam", meta.GetWorld())
	require.Len(t, meta.GetProtagonists(), 2)
	require.Equal(t, "Aria", meta.GetProtagonists()[0].GetName())
	require.Equal(t, "A young mage discovering her powers", meta.GetProtagonists()[0].GetDesc())

	// Verify antagonist was loaded
	antagonists := meta.GetAntagonists()
	require.Len(t, antagonists, 1)
	require.Equal(t, "Malachar", antagonists[0].GetName())
	require.Equal(t, "An ancient sorcerer seeking to drain all magic from the world", antagonists[0].GetDesc())

	// Verify plan was loaded
	plan := book.GetPlan()
	require.Len(t, plan, 2)

	chapter1, err := book.GetPlanChapter(1)
	require.NoError(t, err)
	require.Equal(t, "Chapter 1: The Awakening", chapter1.GetTitle())
	require.Equal(t, "Aria discovers her magical abilities when her village is attacked", chapter1.GetContent())

	// Verify critiques were loaded
	metaCrit := book.GetMetaCrit()
	require.NotNil(t, metaCrit)
	require.Equal(t, "Strong fantasy world-building", metaCrit.GetStrengths())
	require.Equal(t, "Character development could be enhanced", metaCrit.GetImprovements())
	require.Equal(t, "The book's metadata shows a solid foundation for a fantasy adventure story.", metaCrit.GetImpressions())
	require.Equal(t, 8, metaCrit.GetScore())

	planCrit := book.GetPlanCrit()
	require.NotNil(t, planCrit)
	require.Equal(t, "Good story structure", planCrit.GetStrengths())
	require.Equal(t, "More detailed chapter outlines", planCrit.GetImprovements())
	require.Equal(t, "The chapter plan provides a good structure for the story arc.", planCrit.GetImpressions())
	require.Equal(t, 7, planCrit.GetScore())
}

func TestLoadBookNonExistent(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/nonexistent/book"

	book, err := viter.LoadBook(fs, path)
	require.Error(t, err)
	require.Equal(t, viter.ErrNoMetaFile, err)
	require.Nil(t, book)
}

func TestBookSave(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set metadata
	meta := &viter.BookMeta{}
	meta.SetStyle("Science Fiction")
	meta.SetGenres([]string{"Sci-Fi", "Thriller"})
	meta.SetWorld("A dystopian future")
	meta.SetProtagonists([]viter.Character{
		viter.NewCharacter("Alex", "A rebel hacker"),
	})
	meta.SetPlot("The fight against a totalitarian regime")

	err = book.SetMeta(meta)
	require.NoError(t, err)

	// Set plan
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1: The Resistance", "Alex joins the underground"),
		viter.NewChapter(2, "Chapter 2: The Mission", "The first strike against the system"),
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
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Save a chapter
	chapter := viter.NewChapter(1, "Chapter 1: The Beginning", "It was a dark and stormy night...")
	err = book.SetChapter(1, chapter)
	require.NoError(t, err)

	// Verify file was created
	chapterPath := filepath.Join(path, "chapter_1.md")
	exists, err := afero.Exists(fs, chapterPath)
	require.NoError(t, err)
	require.True(t, exists)

	// Verify content
	content, err := afero.ReadFile(fs, chapterPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "## 1. Chapter 1: The Beginning")
	require.Contains(t, string(content), "It was a dark and stormy night...")
}

func TestSaveChapterInvalidIndex(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	chapter := viter.NewChapter(0, "Invalid Chapter", "This shouldn't work")

	// Test negative index
	err = book.SetChapter(-1, chapter)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	// Test index too high
	err = book.SetChapter(3, chapter)
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
		viter.NewChapter(1, "Chapter 1", "First chapter content"),
		viter.NewChapter(2, "Chapter 2", "Second chapter content"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Test valid indices
	chapter0, err := book.GetPlanChapter(1)
	require.NoError(t, err)
	require.Equal(t, "Chapter 1", chapter0.GetTitle())
	require.Equal(t, "First chapter content", chapter0.GetContent())

	chapter1, err := book.GetPlanChapter(2)
	require.NoError(t, err)
	require.Equal(t, "Chapter 2", chapter1.GetTitle())
	require.Equal(t, "Second chapter content", chapter1.GetContent())

	// Test invalid indices
	_, err = book.GetPlanChapter(0)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	_, err = book.GetPlanChapter(3)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)
}

func TestMetaFromString(t *testing.T) {
	metaStr := `## Style
Epic Fantasy

## Genres
Fantasy, Adventure, Magic

## Logline
A young sorceress must master ancient magic to prevent the world's destruction.

## World
A world where magic flows through ancient ley lines

## Protagonists
### Lyra
A young sorceress with untapped potential

### Thane
A gruff dwarf warrior with a heart of gold

## Antagonists
### Lord Shadowbane
An ancient necromancer seeking to corrupt all magic

## Minor Characters
### Wizard Aldric
The mentor figure who guides our heroes

### Queen Morwyn
The ruler of the northern kingdom

## Plot
An ancient evil stirs, threatening to destroy the delicate balance of magic in the world

## Title
The Chronicles of Lyra`

	meta, err := viter.MetaFromString(metaStr)
	require.NoError(t, err)

	require.Equal(t, "The Chronicles of Lyra", meta.GetTitle())
	require.Equal(t, "Epic Fantasy", meta.GetStyle())
	require.Equal(t, []string{"Fantasy", "Adventure", "Magic"}, meta.GetGenres())
	require.Equal(t, "A young sorceress must master ancient magic to prevent the world's destruction.", meta.GetLogline())
	require.Equal(t, "A world where magic flows through ancient ley lines", meta.GetWorld())

	mainChars := meta.GetProtagonists()
	require.Len(t, mainChars, 2)
	require.Equal(t, "Lyra", mainChars[0].GetName())
	require.Equal(t, "A young sorceress with untapped potential", mainChars[0].GetDesc())
	require.Equal(t, "Thane", mainChars[1].GetName())
	require.Equal(t, "A gruff dwarf warrior with a heart of gold", mainChars[1].GetDesc())

	minorChars := meta.GetMinorCharacters()
	require.Len(t, minorChars, 2)
	require.Equal(t, "Wizard Aldric", minorChars[0].GetName())
	require.Equal(t, "The mentor figure who guides our heroes", minorChars[0].GetDesc())

	// Verify antagonist was loaded
	antagonists := meta.GetAntagonists()
	require.Len(t, antagonists, 1)
	require.Equal(t, "Lord Shadowbane", antagonists[0].GetName())
	require.Equal(t, "An ancient necromancer seeking to corrupt all magic", antagonists[0].GetDesc())

	require.Equal(t, "An ancient evil stirs, threatening to destroy the delicate balance of magic in the world", meta.GetPlot())
}

func TestMetaString(t *testing.T) {
	meta := viter.BookMeta{}
	meta.SetTitle("Blood and Badges")
	meta.SetStyle("Urban Fantasy")
	meta.SetGenres([]string{"Fantasy", "Mystery", "Urban"})
	meta.SetLogline("When the supernatural meets police procedure, unlikely alliances form.")
	meta.SetWorld("Modern city with hidden supernatural elements")
	meta.SetProtagonists([]viter.Character{
		viter.NewCharacter("Detective Sarah", "A cop who discovers the supernatural"),
		viter.NewCharacter("Marcus", "A vampire trying to solve his own murder"),
	})
	meta.SetMinorCharacters([]viter.Character{
		viter.NewCharacter("Chief Williams", "Sarah's skeptical boss"),
	})
	meta.SetAntagonists([]viter.Character{
		viter.NewCharacter("The Syndicate Leader", "A powerful vampire controlling the city's underworld"),
	})
	meta.SetPlot("A detective and vampire must work together to solve supernatural crimes")

	result := meta.String()
	require.Contains(t, result, "## Title")
	require.Contains(t, result, "Blood and Badges")
	require.Contains(t, result, "## Style")
	require.Contains(t, result, "Urban Fantasy")
	require.Contains(t, result, "## Genres")
	require.Contains(t, result, "Fantasy, Mystery, Urban")
	require.Contains(t, result, "## Logline")
	require.Contains(t, result, "When the supernatural meets police procedure, unlikely alliances form.")
	require.Contains(t, result, "## World")
	require.Contains(t, result, "Modern city with hidden supernatural elements")
	require.Contains(t, result, "## Protagonists")
	require.Contains(t, result, "### Detective Sarah")
	require.Contains(t, result, "A cop who discovers the supernatural")
	require.Contains(t, result, "## Antagonists")
	require.Contains(t, result, "### The Syndicate Leader")
	require.Contains(t, result, "A powerful vampire controlling the city's underworld")
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
	chapter := viter.NewChapter(5, "Chapter 5: The Revelation", "Everything became clear in that moment. The pieces of the puzzle finally fit together.")

	result := chapter.String()
	require.Equal(t, "## 5. Chapter 5: The Revelation\nEverything became clear in that moment. The pieces of the puzzle finally fit together.", result)
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
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}

	result := plan.String()
	expected := "## 1. Chapter 1\nFirst chapter\n\n## 2. Chapter 2\nSecond chapter"
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
	original.SetProtagonists([]viter.Character{
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
	require.Len(t, parsed.GetProtagonists(), 1)
	require.Equal(t, "Captain Nova", parsed.GetProtagonists()[0].GetName())
}

func TestRoundTripChapter(t *testing.T) {
	original := viter.NewChapter(42, "Chapter 42: The Answer", "The answer to life, the universe, and everything is 42.")

	// Convert to string and back
	str := original.String()
	parsed, err := viter.ChapterFromString(str)
	require.NoError(t, err)

	// Verify all data is preserved
	require.Equal(t, original.GetTitle(), parsed.GetTitle())
	require.Equal(t, original.GetContent(), parsed.GetContent())
	require.Equal(t, original.GetNumber(), parsed.GetNumber())
}

func TestRoundTripPlan(t *testing.T) {
	original := viter.Plan{
		viter.NewChapter(0, "Prologue", "The story begins"),
		viter.NewChapter(1, "Chapter 1", "The adventure starts"),
		viter.NewChapter(0, "Epilogue", "The story ends"),
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
	meta := &viter.BookMeta{}
	meta.SetTitle("Shadows of Millbrook")
	meta.SetStyle("Horror")
	meta.SetGenres([]string{"Horror", "Thriller", "Supernatural"})
	meta.SetLogline("In a town where the dead don't rest, the living must face their darkest fears.")
	meta.SetWorld("A small town with dark secrets")
	meta.SetProtagonists([]viter.Character{
		viter.NewCharacter("Dr. Emma Carter", "A psychiatrist who uncovers the truth"),
		viter.NewCharacter("Father Miguel", "A priest battling ancient evils"),
	})
	meta.SetMinorCharacters([]viter.Character{
		viter.NewCharacter("Sheriff Brooks", "The local law enforcement"),
	})
	meta.SetAntagonists([]viter.Character{
		viter.NewCharacter("The Hollow Man", "An ancient spirit seeking revenge"),
	})
	meta.SetPlot("A town's buried secrets come back to haunt the living")

	err = book.SetMeta(meta)
	require.NoError(t, err)

	// Set up a plan
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1: Arrival", "Dr. Carter arrives in the small town"),
		viter.NewChapter(2, "Chapter 2: Strange Occurrences", "Mysterious events begin to unfold"),
		viter.NewChapter(3, "Chapter 3: The Truth", "The dark history is revealed"),
	}

	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Set up critiques
	metaCrit := &viter.Critique{
		Strengths:    "Well-developed characters",
		Improvements: "Need more horror elements",
		Impressions:  "The horror elements are well-balanced with character development.",
		Score:        8,
	}
	err = book.SetMetaCrit(metaCrit)
	require.NoError(t, err)

	planCrit := &viter.Critique{
		Strengths:    "Good pacing",
		Improvements: "More detailed chapter outlines",
		Impressions:  "The three-act structure provides good pacing for building tension.",
		Score:        7,
	}
	err = book.SetPlanCrit(planCrit)
	require.NoError(t, err)

	// Save everything
	err = book.Save()
	require.NoError(t, err)

	// Save individual chapters
	fullChapter1 := viter.NewChapter(1, "Chapter 1: Arrival - The Beginning", "Dr. Emma Carter stepped off the bus into the dusty main street of Millbrook. The town seemed ordinary enough, but something in the air made her skin crawl.")
	err = book.SetChapter(1, fullChapter1)
	require.NoError(t, err)

	// Load the book fresh
	loadedBook, err := viter.LoadBook(fs, path)
	require.NoError(t, err)

	// Verify metadata was preserved
	loadedMeta := loadedBook.GetMeta()
	require.Equal(t, "Shadows of Millbrook", loadedMeta.GetTitle())
	require.Equal(t, "Horror", loadedMeta.GetStyle())
	require.Equal(t, []string{"Horror", "Thriller", "Supernatural"}, loadedMeta.GetGenres())
	require.Equal(t, "In a town where the dead don't rest, the living must face their darkest fears.", loadedMeta.GetLogline())
	require.Len(t, loadedMeta.GetProtagonists(), 2)
	require.Equal(t, "Dr. Emma Carter", loadedMeta.GetProtagonists()[0].GetName())

	// Verify antagonist was preserved
	antagonists := loadedMeta.GetAntagonists()
	require.Len(t, antagonists, 1)
	require.Equal(t, "The Hollow Man", antagonists[0].GetName())
	require.Equal(t, "An ancient spirit seeking revenge", antagonists[0].GetDesc())

	// Verify plan was preserved
	loadedPlan := loadedBook.GetPlan()
	require.Len(t, loadedPlan, 3)
	require.Equal(t, "Chapter 1: Arrival", loadedPlan[0].GetTitle())
	require.Equal(t, "Dr. Carter arrives in the small town", loadedPlan[0].GetContent())

	// Verify individual chapter file was created
	chapterExists, err := afero.Exists(fs, filepath.Join(path, "chapter_1.md"))
	require.NoError(t, err)
	require.True(t, chapterExists)

	// Verify chapter content
	chapterContent, err := afero.ReadFile(fs, filepath.Join(path, "chapter_1.md"))
	require.NoError(t, err)
	require.Contains(t, string(chapterContent), "Chapter 1: Arrival - The Beginning")
	require.Contains(t, string(chapterContent), "Dr. Emma Carter stepped off the bus")

	// Verify critiques were preserved
	loadedMetaCrit := loadedBook.GetMetaCrit()
	require.NotNil(t, loadedMetaCrit)
	require.Equal(t, metaCrit.GetStrengths(), loadedMetaCrit.GetStrengths())
	require.Equal(t, metaCrit.GetImprovements(), loadedMetaCrit.GetImprovements())
	require.Equal(t, metaCrit.GetImpressions(), loadedMetaCrit.GetImpressions())
	require.Equal(t, metaCrit.GetScore(), loadedMetaCrit.GetScore())

	loadedPlanCrit := loadedBook.GetPlanCrit()
	require.NotNil(t, loadedPlanCrit)
	require.Equal(t, planCrit.GetStrengths(), loadedPlanCrit.GetStrengths())
	require.Equal(t, planCrit.GetImprovements(), loadedPlanCrit.GetImprovements())
	require.Equal(t, planCrit.GetImpressions(), loadedPlanCrit.GetImpressions())
	require.Equal(t, planCrit.GetScore(), loadedPlanCrit.GetScore())

	// Verify critique files were created
	metaCritExists, err := afero.Exists(fs, filepath.Join(path, "meta_critique.md"))
	require.NoError(t, err)
	require.True(t, metaCritExists)

	planCritExists, err := afero.Exists(fs, filepath.Join(path, "plan_critique.md"))
	require.NoError(t, err)
	require.True(t, planCritExists)
}

func TestBookMetaCritGettersSetters(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Test meta critique
	metaCrit := &viter.Critique{
		Strengths:    "Good foundation",
		Improvements: "More detail needed",
		Impressions:  "This is a meta critique for testing purposes.",
		Score:        6,
	}
	err = book.SetMetaCrit(metaCrit)
	require.NoError(t, err)
	require.Equal(t, metaCrit, book.GetMetaCrit())

	// Test plan critique
	planCrit := &viter.Critique{
		Strengths:    "Clear structure",
		Improvements: "Better pacing",
		Impressions:  "This is a plan critique for testing purposes.",
		Score:        5,
	}
	err = book.SetPlanCrit(planCrit)
	require.NoError(t, err)
	require.Equal(t, planCrit, book.GetPlanCrit())

	// Verify files were created
	metaCritExists, err := afero.Exists(fs, filepath.Join(path, "meta_critique.md"))
	require.NoError(t, err)
	require.True(t, metaCritExists)

	planCritExists, err := afero.Exists(fs, filepath.Join(path, "plan_critique.md"))
	require.NoError(t, err)
	require.True(t, planCritExists)
}

func TestRoundTripCritiques(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/critique"

	// Create book
	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set critiques
	metaCrit := &viter.Critique{
		Strengths:    "Good concept",
		Improvements: "More character development",
		Impressions:  "The metadata needs more character development details.",
		Score:        6,
	}
	planCrit := &viter.Critique{
		Strengths:    "Good structure",
		Improvements: "More detailed outlines",
		Impressions:  "The chapter structure could benefit from more detailed outlines.",
		Score:        7,
	}

	err = book.SetMetaCrit(metaCrit)
	require.NoError(t, err)

	err = book.SetPlanCrit(planCrit)
	require.NoError(t, err)

	// Load fresh book and verify critiques persist
	loadedBook, err := viter.LoadBook(fs, path)
	require.NoError(t, err)

	require.Equal(t, metaCrit, loadedBook.GetMetaCrit())
	require.Equal(t, planCrit, loadedBook.GetPlanCrit())
}

func TestChapterGetNumber(t *testing.T) {
	chapter := viter.NewChapter(42, "Test Chapter", "Test content")
	require.Equal(t, 42, chapter.GetNumber())

	// Test chapter with number 0
	chapterZero := viter.NewChapter(0, "Prologue", "Beginning")
	require.Equal(t, 0, chapterZero.GetNumber())
}

func TestChapterFromStringWithNumber(t *testing.T) {
	// Test parsing chapter with number
	chapterStr := `## 5. The Final Confrontation
The hero faces the ultimate challenge.`

	chapter, err := viter.ChapterFromString(chapterStr)
	require.NoError(t, err)
	require.Equal(t, 5, chapter.GetNumber())
	require.Equal(t, "The Final Confrontation", chapter.GetTitle())
	require.Equal(t, "The hero faces the ultimate challenge.", chapter.GetContent())
}

func TestChapterFromStringWithoutNumber(t *testing.T) {
	// Test parsing chapter without number format
	chapterStr := `## Epilogue
The story concludes.`

	chapter, err := viter.ChapterFromString(chapterStr)
	require.NoError(t, err)
	require.Equal(t, 0, chapter.GetNumber())
	require.Equal(t, "Epilogue", chapter.GetTitle())
	require.Equal(t, "The story concludes.", chapter.GetContent())
}

func TestChapterFromStringInvalidNumber(t *testing.T) {
	// Test parsing chapter with invalid number format
	chapterStr := `## abc. Invalid Number
This should not parse the number.`

	chapter, err := viter.ChapterFromString(chapterStr)
	require.NoError(t, err)
	require.Equal(t, 0, chapter.GetNumber())
	require.Equal(t, "abc. Invalid Number", chapter.GetTitle())
	require.Equal(t, "This should not parse the number.", chapter.GetContent())
}

func TestPlanMerge(t *testing.T) {
	// Create original plan
	original := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "Original content 1"),
		viter.NewChapter(2, "Chapter 2", "Original content 2"),
		viter.NewChapter(0, "Prologue", "Original prologue"),
	}

	// Create new plan with overlapping and new chapters
	updates := viter.Plan{
		viter.NewChapter(2, "Chapter 2 Updated", "Updated content 2"),
		viter.NewChapter(3, "Chapter 3", "New content 3"),
		viter.NewChapter(0, "Epilogue", "New epilogue"),
	}

	// Test merge through SetPlan
	fs := afero.NewMemMapFs()
	book, err := viter.CreateBook(fs, "/test")
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
	chaptersByNumber := make(map[int]*viter.Chapter)
	var zeroChapters []*viter.Chapter

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
	book, err := viter.CreateBook(fs, "/test")
	require.NoError(t, err)

	// Set original plan
	original := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "Content 1"),
	}
	err = book.SetPlan(original)
	require.NoError(t, err)

	// Merge with empty plan
	empty := viter.Plan{}
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

	plan, err := viter.PlanFromString(planStr)
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

	plan, err := viter.PlanFromString(planStr)
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
	book, err := viter.CreateBook(fs, "/test")
	require.NoError(t, err)

	// Create original plan with numbered chapters
	original := viter.Plan{
		viter.NewChapter(1, "Chapter 1: The Start", "Beginning of the story"),
		viter.NewChapter(2, "Chapter 2: The Journey", "Middle of the story"),
		viter.NewChapter(0, "Prologue", "Before it all began"),
	}

	err = book.SetPlan(original)
	require.NoError(t, err)

	// Convert plan to string and back
	planStr := original.String()
	parsedPlan, err := viter.PlanFromString(planStr)
	require.NoError(t, err)

	// Verify round-trip preserves numbers
	require.Len(t, parsedPlan, 3)

	// Find chapters by number
	chaptersByNumber := make(map[int]*viter.Chapter)
	var zeroChapters []*viter.Chapter

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
	updates := viter.Plan{
		viter.NewChapter(2, "Chapter 2: The Updated Journey", "Updated middle story"),
		viter.NewChapter(3, "Chapter 3: The End", "End of the story"),
		viter.NewChapter(0, "Epilogue", "After it all ended"),
	}

	// Merge updates
	err = book.SetPlan(updates)
	require.NoError(t, err)

	// Verify merged result maintains all chapters
	merged := book.GetPlan()
	require.Len(t, merged, 5) // 1, updated 2, prologue, new 3, epilogue

	// Reset maps for merged plan
	chaptersByNumber = make(map[int]*viter.Chapter)
	zeroChapters = []*viter.Chapter{}

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
	finalParsed, err := viter.PlanFromString(finalStr)
	require.NoError(t, err)
	require.Len(t, finalParsed, 5)

	// Verify the final parsed version has all the right data
	finalByNumber := make(map[int]*viter.Chapter)
	var finalZeroChapters []*viter.Chapter

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

func TestGetChapter(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Set some chapters
	chapter1 := viter.NewChapter(1, "First Chapter", "This is the first chapter content")
	chapter2 := viter.NewChapter(2, "Second Chapter", "This is the second chapter content")

	err = book.SetChapter(1, chapter1)
	require.NoError(t, err)
	err = book.SetChapter(2, chapter2)
	require.NoError(t, err)

	// Test getting valid chapters
	retrievedChapter1, err := book.GetChapter(1)
	require.NoError(t, err)
	require.Equal(t, chapter1.GetTitle(), retrievedChapter1.GetTitle())
	require.Equal(t, chapter1.GetContent(), retrievedChapter1.GetContent())
	require.Equal(t, chapter1.GetNumber(), retrievedChapter1.GetNumber())

	retrievedChapter2, err := book.GetChapter(2)
	require.NoError(t, err)
	require.Equal(t, chapter2.GetTitle(), retrievedChapter2.GetTitle())
	require.Equal(t, chapter2.GetContent(), retrievedChapter2.GetContent())
	require.Equal(t, chapter2.GetNumber(), retrievedChapter2.GetNumber())

	// Create a new book to test empty chapters
	emptyBook, err := viter.CreateBook(fs, "/test/empty_book")
	require.NoError(t, err)

	// Set up a plan with 3 chapters but don't set any chapter content
	emptyPlan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
		viter.NewChapter(3, "Chapter 3", "Third chapter"),
	}
	err = emptyBook.SetPlan(emptyPlan)
	require.NoError(t, err)

	// Test getting chapter that wasn't set (should return empty chapter)
	emptyChapter, err := emptyBook.GetChapter(2)
	require.NoError(t, err)
	require.Equal(t, "", emptyChapter.GetTitle())
	require.Equal(t, "", emptyChapter.GetContent())
	require.Equal(t, 0, emptyChapter.GetNumber())
}

func TestGetChapterInvalidIndex(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Test negative index
	_, err = book.GetChapter(-1)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	// Test zero index
	_, err = book.GetChapter(0)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	// Test index too high
	_, err = book.GetChapter(3)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)
}

func TestGetChapterCritique(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Create some critiques
	critique1 := &viter.Critique{
		Strengths:    "Good pacing, Strong dialogue",
		Improvements: "Needs more description, Character development",
		Impressions:  "Engaging, Well-written",
		Score:        8,
	}

	critique2 := &viter.Critique{
		Strengths:    "Excellent world-building",
		Improvements: "Plot could be tighter",
		Impressions:  "Creative, Immersive",
		Score:        7,
	}

	// Set the critiques
	err = book.SetChapterCritique(1, critique1)
	require.NoError(t, err)
	err = book.SetChapterCritique(2, critique2)
	require.NoError(t, err)

	// Test getting valid critiques
	retrievedCritique1, err := book.GetChapterCritique(1)
	require.NoError(t, err)
	require.NotNil(t, retrievedCritique1)
	require.Equal(t, critique1.GetStrengths(), retrievedCritique1.GetStrengths())
	require.Equal(t, critique1.GetImprovements(), retrievedCritique1.GetImprovements())
	require.Equal(t, critique1.GetImpressions(), retrievedCritique1.GetImpressions())
	require.Equal(t, critique1.GetScore(), retrievedCritique1.GetScore())

	retrievedCritique2, err := book.GetChapterCritique(2)
	require.NoError(t, err)
	require.NotNil(t, retrievedCritique2)
	require.Equal(t, critique2.GetStrengths(), retrievedCritique2.GetStrengths())
	require.Equal(t, critique2.GetImprovements(), retrievedCritique2.GetImprovements())
	require.Equal(t, critique2.GetImpressions(), retrievedCritique2.GetImpressions())
	require.Equal(t, critique2.GetScore(), retrievedCritique2.GetScore())

	// Create a new book to test empty chapter critiques
	emptyCritBook, err := viter.CreateBook(fs, "/test/empty_crit_book")
	require.NoError(t, err)

	// Set up a plan with 3 chapters but don't set any chapter critiques
	emptyCritPlan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
		viter.NewChapter(3, "Chapter 3", "Third chapter"),
	}
	err = emptyCritBook.SetPlan(emptyCritPlan)
	require.NoError(t, err)

	// Test getting chapter critique that wasn't set (should return empty critique)
	emptyCritique, err := emptyCritBook.GetChapterCritique(2)
	require.NoError(t, err)
	require.NotNil(t, emptyCritique)
	require.Equal(t, "", emptyCritique.GetStrengths())
	require.Equal(t, "", emptyCritique.GetImprovements())
	require.Equal(t, "", emptyCritique.GetImpressions())
	require.Equal(t, 0, emptyCritique.GetScore())
}

func TestGetChapterCritiqueInvalidIndex(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Test negative index
	_, err = book.GetChapterCritique(-1)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	// Test zero index
	_, err = book.GetChapterCritique(0)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	// Test index too high
	_, err = book.GetChapterCritique(3)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)
}

func TestSetChapterCritique(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Create a critique
	critique := &viter.Critique{
		Strengths:    "Good pacing, Strong dialogue",
		Improvements: "Needs more description, Character development",
		Impressions:  "Engaging, Well-written",
		Score:        8,
	}

	// Set the critique
	err = book.SetChapterCritique(1, critique)
	require.NoError(t, err)

	// Verify the critique was saved to filesystem
	critiquePath := filepath.Join(path, "chapter_1_critique.md")
	exists, err := afero.Exists(fs, critiquePath)
	require.NoError(t, err)
	require.True(t, exists)

	// Read the file and verify content
	content, err := afero.ReadFile(fs, critiquePath)
	require.NoError(t, err)
	require.Contains(t, string(content), "Good pacing")
	require.Contains(t, string(content), "Needs more description")
	require.Contains(t, string(content), "Engaging")
	require.Contains(t, string(content), "8")

	// Verify we can retrieve the critique
	retrievedCritique, err := book.GetChapterCritique(1)
	require.NoError(t, err)
	require.NotNil(t, retrievedCritique)
	require.Equal(t, critique.GetStrengths(), retrievedCritique.GetStrengths())
	require.Equal(t, critique.GetImprovements(), retrievedCritique.GetImprovements())
	require.Equal(t, critique.GetImpressions(), retrievedCritique.GetImpressions())
	require.Equal(t, critique.GetScore(), retrievedCritique.GetScore())
}

func TestSetChapterCritiqueInvalidIndex(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	critique := &viter.Critique{
		Strengths:    "Good pacing",
		Improvements: "Needs work",
		Impressions:  "Okay",
		Score:        5,
	}

	// Test negative index
	err = book.SetChapterCritique(-1, critique)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	// Test zero index
	err = book.SetChapterCritique(0, critique)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)

	// Test index too high
	err = book.SetChapterCritique(3, critique)
	require.Error(t, err)
	require.Equal(t, viter.ErrInvalidChapterIndex, err)
}

func TestSetChapterCritiqueExpandsSlice(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 3 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
		viter.NewChapter(3, "Chapter 3", "Third chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Create critiques
	critique1 := &viter.Critique{
		Strengths:    "Good start",
		Improvements: "Needs polish",
		Impressions:  "Promising",
		Score:        6,
	}

	critique3 := &viter.Critique{
		Strengths:    "Great ending",
		Improvements: "Minor issues",
		Impressions:  "Satisfying",
		Score:        9,
	}

	// Set critique for chapter 3 first (should expand slice)
	err = book.SetChapterCritique(3, critique3)
	require.NoError(t, err)

	// Set critique for chapter 1
	err = book.SetChapterCritique(1, critique1)
	require.NoError(t, err)

	// Verify both critiques can be retrieved
	retrievedCritique1, err := book.GetChapterCritique(1)
	require.NoError(t, err)
	require.Equal(t, critique1.GetScore(), retrievedCritique1.GetScore())

	retrievedCritique3, err := book.GetChapterCritique(3)
	require.NoError(t, err)
	require.Equal(t, critique3.GetScore(), retrievedCritique3.GetScore())

	// Chapter 2 should have an empty critique (default)
	retrievedCritique2, err := book.GetChapterCritique(2)
	require.NoError(t, err)
	require.NotNil(t, retrievedCritique2)
	require.Equal(t, "", retrievedCritique2.GetStrengths())
	require.Equal(t, "", retrievedCritique2.GetImprovements())
	require.Equal(t, "", retrievedCritique2.GetImpressions())
	require.Equal(t, 0, retrievedCritique2.GetScore())
}

func TestGetChapterMixedSetAndUnset(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := viter.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 4 chapters
	plan := viter.Plan{
		viter.NewChapter(1, "Chapter 1", "First chapter"),
		viter.NewChapter(2, "Chapter 2", "Second chapter"),
		viter.NewChapter(3, "Chapter 3", "Third chapter"),
		viter.NewChapter(4, "Chapter 4", "Fourth chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Set only chapters 1 and 3, leave 2 and 4 unset
	chapter1 := viter.NewChapter(1, "First Chapter Content", "This is the first chapter")
	chapter3 := viter.NewChapter(3, "Third Chapter Content", "This is the third chapter")

	err = book.SetChapter(1, chapter1)
	require.NoError(t, err)
	err = book.SetChapter(3, chapter3)
	require.NoError(t, err)

	// Test getting set chapters
	retrievedChapter1, err := book.GetChapter(1)
	require.NoError(t, err)
	require.Equal(t, "First Chapter Content", retrievedChapter1.GetTitle())
	require.Equal(t, "This is the first chapter", retrievedChapter1.GetContent())
	require.Equal(t, 1, retrievedChapter1.GetNumber())

	retrievedChapter3, err := book.GetChapter(3)
	require.NoError(t, err)
	require.Equal(t, "Third Chapter Content", retrievedChapter3.GetTitle())
	require.Equal(t, "This is the third chapter", retrievedChapter3.GetContent())
	require.Equal(t, 3, retrievedChapter3.GetNumber())

	// Test getting unset chapters (should return empty chapters)
	emptyChapter2, err := book.GetChapter(2)
	require.NoError(t, err)
	require.Equal(t, "", emptyChapter2.GetTitle())
	require.Equal(t, "", emptyChapter2.GetContent())
	require.Equal(t, 0, emptyChapter2.GetNumber())

	emptyChapter4, err := book.GetChapter(4)
	require.NoError(t, err)
	require.Equal(t, "", emptyChapter4.GetTitle())
	require.Equal(t, "", emptyChapter4.GetContent())
	require.Equal(t, 0, emptyChapter4.GetNumber())

	// Verify that setting a chapter after getting an empty one works
	chapter2 := viter.NewChapter(2, "Second Chapter Content", "This is the second chapter")
	err = book.SetChapter(2, chapter2)
	require.NoError(t, err)

	// Now getting chapter 2 should return the set content
	retrievedChapter2, err := book.GetChapter(2)
	require.NoError(t, err)
	require.Equal(t, "Second Chapter Content", retrievedChapter2.GetTitle())
	require.Equal(t, "This is the second chapter", retrievedChapter2.GetContent())
	require.Equal(t, 2, retrievedChapter2.GetNumber())
}

func TestBookMetaMergedCopy(t *testing.T) {
	// Test merging with empty target
	original := &viter.BookMeta{}
	original.SetStyle("Fantasy")
	original.SetGenres([]string{"Epic Fantasy", "Adventure"})
	original.SetLogline("A hero's journey")
	original.SetWorld("Middle Earth")
	original.SetProtagonists([]viter.Character{
		viter.NewCharacter("Frodo", "A hobbit"),
	})
	original.SetAntagonists([]viter.Character{
		viter.NewCharacter("Sauron", "Dark Lord"),
	})
	original.SetMinorCharacters([]viter.Character{
		viter.NewCharacter("Sam", "Loyal friend"),
	})
	original.SetPlot("The ring must be destroyed")
	original.SetTitle("The Lord of the Rings")

	other := &viter.BookMeta{}
	other.SetStyle("Science Fiction")
	other.SetGenres([]string{"Space Opera"})
	other.SetLogline("A galactic adventure")
	other.SetWorld("Galaxy Far Far Away")
	other.SetProtagonists([]viter.Character{
		viter.NewCharacter("Luke", "Jedi Knight"),
	})
	other.SetAntagonists([]viter.Character{
		viter.NewCharacter("Vader", "Sith Lord"),
	})
	other.SetMinorCharacters([]viter.Character{
		viter.NewCharacter("Han", "Smuggler"),
	})
	other.SetPlot("Destroy the Death Star")
	other.SetTitle("Star Wars")

	// Test complete merge
	merged := original.MergedCopy(other)

	// Verify original is unchanged
	require.Equal(t, "Fantasy", original.GetStyle())
	require.Equal(t, []string{"Epic Fantasy", "Adventure"}, original.GetGenres())
	require.Equal(t, "A hero's journey", original.GetLogline())
	require.Equal(t, "Middle Earth", original.GetWorld())
	require.Equal(t, "The ring must be destroyed", original.GetPlot())
	require.Equal(t, "The Lord of the Rings", original.GetTitle())

	// Verify merged has other's values
	require.Equal(t, "Science Fiction", merged.GetStyle())
	require.Equal(t, []string{"Space Opera"}, merged.GetGenres())
	require.Equal(t, "A galactic adventure", merged.GetLogline())
	require.Equal(t, "Galaxy Far Far Away", merged.GetWorld())
	require.Equal(t, "Destroy the Death Star", merged.GetPlot())
	require.Equal(t, "Star Wars", merged.GetTitle())
	require.Len(t, merged.GetProtagonists(), 1)
	require.Equal(t, "Luke", merged.GetProtagonists()[0].GetName())
	require.Len(t, merged.GetAntagonists(), 1)
	require.Equal(t, "Vader", merged.GetAntagonists()[0].GetName())
	require.Len(t, merged.GetMinorCharacters(), 1)
	require.Equal(t, "Han", merged.GetMinorCharacters()[0].GetName())
}

func TestBookMetaMergedCopyPartial(t *testing.T) {
	original := &viter.BookMeta{}
	original.SetStyle("Fantasy")
	original.SetGenres([]string{"Epic Fantasy"})
	original.SetLogline("A hero's journey")
	original.SetWorld("Middle Earth")
	original.SetPlot("The ring must be destroyed")
	original.SetTitle("The Lord of the Rings")

	// Only some fields set in other
	other := &viter.BookMeta{}
	other.SetStyle("Dark Fantasy")
	other.SetLogline("A darker journey")
	// Other fields empty/nil

	merged := original.MergedCopy(other)

	// Verify original unchanged
	require.Equal(t, "Fantasy", original.GetStyle())
	require.Equal(t, "A hero's journey", original.GetLogline())

	// Verify merged has mixed values
	require.Equal(t, "Dark Fantasy", merged.GetStyle())              // From other
	require.Equal(t, []string{"Epic Fantasy"}, merged.GetGenres())   // From original
	require.Equal(t, "A darker journey", merged.GetLogline())        // From other
	require.Equal(t, "Middle Earth", merged.GetWorld())              // From original
	require.Equal(t, "The ring must be destroyed", merged.GetPlot()) // From original
	require.Equal(t, "The Lord of the Rings", merged.GetTitle())     // From original
}

func TestBookMetaMergedCopyEmpty(t *testing.T) {
	original := &viter.BookMeta{}
	original.SetStyle("Fantasy")
	original.SetTitle("Original Title")

	empty := &viter.BookMeta{}
	merged := original.MergedCopy(empty)

	// Should be identical to original since other is empty
	require.Equal(t, original.GetStyle(), merged.GetStyle())
	require.Equal(t, original.GetTitle(), merged.GetTitle())

	// But should be different objects
	require.NotSame(t, original, merged)
}

func TestBookMetaMergedCopyNil(t *testing.T) {
	original := &viter.BookMeta{}
	original.SetStyle("Fantasy")
	original.SetTitle("Original Title")

	// Test with nil other - should not panic and return copy of original
	merged := original.MergedCopy(nil)
	require.Equal(t, original.GetStyle(), merged.GetStyle())
	require.Equal(t, original.GetTitle(), merged.GetTitle())
	require.NotSame(t, original, merged)
}

func TestPlanMergedCopy(t *testing.T) {
	ch1 := viter.NewChapter(1, "Chapter 1", "Content 1")
	ch2 := viter.NewChapter(2, "Chapter 2", "Content 2")
	ch0 := viter.NewChapter(0, "Unnumbered", "No number")

	original := viter.Plan{ch1, ch2, ch0}

	newCh2 := viter.NewChapter(2, "New Chapter 2", "New Content 2")
	ch3 := viter.NewChapter(3, "Chapter 3", "Content 3")
	anotherCh0 := viter.NewChapter(0, "Another Unnumbered", "Also no number")

	other := viter.Plan{newCh2, ch3, anotherCh0}

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
	ch1 := viter.NewChapter(1, "Chapter 1", "Content 1")
	original := viter.Plan{ch1}

	empty := viter.Plan{}
	merged := original.MergedCopy(empty)

	// Should be identical to original
	require.Len(t, merged, 1)
	require.Equal(t, original[0].GetNumber(), merged[0].GetNumber())
	require.Equal(t, original[0].GetTitle(), merged[0].GetTitle())
	require.Equal(t, original[0].GetContent(), merged[0].GetContent())
}

func TestPlanMergedCopyOnlyUnnumbered(t *testing.T) {
	ch1 := viter.NewChapter(1, "Chapter 1", "Content 1")
	original := viter.Plan{ch1}

	unCh1 := viter.NewChapter(0, "Unnumbered 1", "No number 1")
	unCh2 := viter.NewChapter(0, "Unnumbered 2", "No number 2")
	other := viter.Plan{unCh1, unCh2}

	merged := original.MergedCopy(other)

	require.Len(t, merged, 3) // 1 original + 2 unnumbered
	require.Equal(t, 1, merged[0].GetNumber())
	require.Equal(t, 0, merged[1].GetNumber())
	require.Equal(t, 0, merged[2].GetNumber())
	require.Equal(t, "Unnumbered 1", merged[1].GetTitle())
	require.Equal(t, "Unnumbered 2", merged[2].GetTitle())
}

func TestPlanMergedCopyAllReplaced(t *testing.T) {
	oldCh1 := viter.NewChapter(1, "Old Chapter 1", "Old Content 1")
	oldCh2 := viter.NewChapter(2, "Old Chapter 2", "Old Content 2")
	original := viter.Plan{oldCh1, oldCh2}

	newCh1 := viter.NewChapter(1, "New Chapter 1", "New Content 1")
	newCh2 := viter.NewChapter(2, "New Chapter 2", "New Content 2")
	other := viter.Plan{newCh1, newCh2}

	merged := original.MergedCopy(other)

	require.Len(t, merged, 2)
	require.Equal(t, "New Chapter 1", merged[0].GetTitle())
	require.Equal(t, "New Chapter 2", merged[1].GetTitle())
}

func TestPlanMergedCopyEmptyOriginal(t *testing.T) {
	// Test merging into an empty plan
	empty := viter.Plan{}

	ch1 := viter.NewChapter(1, "Chapter 1", "Content 1")
	ch2 := viter.NewChapter(0, "Unnumbered", "No number")
	other := viter.Plan{ch1, ch2}

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
