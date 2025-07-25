package viter_test

import (
	"fmt"
	"strconv"
	"testing"

	"viter/internal/neural"
	"viter/internal/viter"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

// Mock implementations for testing
type mockPromptProvider struct {
	writeMetaPrompt        string
	critiqueMetaPrompt     string
	updateMetaPrompt       string
	writePlanPrompt        string
	critiquePlanPrompt     string
	updatePlanPrompt       string
	writeChapterPrompt     string
	critiqueChapterPrompt  string
	updateChapterPrompt    string
	writeNChapterPrompt    string
	critiqueNChapterPrompt string
	updateNChapterPrompt   string
	correctTextPrompt      string
}

func (m *mockPromptProvider) WriteMetaPrompt() string {
	return m.writeMetaPrompt
}

func (m *mockPromptProvider) CritiqueMetaPrompt() string {
	return m.critiqueMetaPrompt
}

func (m *mockPromptProvider) UpdateMetaPrompt() string {
	return m.updateMetaPrompt
}

func (m *mockPromptProvider) WritePlanPrompt(chapterCount int) string {
	return m.writePlanPrompt
}

func (m *mockPromptProvider) CritiquePlanPrompt() string {
	return m.critiquePlanPrompt
}

func (m *mockPromptProvider) UpdatePlanPrompt() string {
	return m.updatePlanPrompt
}

func (m *mockPromptProvider) WriteChapterPrompt() string {
	return m.writeChapterPrompt
}

func (m *mockPromptProvider) CritiqueChapterPrompt() string {
	return m.critiqueChapterPrompt
}

func (m *mockPromptProvider) UpdateChapterPrompt() string {
	return m.updateChapterPrompt
}

func (m *mockPromptProvider) WriteNChapterPrompt(chapter int) string {
	return fmt.Sprintf(m.writeNChapterPrompt, chapter)
}

func (m *mockPromptProvider) CritiqueNChapterPrompt(chapter int) string {
	return fmt.Sprintf(m.critiqueNChapterPrompt, chapter)
}

func (m *mockPromptProvider) UpdateNChapterPrompt(chapter int) string {
	return fmt.Sprintf(m.updateNChapterPrompt, chapter)
}

func (m *mockPromptProvider) CorrectTextPrompt() string {
	return m.correctTextPrompt
}

type mockTextGenerator struct {
	responses       []string
	callCount       int
	shouldFail      bool
	infiniteMode    bool
	defaultResponse string
}

func (m *mockTextGenerator) GenerateText(messages []llms.MessageContent) (string, bool) {
	if m.shouldFail {
		return "", false
	}

	if m.callCount >= len(m.responses) {
		if m.infiniteMode {
			m.callCount++
			if m.defaultResponse != "" {
				return m.defaultResponse, true
			}
			return fmt.Sprintf("Generated response %d", m.callCount), true
		}
		return "", false
	}

	response := m.responses[m.callCount]
	m.callCount++
	return response, true
}

func (m *mockTextGenerator) reset() {
	m.callCount = 0
	m.shouldFail = false
}

func createTestConfig() viter.Config {
	return viter.Config{
		Ai: neural.AiConfig{
			Provider: "test",
			Model:    "test-model",
		},
	}
}

func createTestPromptProvider() *mockPromptProvider {
	return &mockPromptProvider{
		writeMetaPrompt:        "Write meta prompt",
		critiqueMetaPrompt:     "Critique meta prompt",
		updateMetaPrompt:       "Update meta prompt",
		writePlanPrompt:        "Write plan prompt",
		critiquePlanPrompt:     "Critique plan prompt",
		updatePlanPrompt:       "Update plan prompt",
		writeChapterPrompt:     "Write chapter prompt",
		critiqueChapterPrompt:  "Critique chapter prompt",
		updateChapterPrompt:    "Update chapter prompt",
		writeNChapterPrompt:    "Write chapter %d prompt",
		critiqueNChapterPrompt: "Critique chapter %d prompt",
		updateNChapterPrompt:   "Update chapter %d prompt",
		correctTextPrompt:      "Correct text prompt",
	}
}

func createValidMetaResponse() string {
	return `## Style
Epic Fantasy

## Genres
Fantasy, Adventure

## Logline
A young hero embarks on a quest to save the kingdom

## World
A medieval fantasy world with magic and dragons

## Protagonists
### Hero
A brave young warrior

## Antagonists
### Dark Lord
An evil sorcerer

## Minor Characters
### Mentor
A wise old wizard

## Plot
The hero must find the magical sword to defeat the dark lord

## Title
The Quest for the Sword`
}

func createValidCritiqueResponse(score int) string {
	return `## Strengths
Good character development

## Improvements
Needs more world building

## Impressions
Shows promise but needs work

## Score
` + strconv.Itoa(score)
}

func createValidPlanResponse() string {
	return `## 1. The Beginning
The hero starts his journey

## 2. The Quest
The hero searches for the sword

## 3. The Final Battle
The hero fights the dark lord`
}

func createValidChapterResponse(number int, title, content string) string {
	return fmt.Sprintf(`## %d. %s
%s`, number, title, content)
}

func createValidDiffResponse() string {
	return `+This is an improved line of text
 This line remains unchanged
-This line was removed`
}

// Basic Constructor Tests
func TestNewViter(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}

	v, ok := viter.NewViter(cfg, pp, tg)

	require.True(t, ok)
	require.NotNil(t, v)
}

