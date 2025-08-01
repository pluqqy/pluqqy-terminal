package models

// Settings represents the application configuration
type Settings struct {
	Output OutputSettings `yaml:"output"`
}

// OutputSettings controls pipeline output behavior
type OutputSettings struct {
	DefaultFilename string             `yaml:"default_filename"`
	ExportPath      string             `yaml:"export_path"`
	OutputPath      string             `yaml:"output_path"`      // Directory for pipeline-generated output files
	Formatting      FormattingSettings `yaml:"formatting"`
}

// FormattingSettings controls output formatting
type FormattingSettings struct {
	ShowHeadings bool      `yaml:"show_headings"`
	Sections     []Section `yaml:"sections"`
}

// Section defines a component section with its type and heading
type Section struct {
	Type    string `yaml:"type"`
	Heading string `yaml:"heading"`
}



// DefaultSettings returns the default configuration
func DefaultSettings() *Settings {
	return &Settings{
		Output: OutputSettings{
			DefaultFilename: "PLUQQY.md",
			ExportPath:      "./",
			OutputPath:      "tmp/",
			Formatting: FormattingSettings{
				ShowHeadings: true,
				Sections: []Section{
					{Type: "rules", Heading: "## IMPORTANT RULES"},
					{Type: "contexts", Heading: "## CONTEXT"},
					{Type: "prompts", Heading: "## PROMPTS"},
				},
			},
		},
	}
}