package composer

import (
	"fmt"
	"strings"

	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// ComposeComponentWithSettings composes a single component using provided settings
func ComposeComponentWithSettings(component *models.Component, settings *models.Settings) (string, error) {
	if component == nil {
		return "", fmt.Errorf("cannot compose component: nil component provided")
	}

	var output strings.Builder

	// Find the section settings for this component type
	var sectionHeading string
	if settings.Output.Formatting.ShowHeadings {
		for _, section := range settings.Output.Formatting.Sections {
			if strings.ToLower(section.Type) == strings.ToLower(component.Type) ||
			   strings.ToLower(section.Type) == strings.ToLower(component.Type)+"s" {
				sectionHeading = section.Heading
				break
			}
		}
	}

	// Add section heading if found and enabled
	if sectionHeading != "" {
		output.WriteString(fmt.Sprintf("%s\n\n", sectionHeading))
	}

	// Add the component content
	output.WriteString(component.Content)
	if !strings.HasSuffix(component.Content, "\n") {
		output.WriteString("\n")
	}

	return output.String(), nil
}