func TestViter_CreateBook_Success(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	result := v.CreateBook(fs, path)

	require.True(t, result)
	require.NotNil(t, v.GetBook())
}

func TestViter_CreateBook_FileSystemError(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
	path := "/test/book"

	result := v.CreateBook(fs, path)

	require.False(t, result)
}

func TestViter_LoadBook_Success(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create the book first
	require.True(t, v.CreateBook(fs, path))

	// Then load it
	result := v.LoadBook(fs, path)

	require.True(t, result)
}

// UpdateMeta Tests
func TestViter_UpdateMeta_NoBook(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	result := v.UpdateMeta(1)

	require.False(t, result)
}

func TestViter_UpdateMeta_EmptyMeta_SingleIteration(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(9),
			createValidMetaResponse(),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(1)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // write meta + 1 critique (no update after last critique)
}

func TestViter_UpdateMeta_MultipleIterations(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),      // Initial write
			createValidCritiqueResponse(5), // First iteration critique
			createValidMetaResponse(),      // First iteration update
			createValidCritiqueResponse(7), // Second iteration critique
			createValidMetaResponse(),      // Second iteration update
			createValidCritiqueResponse(9), // Third iteration critique
			createValidMetaResponse(),      // Third iteration update
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(3)

	require.True(t, result)
	require.Equal(t, 6, tg.callCount) // write meta + 3 critiques + 2 updates (no update after last critique)
}

func TestViter_UpdateMeta_ZeroIterations(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(0)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // write meta + 1 critique (always at least one critique)
}

func TestViter_UpdateMeta_NegativeIterations(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(-1)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // write meta + 1 critique (always at least one critique)
}

func TestViter_UpdateMeta_WriteMetaFailure(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		shouldFail: true,
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(1)

	require.False(t, result)
	require.Equal(t, 0, tg.callCount)
}

func TestViter_UpdateMeta_CritiqueFailure(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	// Make it fail on the second call (critique)
	tg.shouldFail = true
	tg.callCount = 1

	result := v.UpdateMeta(1)

	require.False(t, result)
	require.Equal(t, 1, tg.callCount) // Should fail on critique call
}

func TestViter_UpdateMeta_AlreadyFilledMeta(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidCritiqueResponse(9),
			createValidMetaResponse(),
			createValidCritiqueResponse(19),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book with already filled meta
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)

	metaContent := createValidMetaResponse()
	err = afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateMeta(1)

	require.True(t, result)
	require.Equal(t, 3, tg.callCount)
}

// UpdatePlan Tests
func TestViter_UpdatePlan_NoBook(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	result := v.UpdatePlan(3, 1)

	require.False(t, result)
	require.Equal(t, 0, tg.callCount)
}

func TestViter_UpdatePlan_Success(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),      // CreateBook writes meta
			createValidCritiqueResponse(8), // UpdateMeta(0) critiques meta
			createValidPlanResponse(),      // UpdatePlan writes plan
			createValidCritiqueResponse(8), // UpdatePlan critiques plan
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(0)) // critique only (meta already exists)

	result := v.UpdatePlan(3, 1)

	require.True(t, result)

	require.Equal(t, 4, tg.callCount) // write meta + critique meta + write plan + critique plan
}

// UpdateChapter Tests
func TestViter_UpdateChapter_NoBook(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	result := v.UpdateChapter(1, 1)

	require.False(t, result)
	require.Equal(t, 0, tg.callCount)
}

