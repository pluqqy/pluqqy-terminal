package testhelpers

import (
	"github.com/pluqqy/pluqqy-terminal/pkg/models"
)

// Note: MakeTestPipelineModel is now in builders.go to avoid duplication

// MakeTestSettings creates test settings with custom section order
func MakeTestSettings(order ...string) *models.Settings {
	sections := make([]models.Section, len(order))
	for i, typ := range order {
		sections[i] = models.Section{Type: typ}
	}
	return &models.Settings{
		Output: models.OutputSettings{
			Formatting: models.FormattingSettings{
				Sections: sections,
			},
		},
	}
}