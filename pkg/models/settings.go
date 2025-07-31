package models

// Settings represents the application configuration
type Settings struct {
	Output OutputSettings `yaml:"output"`
	UI     UISettings     `yaml:"ui"`
	Editor EditorSettings `yaml:"editor"`
}

// OutputSettings controls pipeline output behavior
type OutputSettings struct {
	DefaultFilename string             `yaml:"default_filename"`
	ExportPath      string             `yaml:"export_path"`
	Formatting      FormattingSettings `yaml:"formatting"`
}

// FormattingSettings controls output formatting
type FormattingSettings struct {
	ShowHeadings bool              `yaml:"show_headings"`
	Headings     HeadingSettings   `yaml:"headings"`
}

// HeadingSettings allows customization of section headings
type HeadingSettings struct {
	Context string `yaml:"context"`
	Prompts string `yaml:"prompts"`
	Rules   string `yaml:"rules"`
}

// UISettings controls UI preferences
type UISettings struct {
	ShowPreview    bool   `yaml:"show_preview"`
	ComponentView  string `yaml:"component_view"` // "list" or "table"
}

// EditorSettings controls editor preferences
type EditorSettings struct {
	Command        string `yaml:"command"`
	PreferInternal bool   `yaml:"prefer_internal"`
}

// DefaultSettings returns the default configuration
func DefaultSettings() *Settings {
	return &Settings{
		Output: OutputSettings{
			DefaultFilename: "PLUQQY.md",
			ExportPath:      "./",
			Formatting: FormattingSettings{
				ShowHeadings: true,
				Headings: HeadingSettings{
					Context: "## CONTEXT",
					Prompts: "## PROMPTS",
					Rules:   "## IMPORTANT RULES",
				},
			},
		},
		UI: UISettings{
			ShowPreview:   true,
			ComponentView: "list",
		},
		Editor: EditorSettings{
			Command:        "",
			PreferInternal: false,
		},
	}
}