func TestViter_UpdateChapter_InvalidChapterNumber(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book with meta and plan
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)

	metaContent := createValidMetaResponse()
	err = afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)

	planContent := createValidPlanResponse()
	err = afero.WriteFile(fs, path+"/plan.md", []byte(planContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	// Test invalid chapter numbers
	require.False(t, v.UpdateChapter(0, 1))  // Chapter 0 doesn't exist
	require.False(t, v.UpdateChapter(-1, 1)) // Negative chapter
	require.False(t, v.UpdateChapter(5, 1))  // Chapter beyond plan length

	require.Equal(t, 0, tg.callCount)
}

func TestViter_UpdateChapter_EmptyChapter_Success(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidChapterResponse(1, "The Beginning", "This is the start of our story."),
			createValidCritiqueResponse(85),
			createValidDiffResponse(),
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book with meta and plan
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)

	metaContent := createValidMetaResponse()
	err = afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)

	planContent := createValidPlanResponse()
	err = afero.WriteFile(fs, path+"/plan.md", []byte(planContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateChapter(1, 1)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // write chapter + 1 critique (no update after last critique)

	chapter, err := v.GetBook().GetChapter(1)
	require.NoError(t, err)
	require.NotNil(t, chapter)
	require.Equal(t, "The Beginning", chapter.GetTitle())
}

func TestViter_UpdateChapter_MultipleIterations(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidChapterResponse(1, "The Beginning", "This is the start of our story."),
			createValidCritiqueResponse(60), // First iteration critique
			createValidDiffResponse(),       // First iteration update
			createValidCritiqueResponse(70), // Second iteration critique
			createValidDiffResponse(),       // Second iteration update
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book with meta and plan
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)

	metaContent := createValidMetaResponse()
	err = afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)

	planContent := createValidPlanResponse()
	err = afero.WriteFile(fs, path+"/plan.md", []byte(planContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateChapter(1, 2)

	require.True(t, result)
	require.Equal(t, 4, tg.callCount) // write chapter + 2 critiques + 1 update (no update after last critique)
}

func TestViter_UpdateChapter_ZeroIterations(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidChapterResponse(1, "The Beginning", "This is the start of our story."),
			createValidCritiqueResponse(85),
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book with meta and plan
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)

	metaContent := createValidMetaResponse()
	err = afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)

	planContent := createValidPlanResponse()
	err = afero.WriteFile(fs, path+"/plan.md", []byte(planContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateChapter(1, 0)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // write chapter + 1 critique (always at least one critique)
}

func TestViter_UpdateChapter_NegativeIterations(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidChapterResponse(1, "The Beginning", "This is the start of our story."),
			createValidCritiqueResponse(85),
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book with meta and plan
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)

	metaContent := createValidMetaResponse()
	err = afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)

	planContent := createValidPlanResponse()
	err = afero.WriteFile(fs, path+"/plan.md", []byte(planContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateChapter(1, -1)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // write chapter + 1 critique (always at least one critique)
}

func TestViter_UpdateChapter_WriteChapterFailure(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		shouldFail: true,
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)
	require.NotNil(t, v)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book with meta and plan
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)

	metaContent := createValidMetaResponse()
	err = afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)

	planContent := createValidPlanResponse()
	err = afero.WriteFile(fs, path+"/plan.md", []byte(planContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateChapter(1, 1)

	require.False(t, result)
	require.Equal(t, 0, tg.callCount)
}

// Integration Tests
func TestViter_FullWorkflow(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(9),
			createValidPlanResponse(),
			createValidCritiqueResponse(8),
			createValidChapterResponse(1, "Chapter 1", "Content of chapter 1"),
			createValidCritiqueResponse(85),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book
	require.True(t, v.CreateBook(fs, path))

	// Update meta with 1 iteration
	require.True(t, v.UpdateMeta(1))

	// Update plan with 1 iteration
	require.True(t, v.UpdatePlan(3, 1))

	// Update chapter with 1 iteration
	require.True(t, v.UpdateChapter(1, 1))

	require.Equal(t, 6, tg.callCount) // All operations completed successfully
}

func TestViter_CorrectChapter_NoBook(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Should fail when no book is loaded
	result := v.CorrectChapter(1, 1000)
	require.False(t, result)
}

func TestViter_CorrectChapter_NoPlan(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Create book without plan
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))

	// Should fail when no plan exists
	result := v.CorrectChapter(1, 1000)
	require.False(t, result)
}

