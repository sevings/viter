package books_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"viter/internal/books"
)

func TestCreateBook(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := books.CreateBook(fs, path)
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
	book, err := books.LoadBook(fs, path)
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

	book, err := books.LoadBook(fs, path)
	require.Error(t, err)
	require.Equal(t, books.ErrNoMetaFile, err)
	require.Nil(t, book)
}

func TestBookSave(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set metadata
	meta := &books.BookMeta{}
	meta.SetStyle("Science Fiction")
	meta.SetGenres([]string{"Sci-Fi", "Thriller"})
	meta.SetWorld("A dystopian future")
	meta.SetProtagonists([]books.Character{
		books.NewCharacter("Alex", "A rebel hacker"),
	})
	meta.SetPlot("The fight against a totalitarian regime")

	err = book.SetMeta(meta)
	require.NoError(t, err)

	// Set plan
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1: The Resistance", "Alex joins the underground"),
		books.NewChapter(2, "Chapter 2: The Mission", "The first strike against the system"),
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

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan first
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Save a chapter
	chapter := books.NewChapter(1, "Chapter 1: The Beginning", "It was a dark and stormy night...")
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

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	chapter := books.NewChapter(0, "Invalid Chapter", "This shouldn't work")

	// Test negative index
	err = book.SetChapter(-1, chapter)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)

	// Test index too high
	err = book.SetChapter(3, chapter)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)
}

func TestGetPlanChapter(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter content"),
		books.NewChapter(2, "Chapter 2", "Second chapter content"),
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
	require.Equal(t, books.ErrInvalidChapterIndex, err)

	_, err = book.GetPlanChapter(3)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)
}

func TestIntegrationSaveAndLoad(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/integration"

	// Create and configure a book
	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up complete metadata
	meta := &books.BookMeta{}
	meta.SetTitle("Shadows of Millbrook")
	meta.SetStyle("Horror")
	meta.SetGenres([]string{"Horror", "Thriller", "Supernatural"})
	meta.SetLogline("In a town where the dead don't rest, the living must face their darkest fears.")
	meta.SetWorld("A small town with dark secrets")
	meta.SetProtagonists([]books.Character{
		books.NewCharacter("Dr. Emma Carter", "A psychiatrist who uncovers the truth"),
		books.NewCharacter("Father Miguel", "A priest battling ancient evils"),
	})
	meta.SetMinorCharacters([]books.Character{
		books.NewCharacter("Sheriff Brooks", "The local law enforcement"),
	})
	meta.SetAntagonists([]books.Character{
		books.NewCharacter("The Hollow Man", "An ancient spirit seeking revenge"),
	})
	meta.SetPlot("A town's buried secrets come back to haunt the living")

	err = book.SetMeta(meta)
	require.NoError(t, err)

	// Set up a plan
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1: Arrival", "Dr. Carter arrives in the small town"),
		books.NewChapter(2, "Chapter 2: Strange Occurrences", "Mysterious events begin to unfold"),
		books.NewChapter(3, "Chapter 3: The Truth", "The dark history is revealed"),
	}

	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Set up critiques
	metaCrit := &books.Critique{
		Strengths:    "Well-developed characters",
		Improvements: "Need more horror elements",
		Impressions:  "The horror elements are well-balanced with character development.",
		Score:        8,
	}
	err = book.SetMetaCrit(metaCrit)
	require.NoError(t, err)

	planCrit := &books.Critique{
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
	fullChapter1 := books.NewChapter(1, "Chapter 1: Arrival - The Beginning", "Dr. Emma Carter stepped off the bus into the dusty main street of Millbrook. The town seemed ordinary enough, but something in the air made her skin crawl.")
	err = book.SetChapter(1, fullChapter1)
	require.NoError(t, err)

	// Load the book fresh
	loadedBook, err := books.LoadBook(fs, path)
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

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Test meta critique
	metaCrit := &books.Critique{
		Strengths:    "Good foundation",
		Improvements: "More detail needed",
		Impressions:  "This is a meta critique for testing purposes.",
		Score:        6,
	}
	err = book.SetMetaCrit(metaCrit)
	require.NoError(t, err)
	require.Equal(t, metaCrit, book.GetMetaCrit())

	// Test plan critique
	planCrit := &books.Critique{
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
	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set critiques
	metaCrit := &books.Critique{
		Strengths:    "Good concept",
		Improvements: "More character development",
		Impressions:  "The metadata needs more character development details.",
		Score:        6,
	}
	planCrit := &books.Critique{
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
	loadedBook, err := books.LoadBook(fs, path)
	require.NoError(t, err)

	require.Equal(t, metaCrit, loadedBook.GetMetaCrit())
	require.Equal(t, planCrit, loadedBook.GetPlanCrit())
}

func TestGetChapter(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Set some chapters
	chapter1 := books.NewChapter(1, "First Chapter", "This is the first chapter content")
	chapter2 := books.NewChapter(2, "Second Chapter", "This is the second chapter content")

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
	emptyBook, err := books.CreateBook(fs, "/test/empty_book")
	require.NoError(t, err)

	// Set up a plan with 3 chapters but don't set any chapter content
	emptyPlan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
		books.NewChapter(3, "Chapter 3", "Third chapter"),
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

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Test negative index
	_, err = book.GetChapter(-1)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)

	// Test zero index
	_, err = book.GetChapter(0)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)

	// Test index too high
	_, err = book.GetChapter(3)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)
}

