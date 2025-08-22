package tui

import (
	"testing"
)

func TestSearchFilterHelper_ToggleArchivedFilter(t *testing.T) {
	sfh := NewSearchFilterHelper()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "add archived filter to empty query",
			input: "",
			want:  "status:archived",
		},
		{
			name:  "add archived filter to existing query",
			input: "api",
			want:  "api status:archived",
		},
		{
			name:  "remove archived filter from query",
			input: "api status:archived",
			want:  "api",
		},
		{
			name:  "remove archived filter when it's the only filter",
			input: "status:archived",
			want:  "",
		},
		{
			name:  "handle multiple spaces correctly",
			input: "api   status:archived   test",
			want:  "api test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sfh.ToggleArchivedFilter(tt.input)
			if got != tt.want {
				t.Errorf("ToggleArchivedFilter(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSearchFilterHelper_CycleTypeFilter(t *testing.T) {
	sfh := NewSearchFilterHelper()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "add pipelines filter to empty query",
			input: "",
			want:  "type:pipelines",
		},
		{
			name:  "cycle from pipelines to prompts",
			input: "type:pipelines",
			want:  "type:prompts",
		},
		{
			name:  "cycle from prompts to contexts",
			input: "type:prompts",
			want:  "type:contexts",
		},
		{
			name:  "cycle from contexts to rules",
			input: "type:contexts",
			want:  "type:rules",
		},
		{
			name:  "cycle from rules back to all (no filter)",
			input: "type:rules",
			want:  "",
		},
		{
			name:  "add type filter to existing query",
			input: "api",
			want:  "api type:pipelines",
		},
		{
			name:  "cycle type filter with other text",
			input: "api type:prompts test",
			want:  "api test type:contexts",
		},
		{
			name:  "handle query with archived filter",
			input: "status:archived type:pipelines",
			want:  "status:archived type:prompts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sfh.CycleTypeFilter(tt.input)
			if got != tt.want {
				t.Errorf("CycleTypeFilter(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSearchFilterHelper_ExtractTypeFilter(t *testing.T) {
	sfh := NewSearchFilterHelper()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "extract pipelines type",
			input: "type:pipelines",
			want:  "pipelines",
		},
		{
			name:  "extract from middle of query",
			input: "api type:prompts test",
			want:  "prompts",
		},
		{
			name:  "no type filter present",
			input: "api test",
			want:  "",
		},
		{
			name:  "multiple filters present",
			input: "status:archived type:rules tag:test",
			want:  "rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sfh.extractTypeFilter(tt.input)
			if got != tt.want {
				t.Errorf("extractTypeFilter(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSearchFilterHelper_CombinedFilters(t *testing.T) {
	sfh := NewSearchFilterHelper()

	// Test combining both filters
	query := ""
	
	// Add archived filter
	query = sfh.ToggleArchivedFilter(query)
	if query != "status:archived" {
		t.Errorf("Expected 'status:archived', got %q", query)
	}
	
	// Add type filter
	query = sfh.CycleTypeFilter(query)
	if query != "status:archived type:pipelines" {
		t.Errorf("Expected 'status:archived type:pipelines', got %q", query)
	}
	
	// Cycle type filter
	query = sfh.CycleTypeFilter(query)
	if query != "status:archived type:prompts" {
		t.Errorf("Expected 'status:archived type:prompts', got %q", query)
	}
	
	// Remove archived filter
	query = sfh.ToggleArchivedFilter(query)
	if query != "type:prompts" {
		t.Errorf("Expected 'type:prompts', got %q", query)
	}
	
	// Cycle through all types back to none
	query = sfh.CycleTypeFilter(query) // -> contexts
	query = sfh.CycleTypeFilter(query) // -> rules
	query = sfh.CycleTypeFilter(query) // -> all (empty)
	if query != "" {
		t.Errorf("Expected empty query, got %q", query)
	}
}

func TestSearchFilterHelper_CycleTypeFilterForComponents(t *testing.T) {
	sfh := NewSearchFilterHelper()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "add prompts filter to empty query",
			input: "",
			want:  "type:prompts",
		},
		{
			name:  "cycle from prompts to contexts",
			input: "type:prompts",
			want:  "type:contexts",
		},
		{
			name:  "cycle from contexts to rules",
			input: "type:contexts",
			want:  "type:rules",
		},
		{
			name:  "cycle from rules back to all (no filter)",
			input: "type:rules",
			want:  "",
		},
		{
			name:  "handle pipelines filter gracefully (skip to prompts)",
			input: "type:pipelines",
			want:  "type:prompts",
		},
		{
			name:  "add type filter to existing query",
			input: "api",
			want:  "api type:prompts",
		},
		{
			name:  "cycle type filter with other text",
			input: "api type:contexts test",
			want:  "api test type:rules",
		},
		{
			name:  "handle query with archived filter",
			input: "status:archived type:prompts",
			want:  "status:archived type:contexts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sfh.CycleTypeFilterForComponents(tt.input)
			if got != tt.want {
				t.Errorf("CycleTypeFilterForComponents(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}