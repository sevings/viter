package books_test

import (
	"testing"

	"viter/internal/books"

	"github.com/stretchr/testify/require"
)

func TestDiffFromText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "lines with empty lines",
			input:    "line1\n\nline3",
			expected: "line1\n\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := books.DiffFromText(tt.input)
			require.NotNil(t, diff)
			require.Equal(t, tt.expected, diff.Text())
		})
	}
}

func TestDiffText(t *testing.T) {
	// Test that Text() reconstructs the original text correctly
	originalText := "first line\nsecond line\nthird line"
	diff := books.DiffFromText(originalText)
	require.Equal(t, originalText, diff.Text())
}

func TestDiffString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "<1></1>",
		},
		{
			name:     "single line",
			input:    "hello",
			expected: "<1>hello</1>",
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: "<1>line1</1>\n<2>line2</2>\n<3>line3</3>",
		},
		{
			name:     "lines with empty line",
			input:    "line1\n\nline3",
			expected: "<1>line1</1>\n<2></2>\n<3>line3</3>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := books.DiffFromText(tt.input)
			require.Equal(t, tt.expected, diff.String())
		})
	}
}

func TestDiffFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty tags",
			input:    "<1></1>",
			expected: "",
		},
		{
			name:     "single tagged line",
			input:    "<1>hello world</1>",
			expected: "hello world",
		},
		{
			name:     "multiple tagged lines",
			input:    "<1>line1</1>\n<2>line2</2>\n<3>line3</3>",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "tagged lines with empty content",
			input:    "<1>line1</1>\n<2></2>\n<3>line3</3>",
			expected: "line1\n\nline3",
		},
		{
			name:     "non-sequential indices",
			input:    "<1>first</1>\n<3>third</3>",
			expected: "first\n\nthird",
		},
		{
			name:     "newlines inside tags",
			input:    "<1>line1\nline2</1>\n<2>line3</2>",
			expected: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := books.DiffFromString(tt.input)
			require.NoError(t, err)
			require.NotNil(t, diff)
			require.Equal(t, tt.expected, diff.Text())
		})
	}
}

func TestDiffFromStringInvalidIndices(t *testing.T) {
	// Test that invalid indices are ignored, not cause errors
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "invalid index ignored",
			input:    "<abc>content</abc>",
			expected: "",
		},
		{
			name:     "mixed valid and invalid",
			input:    "<1>valid</1>\n<invalid>content</invalid>",
			expected: "valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := books.DiffFromString(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, diff.Text())
		})
	}
}

func TestDiffMerge(t *testing.T) {
	// Create first diff
	diff1 := books.DiffFromText("line1\nline2\nline3")

	// Create second diff
	diff2 := books.DiffFromText("new1\nnew2")

	// Merge diff2 into diff1
	diff1.Merge(diff2)

	// The result should have parts from diff2 replacing parts in diff1
	expected := "new1\nnew2\nline3"
	require.Equal(t, expected, diff1.Text())
}

func TestDiffMergeNil(t *testing.T) {
	diff := books.DiffFromText("original")

	diff.Merge(nil)
	require.Equal(t, "original", diff.Text())
}

func TestDiffRoundTrip(t *testing.T) {
	// Test that DiffFromText -> String -> DiffFromString -> Text preserves content
	originalText := "first line\nsecond line\nthird line"

	// Original -> Diff
	diff1 := books.DiffFromText(originalText)

	// Diff -> Tagged String
	taggedString := diff1.String()

	// Tagged String -> Diff
	diff2, err := books.DiffFromString(taggedString)
	require.NoError(t, err)

	// Diff -> Text (should match original)
	finalText := diff2.Text()
	require.Equal(t, originalText, finalText)
}

func TestDiffMergeComplexScenario(t *testing.T) {
	// Create base diff
	base := books.DiffFromText("line0\nline1\nline2\nline3\nline4")

	// Create patch that modifies some lines
	patch := books.DiffFromText("modified1\nmodified2")

	// Merge patch into base
	base.Merge(patch)

	// Expected result: first two lines replaced, rest unchanged
	expected := "modified1\nmodified2\nline2\nline3\nline4"
	require.Equal(t, expected, base.Text())
}
