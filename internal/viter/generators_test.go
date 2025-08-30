package viter

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// MockTextGenerator is a mock implementation of TextGenerator for testing
type MockTextGenerator struct {
	name string
}

func (m *MockTextGenerator) GenerateText(messages []llms.MessageContent) (string, bool) {
	return "Generated text from " + m.name, true
}

func TestGeneratorManager_SetAndGetGenerator(t *testing.T) {
	defaultGen := &MockTextGenerator{name: "default"}
	specificGen := &MockTextGenerator{name: "specific"}

	gm := NewGeneratorManager(defaultGen)

	// Test default generator
	gen := gm.GetGenerator(WriteMeta)
	if gen != defaultGen {
		t.Errorf("Expected default generator, got different generator")
	}

	// Set specific generator
	gm.SetGenerator(WriteMeta, specificGen)

	// Test specific generator
	gen = gm.GetGenerator(WriteMeta)
	if gen != specificGen {
		t.Errorf("Expected specific generator, got different generator")
	}

	// Test that other types still return default
	gen = gm.GetGenerator(CritiqueMeta)
	if gen != defaultGen {
		t.Errorf("Expected default generator for CritiqueMeta, got different generator")
	}
}

func TestGeneratorManager_GenerateText(t *testing.T) {
	defaultGen := &MockTextGenerator{name: "default"}
	specificGen := &MockTextGenerator{name: "specific"}

	gm := NewGeneratorManager(defaultGen)
	gm.SetGenerator(WriteMeta, specificGen)

	messages := []llms.MessageContent{}

	// Test specific generator
	text, ok := gm.GenerateText(WriteMeta, messages)
	if !ok {
		t.Errorf("Expected GenerateText to succeed")
	}
	if text != "Generated text from specific" {
		t.Errorf("Expected text from specific generator, got: %s", text)
	}

	// Test default generator fallback
	text, ok = gm.GenerateText(CritiqueMeta, messages)
	if !ok {
		t.Errorf("Expected GenerateText to succeed")
	}
	if text != "Generated text from default" {
		t.Errorf("Expected text from default generator, got: %s", text)
	}
}

func TestGeneratorManager_ConfigureFromModels(t *testing.T) {
	defaultGen := &MockTextGenerator{name: "default"}
	writeGen := &MockTextGenerator{name: "write"}
	critiqueGen := &MockTextGenerator{name: "critique"}
	updateGen := &MockTextGenerator{name: "update"}
	correctGen := &MockTextGenerator{name: "correct"}

	gm := NewGeneratorManager(defaultGen)

	tgs := map[string]TextGenerator{
		"write-model":    writeGen,
		"critique-model": critiqueGen,
		"update-model":   updateGen,
		"correct-model":  correctGen,
	}

	models := Models{
		WriteMeta:    "write-model",
		CritiqueMeta: "critique-model",
		UpdateMeta:   "update-model",
		CorrectText:  "correct-model",
		// Test legacy fallback
		Write:    "write-model",
		Critique: "critique-model",
		Update:   "update-model",
		Correct:  "correct-model",
	}

	gm.ConfigureFromModels(models, tgs)

	// Test specific generators
	if gm.GetGenerator(WriteMeta) != writeGen {
		t.Errorf("Expected write generator for WriteMeta")
	}
	if gm.GetGenerator(CritiqueMeta) != critiqueGen {
		t.Errorf("Expected critique generator for CritiqueMeta")
	}
	if gm.GetGenerator(UpdateMeta) != updateGen {
		t.Errorf("Expected update generator for UpdateMeta")
	}
	if gm.GetGenerator(CorrectText) != correctGen {
		t.Errorf("Expected correct generator for CorrectText")
	}

	// Test legacy fallback - WritePlan should use write generator since it's not specifically set
	if gm.GetGenerator(WritePlan) != writeGen {
		t.Errorf("Expected write generator for WritePlan (legacy fallback)")
	}
}

func TestGeneratorManager_LegacyCompatibility(t *testing.T) {
	defaultGen := &MockTextGenerator{name: "default"}
	writeGen := &MockTextGenerator{name: "write"}

	gm := NewGeneratorManager(defaultGen)

	tgs := map[string]TextGenerator{
		"write-model": writeGen,
	}

	// Test with only legacy configuration
	models := Models{
		Write: "write-model",
	}

	gm.ConfigureFromModels(models, tgs)

	// All write-related operations should use the write generator
	writeOperations := []GenerationType{
		WriteMeta,
		WritePlan,
		WriteChapter,
	}

	for _, op := range writeOperations {
		if gm.GetGenerator(op) != writeGen {
			t.Errorf("Expected write generator for %s (legacy compatibility)", op.String())
		}
	}

	// Non-write operations should still use default
	if gm.GetGenerator(CritiqueMeta) != defaultGen {
		t.Errorf("Expected default generator for CritiqueMeta")
	}
}