func TestGetChapterCritique(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Create some critiques
	critique1 := &books.Critique{
		Strengths:    "Good pacing, Strong dialogue",
		Improvements: "Needs more description, Character development",
		Impressions:  "Engaging, Well-written",
		Score:        8,
	}

	critique2 := &books.Critique{
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
	emptyCritBook, err := books.CreateBook(fs, "/test/empty_crit_book")
	require.NoError(t, err)

	// Set up a plan with 3 chapters but don't set any chapter critiques
	emptyCritPlan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
		books.NewChapter(3, "Chapter 3", "Third chapter"),
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

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Test negative index
	_, err = book.GetChapterCritique(-1)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)

	// Test zero index
	_, err = book.GetChapterCritique(0)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)

	// Test index too high
	_, err = book.GetChapterCritique(3)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)
}

func TestSetChapterCritique(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Create a critique
	critique := &books.Critique{
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

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 2 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	critique := &books.Critique{
		Strengths:    "Good pacing",
		Improvements: "Needs work",
		Impressions:  "Okay",
		Score:        5,
	}

	// Test negative index
	err = book.SetChapterCritique(-1, critique)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)

	// Test zero index
	err = book.SetChapterCritique(0, critique)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)

	// Test index too high
	err = book.SetChapterCritique(3, critique)
	require.Error(t, err)
	require.Equal(t, books.ErrInvalidChapterIndex, err)
}

func TestSetChapterCritiqueExpandsSlice(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/test/book"

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 3 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
		books.NewChapter(3, "Chapter 3", "Third chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Create critiques
	critique1 := &books.Critique{
		Strengths:    "Good start",
		Improvements: "Needs polish",
		Impressions:  "Promising",
		Score:        6,
	}

	critique3 := &books.Critique{
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

	book, err := books.CreateBook(fs, path)
	require.NoError(t, err)

	// Set up a plan with 4 chapters
	plan := books.Plan{
		books.NewChapter(1, "Chapter 1", "First chapter"),
		books.NewChapter(2, "Chapter 2", "Second chapter"),
		books.NewChapter(3, "Chapter 3", "Third chapter"),
		books.NewChapter(4, "Chapter 4", "Fourth chapter"),
	}
	err = book.SetPlan(plan)
	require.NoError(t, err)

	// Set only chapters 1 and 3, leave 2 and 4 unset
	chapter1 := books.NewChapter(1, "First Chapter Content", "This is the first chapter")
	chapter3 := books.NewChapter(3, "Third Chapter Content", "This is the third chapter")

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
	chapter2 := books.NewChapter(2, "Second Chapter Content", "This is the second chapter")
	err = book.SetChapter(2, chapter2)
	require.NoError(t, err)

	// Now getting chapter 2 should return the set content
	retrievedChapter2, err := book.GetChapter(2)
	require.NoError(t, err)
	require.Equal(t, "Second Chapter Content", retrievedChapter2.GetTitle())
	require.Equal(t, "This is the second chapter", retrievedChapter2.GetContent())
	require.Equal(t, 2, retrievedChapter2.GetNumber())
}