func TestViter_CorrectChapter_SmallText(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Test Chapter", "Test chapter content."),
			createValidCritiqueResponse(8),
			"Corrected small text content.",
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with chapter
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(3, 1))
	require.True(t, v.UpdateChapter(1, 1))

	// Correct chapter with maxLen larger than content
	result := v.CorrectChapter(1, 1000)
	require.True(t, result)
	require.Equal(t, 7, tg.callCount) // Should have made 6 setup calls + 1 correction call

	// Verify chapter content was updated
	book := v.GetBook()
	chp, err := book.GetChapter(1)
	require.NoError(t, err)
	require.Equal(t, "Corrected small text content.", chp.GetContent())
}

func TestViter_CorrectAllChapters_NoBook(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Should fail when no book is loaded
	result := v.CorrectAllChapters(1000)
	require.False(t, result)
}

func TestViter_CorrectAllChapters_NoPlan(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Create book without plan
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))

	// Should fail when no plan exists
	result := v.CorrectAllChapters(1000)
	require.False(t, result)
}

func TestViter_CorrectAllChapters_Success(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Chapter 1", "First chapter content."),
			createValidCritiqueResponse(8),
			createValidChapterResponse(2, "Chapter 2", "Second chapter content."),
			createValidCritiqueResponse(8),
			createValidChapterResponse(3, "Chapter 3", "Third chapter content."),
			createValidCritiqueResponse(8),
			"Corrected first chapter content.",
			"Corrected second chapter content.",
			"Corrected third chapter content.",
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with chapters
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(3, 1))
	require.True(t, v.UpdateAllChapters(1))

	// Correct all chapters
	result := v.CorrectAllChapters(1000)
	require.True(t, result)

	// Verify all chapters were corrected
	book := v.GetBook()
	for i := 1; i <= 3; i++ {
		chp, err := book.GetChapter(i)
		require.NoError(t, err)
		require.Contains(t, chp.GetContent(), "Corrected")
	}
}

func TestViter_CorrectChapter_LargeText_SplitByLines(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()

	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Test Chapter", "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"),
			createValidCritiqueResponse(8),
		},
		infiniteMode:    true,
		defaultResponse: "Corrected text",
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with chapter containing large text
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(1, 1))
	require.True(t, v.UpdateChapter(1, 1))

	// Correct chapter with maxLen that forces splitting
	maxLen := 15 // Small enough to force splitting
	result := v.CorrectChapter(1, maxLen)
	if !result {
		t.Logf("CorrectChapter failed, tg.callCount: %d", tg.callCount)
	}
	require.True(t, result)

	// Verify chapter content was processed
	book := v.GetBook()
	chp, err := book.GetChapter(1)
	require.NoError(t, err)
	content := chp.GetContent()

	// Should contain corrected content
	require.Contains(t, content, "Corrected")
}

func TestViter_CorrectChapter_TextGeneratorFailure(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Test Chapter", "Test content to correct."),
			createValidCritiqueResponse(8),
			"Should not be called", // This shouldn't be reached due to shouldFail
		},
		shouldFail: false, // Start with success for setup
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with chapter
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(1, 1))

	// Setup the chapter successfully first
	require.True(t, v.UpdateChapter(1, 1))

	// Now enable failure for correction attempts
	tg.shouldFail = true

	// Correction should fail when text generator fails
	result := v.CorrectChapter(1, 1000)
	require.False(t, result)
}

func TestViter_CorrectChapter_EmptyContent(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Empty Chapter", ""),
			createValidCritiqueResponse(8),
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with empty chapter
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(1, 1))
	require.True(t, v.UpdateChapter(1, 1))

	// Correct empty chapter
	result := v.CorrectChapter(1, 1000)
	require.True(t, result)

	// Verify content was set
	book := v.GetBook()
	chp, err := book.GetChapter(1)
	require.NoError(t, err)
	require.Equal(t, "", chp.GetContent())
}

func TestViter_CorrectChapter_SingleLineLongerThanMaxLen(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()

	longLine := "This is a very long line that exceeds the maximum length limit and should be handled properly by the correction function"

	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Long Line Chapter", longLine),
			createValidCritiqueResponse(8),
		},
		infiniteMode:    true,
		defaultResponse: "Corrected content",
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with chapter containing single long line
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(1, 1))
	require.True(t, v.UpdateChapter(1, 1))

	// Correct chapter with maxLen smaller than the line length
	maxLen := 50 // Smaller than longLine length
	result := v.CorrectChapter(1, maxLen)
	require.True(t, result)

	// Verify correction was applied
	book := v.GetBook()
	chp, err := book.GetChapter(1)
	require.NoError(t, err)
	require.Equal(t, "Corrected content", chp.GetContent())
}

