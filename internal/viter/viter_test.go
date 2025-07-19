package viter_test

import (
	"testing"

	"viter/internal/neural"
	"viter/internal/viter"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

// Mock implementations for testing
type mockPromptProvider struct {
	writeMetaPrompt    string
	critiqueMetaPrompt string
	updateMetaPrompt   string
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

type mockTextGenerator struct {
	responses  []string
	callCount  int
	shouldFail bool
}

func (m *mockTextGenerator) GenerateText(messages []llms.MessageContent) (string, bool) {
	if m.shouldFail {
		return "", false
	}

	if m.callCount >= len(m.responses) {
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
		writeMetaPrompt:    "Write meta prompt",
		critiqueMetaPrompt: "Critique meta prompt",
		updateMetaPrompt:   "Update meta prompt",
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
` + string(rune(score+'0'))
}

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

	// Verify the meta file was created
	exists, err := afero.Exists(fs, path+"/meta.md")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestViter_CreateBook_FileSystemError(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	// Use a read-only filesystem to simulate failure
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

	// Create a book first
	require.True(t, v.CreateBook(fs, path))

	// Now test loading it
	result := v.LoadBook(fs, path)

	require.True(t, result)
}

func TestViter_LoadBook_NoMetaFile(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/nonexistent/book"

	result := v.LoadBook(fs, path)

	require.False(t, result)
}

func TestViter_LoadBook_InvalidMetaFile(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create directory and invalid meta file
	err := fs.MkdirAll(path, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, path+"/meta.md", []byte("invalid meta content"), 0644)
	require.NoError(t, err)

	result := v.LoadBook(fs, path)

	require.True(t, result) // Should still succeed as MetaFromString handles invalid content gracefully
}

func TestViter_UpdateMeta_NoBook(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	result := v.UpdateMeta(8)

	require.False(t, result)
}

func TestViter_UpdateMeta_EmptyMeta_Success(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(9),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Create book with empty meta
	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(8)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // Should call twice: write meta + critique
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

	result := v.UpdateMeta(8)

	require.False(t, result)
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

	// Make text generator fail after first call
	tg.shouldFail = true
	tg.callCount = 1

	result := v.UpdateMeta(8)

	require.False(t, result)
}

func TestViter_UpdateMeta_IterativeImprovement(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),      // Initial write
			createValidCritiqueResponse(5), // First critique (score too low)
			createValidMetaResponse(),      // Update meta
			createValidCritiqueResponse(9), // Second critique (score good)
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(8)

	require.True(t, result)
	require.Equal(t, 4, tg.callCount) // Should iterate until score >= 8
}

func TestViter_UpdateMeta_NoImprovement(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),      // Initial write
			createValidCritiqueResponse(5), // First critique
			createValidMetaResponse(),      // Update meta
			createValidCritiqueResponse(4), // Second critique (worse score)
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(8)

	require.True(t, result) // Should return true when no improvement possible
	require.Equal(t, 4, tg.callCount)
}

func TestViter_UpdateMeta_AlreadyFilledMeta(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidCritiqueResponse(9),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	// Set a filled meta manually
	metaContent := createValidMetaResponse()
	err := afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)

	// Reload the book to get the filled meta
	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateMeta(8)

	require.True(t, result)
	require.Equal(t, 1, tg.callCount) // Should only critique, not write meta
}

func TestViter_UpdateMeta_ExistingCritique(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),      // Update meta
			createValidCritiqueResponse(9), // New critique
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	// Set a filled meta and existing critique
	metaContent := createValidMetaResponse()
	critiqueContent := createValidCritiqueResponse(5)
	err := afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, path+"/meta_critique.md", []byte(critiqueContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateMeta(8)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // Should update meta and critique
}

func TestViter_UpdateMeta_InvalidMetaResponse(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			"invalid meta format",
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(8)

	require.False(t, result) // Should fail due to invalid meta format
}

func TestViter_UpdateMeta_InvalidCritiqueResponse(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			"invalid critique format",
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(8)

	require.False(t, result) // Should fail due to invalid critique format
}

func TestViter_UpdateMeta_ZeroMinScore(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(0),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(0)

	require.True(t, result)
	require.Equal(t, 2, tg.callCount) // Should succeed immediately with score 0
}

func TestViter_UpdateMeta_HighMinScore(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(5),
			createValidMetaResponse(),
			createValidCritiqueResponse(7),
			createValidMetaResponse(),
			createValidCritiqueResponse(10),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(10)

	require.True(t, result)
	require.Equal(t, 6, tg.callCount) // Should iterate until score reaches 10
}

// Integration test combining CreateBook, LoadBook, and UpdateMeta
func TestViter_FullWorkflow(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(9),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	// Test complete workflow
	require.True(t, v.CreateBook(fs, path))
	require.True(t, v.LoadBook(fs, path))
	require.True(t, v.UpdateMeta(8))

	// Verify files were created
	exists, err := afero.Exists(fs, path+"/meta.md")
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = afero.Exists(fs, path+"/meta_critique.md")
	require.NoError(t, err)
	require.True(t, exists)
}

// Test edge cases and boundary conditions
func TestViter_UpdateMeta_NegativeMinScore(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(0),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(-1)

	require.True(t, result) // Should succeed with any score >= -1
}

func TestViter_MultipleOperationsOnSameInstance(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidMetaResponse(),
			createValidCritiqueResponse(9),
			createValidMetaResponse(),
			createValidCritiqueResponse(9),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path1 := "/test/book1"
	path2 := "/test/book2"

	// Test multiple book operations on same Viter instance
	require.True(t, v.CreateBook(fs, path1))
	require.True(t, v.UpdateMeta(8))

	// Reset text generator for second book
	tg.callCount = 2

	require.True(t, v.CreateBook(fs, path2))
	require.True(t, v.UpdateMeta(8))
}

func TestViter_UpdateMeta_EmptyCritiqueImprovements(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{
			createValidCritiqueResponse(9),
		},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	// Set a filled meta and critique with empty improvements
	metaContent := createValidMetaResponse()
	critiqueContent := `## Strengths
Good character development

## Improvements

## Impressions
Shows promise

## Score
9`
	err := afero.WriteFile(fs, path+"/meta.md", []byte(metaContent), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, path+"/meta_critique.md", []byte(critiqueContent), 0644)
	require.NoError(t, err)

	require.True(t, v.LoadBook(fs, path))

	result := v.UpdateMeta(8)

	require.True(t, result)           // Should succeed with high score after generating new critique
	require.Equal(t, 1, tg.callCount) // Should call text generator once to create new critique
}

func TestViter_UpdateMeta_MaxIterationSafety(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()

	// Create many responses to test that iteration eventually stops
	responses := make([]string, 20)
	for i := range 20 {
		if i%2 == 0 {
			responses[i] = createValidMetaResponse()
		} else {
			responses[i] = createValidCritiqueResponse(5) // Always low score
		}
	}

	tg := &mockTextGenerator{responses: responses}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(10) // Very high score that won't be reached

	require.True(t, result) // Should eventually return true when no improvement
	// Should stop iterating when score doesn't improve
}

func TestViter_CreateBook_EmptyPath(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := ""

	result := v.CreateBook(fs, path)

	require.True(t, result) // Should still work with empty path
}

func TestViter_LoadBook_EmptyPath(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := ""

	// Create book first
	require.True(t, v.CreateBook(fs, path))

	result := v.LoadBook(fs, path)

	require.True(t, result) // Should work with empty path if meta exists
}

func TestViter_UpdateMeta_TextGeneratorReturnsEmptyString(t *testing.T) {
	cfg := createTestConfig()
	pp := createTestPromptProvider()
	tg := &mockTextGenerator{
		responses: []string{""},
	}
	v, _ := viter.NewViter(cfg, pp, tg)

	fs := afero.NewMemMapFs()
	path := "/test/book"

	require.True(t, v.CreateBook(fs, path))

	result := v.UpdateMeta(8)

	require.False(t, result) // Should fail with empty response
}

func TestViter_NewViter_NilDependencies(t *testing.T) {
	cfg := createTestConfig()

	// Test with nil prompt provider
	v1, ok1 := viter.NewViter(cfg, nil, &mockTextGenerator{})
	require.True(t, ok1)
	require.NotNil(t, v1)

	// Test with nil text generator
	v2, ok2 := viter.NewViter(cfg, createTestPromptProvider(), nil)
	require.True(t, ok2)
	require.NotNil(t, v2)

	// Test with both nil
	v3, ok3 := viter.NewViter(cfg, nil, nil)
	require.True(t, ok3)
	require.NotNil(t, v3)
}
