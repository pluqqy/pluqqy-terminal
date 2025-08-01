package models

import (
	"testing"
)

func TestNormalizeTagName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "API", "api"},
		{"trim spaces", "  api  ", "api"},
		{"replace spaces", "api endpoint", "api-endpoint"},
		{"remove invalid chars", "api@endpoint!", "apiendpoint"},
		{"keep hyphens", "api-endpoint", "api-endpoint"},
		{"keep slashes", "project/frontend", "project/frontend"},
		{"mixed case with spaces", "API Endpoint", "api-endpoint"},
		{"numbers allowed", "api-v2", "api-v2"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeTagName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeTagName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateTagName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errType error
	}{
		{"valid simple", "api", false, nil},
		{"valid with hyphen", "api-endpoint", false, nil},
		{"valid with slash", "project/frontend", false, nil},
		{"valid with numbers", "api-v2", false, nil},
		{"empty string", "", true, ErrEmptyTagName},
		{"too long", "this-is-a-very-long-tag-name-that-exceeds-fifty-characters-limit", true, ErrTagNameTooLong},
		{"valid with spaces", "api endpoint", false, nil}, // spaces are allowed, will be normalized
		{"invalid chars", "api@endpoint", true, ErrInvalidTagCharacter},
		{"special chars", "api#endpoint!", true, ErrInvalidTagCharacter},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTagName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTagName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr && err != tt.errType {
				t.Errorf("ValidateTagName(%q) error = %v, want %v", tt.input, err, tt.errType)
			}
		})
	}
}

func TestGetTagColor(t *testing.T) {
	tests := []struct {
		name         string
		tagName      string
		registryColor string
		expectCustom  bool
	}{
		{"use registry color", "api", "#custom", true},
		{"auto assign color", "api", "", false},
		{"consistent color", "api", "", false},
		{"empty registry color", "api", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := GetTagColor(tt.tagName, tt.registryColor)
			
			if tt.expectCustom && color != tt.registryColor {
				t.Errorf("GetTagColor(%q, %q) = %q, want %q", tt.tagName, tt.registryColor, color, tt.registryColor)
			}
			
			if !tt.expectCustom {
				// Check if color is from default palette
				found := false
				for _, c := range DefaultColorPalette {
					if c == color {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetTagColor(%q, %q) = %q, not in default palette", tt.tagName, tt.registryColor, color)
				}
				
				// Check consistency - same tag should always get same color
				color2 := GetTagColor(tt.tagName, tt.registryColor)
				if color != color2 {
					t.Errorf("GetTagColor(%q, %q) not consistent: %q != %q", tt.tagName, tt.registryColor, color, color2)
				}
			}
		})
	}
}

func TestIsHierarchicalTag(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected bool
	}{
		{"simple tag", "api", false},
		{"hierarchical tag", "project/frontend", true},
		{"deep hierarchy", "project/web/frontend", true},
		{"no hierarchy", "api-endpoint", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHierarchicalTag(tt.tagName)
			if result != tt.expected {
				t.Errorf("IsHierarchicalTag(%q) = %v, want %v", tt.tagName, result, tt.expected)
			}
		})
	}
}

func TestGetTagParent(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected string
	}{
		{"simple tag", "api", ""},
		{"hierarchical tag", "project/frontend", "project"},
		{"deep hierarchy", "project/web/frontend", "project/web"},
		{"no hierarchy", "api-endpoint", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTagParent(tt.tagName)
			if result != tt.expected {
				t.Errorf("GetTagParent(%q) = %q, want %q", tt.tagName, result, tt.expected)
			}
		})
	}
}

func TestGetTagLeaf(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected string
	}{
		{"simple tag", "api", "api"},
		{"hierarchical tag", "project/frontend", "frontend"},
		{"deep hierarchy", "project/web/frontend", "frontend"},
		{"no hierarchy", "api-endpoint", "api-endpoint"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTagLeaf(tt.tagName)
			if result != tt.expected {
				t.Errorf("GetTagLeaf(%q) = %q, want %q", tt.tagName, result, tt.expected)
			}
		})
	}
}