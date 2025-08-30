package viter

import (
	"github.com/tmc/langchaingo/llms"
)

// GenerationType represents different types of text generation operations
type GenerationType int

const (
	WriteMeta GenerationType = iota
	CritiqueMeta
	UpdateMeta
	WritePlan
	CritiquePlan
	UpdatePlan
	WriteChapter
	CritiqueChapter
	UpdateChapter
	CritiqueBook
	CorrectText
	generateTypeCount
)

// String returns the string representation of GenerationType
func (gt GenerationType) String() string {
	switch gt {
	case WriteMeta:
		return "WriteMeta"
	case CritiqueMeta:
		return "CritiqueMeta"
	case UpdateMeta:
		return "UpdateMeta"
	case WritePlan:
		return "WritePlan"
	case CritiquePlan:
		return "CritiquePlan"
	case UpdatePlan:
		return "UpdatePlan"
	case WriteChapter:
		return "WriteChapter"
	case CritiqueChapter:
		return "CritiqueChapter"
	case UpdateChapter:
		return "UpdateChapter"
	case CritiqueBook:
		return "CritiqueBook"
	case CorrectText:
		return "CorrectText"
	default:
		return "Unknown"
	}
}

// GeneratorManager manages text generators for different operations
type GeneratorManager struct {
	generators []TextGenerator
}

// NewGeneratorManager creates a new generator manager
func NewGeneratorManager(defaultGen TextGenerator) *GeneratorManager {
	gm := &GeneratorManager{
		generators: make([]TextGenerator, generateTypeCount),
	}

	for i := range generateTypeCount {
		gm.generators[i] = defaultGen
	}

	return gm
}

// SetGenerator sets a generator for a specific generation type
func (gm *GeneratorManager) SetGenerator(genType GenerationType, tg TextGenerator) {
	gm.generators[genType] = tg
}

// GetGenerator returns a generator for the specified type, or default if not found
func (gm *GeneratorManager) GetGenerator(genType GenerationType) TextGenerator {
	if genType >= generateTypeCount {
		return nil
	}
	return gm.generators[genType]
}

// GenerateText generates text using the appropriate generator for the given type
func (gm *GeneratorManager) GenerateText(genType GenerationType, messages []llms.MessageContent) (string, bool) {
	if genType >= generateTypeCount {
		return "", false
	}
	return gm.generators[genType].GenerateText(messages)
}

// ConfigureFromModels sets up generators based on config models
func (gm *GeneratorManager) ConfigureFromModels(models Models, tgs map[string]TextGenerator) {
	if models.Write != "" {
		if tg, exists := tgs[models.Write]; exists {
			gm.SetGenerator(WriteMeta, tg)
			gm.SetGenerator(WritePlan, tg)
			gm.SetGenerator(WriteChapter, tg)
		}
	}

	if models.Critique != "" {
		if tg, exists := tgs[models.Critique]; exists {
			gm.SetGenerator(CritiqueMeta, tg)
			gm.SetGenerator(CritiquePlan, tg)
			gm.SetGenerator(CritiqueChapter, tg)
			gm.SetGenerator(CritiqueBook, tg)
		}
	}

	if models.Update != "" {
		if tg, exists := tgs[models.Update]; exists {
			gm.SetGenerator(UpdateMeta, tg)
			gm.SetGenerator(UpdatePlan, tg)
			gm.SetGenerator(UpdateChapter, tg)
		}
	}

	if models.Correct != "" {
		if tg, exists := tgs[models.Correct]; exists {
			gm.SetGenerator(CorrectText, tg)
		}
	}

	gm.setIfExists(WriteMeta, models.WriteMeta, tgs)
	gm.setIfExists(CritiqueMeta, models.CritiqueMeta, tgs)
	gm.setIfExists(UpdateMeta, models.UpdateMeta, tgs)
	gm.setIfExists(WritePlan, models.WritePlan, tgs)
	gm.setIfExists(CritiquePlan, models.CritiquePlan, tgs)
	gm.setIfExists(UpdatePlan, models.UpdatePlan, tgs)
	gm.setIfExists(WriteChapter, models.WriteChapter, tgs)
	gm.setIfExists(CritiqueChapter, models.CritiqueChapter, tgs)
	gm.setIfExists(UpdateChapter, models.UpdateChapter, tgs)
	gm.setIfExists(CritiqueBook, models.CritiqueBook, tgs)
	gm.setIfExists(CorrectText, models.CorrectText, tgs)
}

// setIfExists sets a generator if the model name exists and the generator is available
func (gm *GeneratorManager) setIfExists(genType GenerationType, modelName string, tgs map[string]TextGenerator) {
	if modelName != "" {
		if tg, exists := tgs[modelName]; exists {
			gm.SetGenerator(genType, tg)
		}
	}
}