func TestViter_CorrectChapter_MultipleNewlines(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()

	contentWithMultipleNewlines := "Line 1\n\n\nLine 2\n\nLine 3\n\n\n\nLine 4"

	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Multiple Newlines", contentWithMultipleNewlines),
			createValidCritiqueResponse(8),
		},
		infiniteMode:    true,
		defaultResponse: "Corrected part",
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with chapter containing multiple newlines
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(1, 1))
	require.True(t, v.UpdateChapter(1, 1))

	// Correct chapter with maxLen that forces splitting
	maxLen := 25
	result := v.CorrectChapter(1, maxLen)
	require.True(t, result)

	// Verify content was corrected
	book := v.GetBook()
	chp, err := book.GetChapter(1)
	require.NoError(t, err)
	content := chp.GetContent()
	require.Contains(t, content, "Corrected")
}

func TestViter_CorrectChapterText_Bug_Isolation(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()

	// Simple case that should require multiple correction calls due to splitting
	// Test case that will expose the second loop bug
	// Create content that will trigger both loops
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Test Chapter", "Line1\nLine2\nLine3\nLine4"),
			createValidCritiqueResponse(8),
			"Corrected1", // First pass corrections
			"Corrected2",
			"Corrected3",
			"Corrected4", // Second pass corrections
			"Corrected5",
			"Corrected6",
			"Corrected7", // Extra responses for safety
			"Corrected8",
			"Corrected9",
			"Corrected10",
		},
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with chapter
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(1, 1))
	require.True(t, v.UpdateChapter(1, 1))

	// Record call count before correction
	callsBefore := tg.callCount
	t.Logf("Calls before correction: %d", callsBefore)

	// Get original content for comparison
	book := v.GetBook()
	chp, err := book.GetChapter(1)
	require.NoError(t, err)
	originalContent := chp.GetContent()
	t.Logf("Original content: %q", originalContent)
	t.Logf("Original content length: %d", len(originalContent))

	// Correct chapter with very small maxLen to force splitting
	maxLen := 10 // Very small to force multiple splits
	result := v.CorrectChapter(1, maxLen)

	callsAfter := tg.callCount
	correctionCalls := callsAfter - callsBefore
	t.Logf("Calls after correction: %d", callsAfter)
	t.Logf("Correction calls made: %d", correctionCalls)

	if !result {
		t.Logf("CorrectChapter failed")
	}

	// Get final content
	chp, err = book.GetChapter(1)
	require.NoError(t, err)
	finalContent := chp.GetContent()
	t.Logf("Final content: %q", finalContent)
	t.Logf("Final content length: %d", len(finalContent))

	require.True(t, result)
}

func TestViter_CorrectChapterText_SecondLoopBug(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()

	// Test case that will expose the second loop bug
	// Create content that will trigger both loops
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(8),
			createValidPlanResponse(),
			createValidCritiqueResponse(7),
			createValidChapterResponse(1, "Test Chapter", "A\nB\nC\nD\nE"),
			createValidCritiqueResponse(8),
		},
		infiniteMode:    true,
		defaultResponse: "Corrected",
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	require.True(t, ok)

	// Setup book with chapter
	fs := afero.NewMemMapFs()
	path := "/test/book"
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.UpdateMeta(1))
	require.True(t, v.UpdatePlan(1, 1))
	require.True(t, v.UpdateChapter(1, 1))

	// Get the chapter and verify initial content
	book := v.GetBook()
	chp, err := book.GetChapter(1)
	require.NoError(t, err)
	originalContent := chp.GetContent()
	t.Logf("Original content: %q (len=%d)", originalContent, len(originalContent))

	// Use maxLen that will cause splitting and trigger the buggy second loop
	maxLen := 3 // Very small to force multiple corrections
	result := v.CorrectChapter(1, maxLen)

	// This should fail due to the second loop bug
	t.Logf("Correction result: %v", result)
	t.Logf("Final call count: %d", tg.callCount)

	// After fixing the bugs, this should now succeed
	require.True(t, result, "Expected success after fixing second loop bugs")
}
