package books_test

import (
	"testing"
	"viter/internal/books"

	"github.com/stretchr/testify/require"
)

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

	meta, err := books.MetaFromString(metaStr)
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
	meta := books.BookMeta{}
	meta.SetTitle("Blood and Badges")
	meta.SetStyle("Urban Fantasy")
	meta.SetGenres([]string{"Fantasy", "Mystery", "Urban"})
	meta.SetLogline("When the supernatural meets police procedure, unlikely alliances form.")
	meta.SetWorld("Modern city with hidden supernatural elements")
	meta.SetProtagonists([]books.Character{
		books.NewCharacter("Detective Sarah", "A cop who discovers the supernatural"),
		books.NewCharacter("Marcus", "A vampire trying to solve his own murder"),
	})
	meta.SetMinorCharacters([]books.Character{
		books.NewCharacter("Chief Williams", "Sarah's skeptical boss"),
	})
	meta.SetAntagonists([]books.Character{
		books.NewCharacter("The Syndicate Leader", "A powerful vampire controlling the city's underworld"),
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

func TestRoundTripMetadata(t *testing.T) {
	// Create original metadata
	original := books.BookMeta{}
	original.SetStyle("Space Opera")
	original.SetGenres([]string{"Science Fiction", "Adventure"})
	original.SetWorld("A galaxy far, far away")
	original.SetProtagonists([]books.Character{
		books.NewCharacter("Captain Nova", "A fearless space explorer"),
	})
	original.SetPlot("The quest to save the galaxy")

	// Convert to string and back
	str := original.String()
	parsed, err := books.MetaFromString(str)
	require.NoError(t, err)

	// Verify all data is preserved
	require.Equal(t, original.GetStyle(), parsed.GetStyle())
	require.Equal(t, original.GetGenres(), parsed.GetGenres())
	require.Equal(t, original.GetWorld(), parsed.GetWorld())
	require.Equal(t, original.GetPlot(), parsed.GetPlot())
	require.Len(t, parsed.GetProtagonists(), 1)
	require.Equal(t, "Captain Nova", parsed.GetProtagonists()[0].GetName())
}

func TestBookMetaMergedCopy(t *testing.T) {
	// Test merging with empty target
	original := &books.BookMeta{}
	original.SetStyle("Fantasy")
	original.SetGenres([]string{"Epic Fantasy", "Adventure"})
	original.SetLogline("A hero's journey")
	original.SetWorld("Middle Earth")
	original.SetProtagonists([]books.Character{
		books.NewCharacter("Frodo", "A hobbit"),
	})
	original.SetAntagonists([]books.Character{
		books.NewCharacter("Sauron", "Dark Lord"),
	})
	original.SetMinorCharacters([]books.Character{
		books.NewCharacter("Sam", "Loyal friend"),
	})
	original.SetPlot("The ring must be destroyed")
	original.SetTitle("The Lord of the Rings")

	other := &books.BookMeta{}
	other.SetStyle("Science Fiction")
	other.SetGenres([]string{"Space Opera"})
	other.SetLogline("A galactic adventure")
	other.SetWorld("Galaxy Far Far Away")
	other.SetProtagonists([]books.Character{
		books.NewCharacter("Luke", "Jedi Knight"),
	})
	other.SetAntagonists([]books.Character{
		books.NewCharacter("Vader", "Sith Lord"),
	})
	other.SetMinorCharacters([]books.Character{
		books.NewCharacter("Han", "Smuggler"),
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
	original := &books.BookMeta{}
	original.SetStyle("Fantasy")
	original.SetGenres([]string{"Epic Fantasy"})
	original.SetLogline("A hero's journey")
	original.SetWorld("Middle Earth")
	original.SetPlot("The ring must be destroyed")
	original.SetTitle("The Lord of the Rings")

	// Only some fields set in other
	other := &books.BookMeta{}
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
	original := &books.BookMeta{}
	original.SetStyle("Fantasy")
	original.SetTitle("Original Title")

	empty := &books.BookMeta{}
	merged := original.MergedCopy(empty)

	// Should be identical to original since other is empty
	require.Equal(t, original.GetStyle(), merged.GetStyle())
	require.Equal(t, original.GetTitle(), merged.GetTitle())

	// But should be different objects
	require.NotSame(t, original, merged)
}

func TestBookMetaMergedCopyNil(t *testing.T) {
	original := &books.BookMeta{}
	original.SetStyle("Fantasy")
	original.SetTitle("Original Title")

	// Test with nil other - should not panic and return copy of original
	merged := original.MergedCopy(nil)
	require.Equal(t, original.GetStyle(), merged.GetStyle())
	require.Equal(t, original.GetTitle(), merged.GetTitle())
	require.NotSame(t, original, merged)
}
