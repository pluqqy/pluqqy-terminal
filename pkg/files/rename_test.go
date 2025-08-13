package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		want        string
	}{
		{
			name:        "simple name",
			displayName: "Test Component",
			want:        "test-component",
		},
		{
			name:        "name with apostrophe",
			displayName: "User's Profile",
			want:        "user-s-profile",
		},
		{
			name:        "name with special characters",
			displayName: "My Component #1!",
			want:        "my-component-1",
		},
		{
			name:        "name with multiple spaces",
			displayName: "Component   With    Spaces",
			want:        "component-with-spaces",
		},
		{
			name:        "name with leading/trailing spaces",
			displayName: "  Trimmed Name  ",
			want:        "trimmed-name",
		},
		{
			name:        "name with consecutive special chars",
			displayName: "Component!!!###Name",
			want:        "component-name",
		},
		{
			name:        "empty name",
			displayName: "",
			want:        "unnamed",
		},
		{
			name:        "only special characters",
			displayName: "!!!###",
			want:        "unnamed",
		},
		{
			name:        "mixed case",
			displayName: "CamelCaseComponent",
			want:        "camelcasecomponent",
		},
		{
			name:        "numbers and letters",
			displayName: "Component 123 Test",
			want:        "component-123-test",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.displayName)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.displayName, got, tt.want)
			}
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	// SanitizeFileName should be an alias for Slugify
	tests := []struct {
		name        string
		displayName string
		want        string
	}{
		{
			name:        "simple name",
			displayName: "Test Component",
			want:        "test-component",
		},
		{
			name:        "complex name",
			displayName: "User's Profile #1",
			want:        "user-s-profile-1",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFileName(tt.displayName)
			slugified := Slugify(tt.displayName)
			
			if got != tt.want {
				t.Errorf("SanitizeFileName(%q) = %q, want %q", tt.displayName, got, tt.want)
			}
			
			if got != slugified {
				t.Errorf("SanitizeFileName and Slugify should return same result, got %q vs %q", got, slugified)
			}
		})
	}
}

func TestExtractDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "markdown file",
			filename: "auth-context.md",
			want:     "Auth Context",
		},
		{
			name:     "yaml file",
			filename: "users-profile.yaml",
			want:     "Users Profile",
		},
		{
			name:     "single word",
			filename: "component.md",
			want:     "Component",
		},
		{
			name:     "multiple hyphens",
			filename: "my-test-component.md",
			want:     "My Test Component",
		},
		{
			name:     "no extension",
			filename: "test-file",
			want:     "Test File",
		},
		{
			name:     "empty filename",
			filename: "",
			want:     "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractDisplayName(tt.filename)
			if got != tt.want {
				t.Errorf("ExtractDisplayName(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestExtractMarkdownDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "simple header",
			content: `# Test Component

Some content here`,
			want: "Test Component",
		},
		{
			name: "header after frontmatter",
			content: `---
tags: [test]
---

# My Component

Content`,
			want: "My Component",
		},
		{
			name: "no header",
			content: `Just some content
without any header`,
			want: "",
		},
		{
			name: "header with extra spaces",
			content: `#   Spaced Header   

Content`,
			want: "Spaced Header",
		},
		{
			name: "subheader ignored",
			content: `## This is H2
# This is H1

Content`,
			want: "This is H1",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractMarkdownDisplayName(tt.content)
			if got != tt.want {
				t.Errorf("ExtractMarkdownDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdateMarkdownDisplayName(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		newDisplayName string
		wantContains   string
	}{
		{
			name: "update existing header",
			content: `# Old Name

Some content`,
			newDisplayName: "New Name",
			wantContains:   "# New Name",
		},
		{
			name: "add header when missing",
			content: `Just content without header`,
			newDisplayName: "New Header",
			wantContains:   "# New Header",
		},
		{
			name: "update header after frontmatter",
			content: `---
tags: [test]
---

# Old Header

Content`,
			newDisplayName: "Updated Header",
			wantContains:   "# Updated Header",
		},
		{
			name: "add header after frontmatter when missing",
			content: `---
tags: [test]
---

Content without header`,
			newDisplayName: "Added Header",
			wantContains:   "# Added Header",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateMarkdownDisplayName(tt.content, tt.newDisplayName)
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("updateMarkdownDisplayName() result missing %q\nGot:\n%s", tt.wantContains, got)
			}
		})
	}
}

func TestValidateRename(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "rename-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test file
	existingFile := filepath.Join(tempDir, "existing.md")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name           string
		oldPath        string
		newDisplayName string
		itemType       string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid rename",
			oldPath:        "components/test.md",
			newDisplayName: "New Component",
			itemType:       "component",
			wantErr:        false,
		},
		{
			name:           "empty name",
			oldPath:        "components/test.md",
			newDisplayName: "",
			itemType:       "component",
			wantErr:        true,
			errContains:    "empty",
		},
		{
			name:           "only special characters becomes unnamed",
			oldPath:        "components/test.md",
			newDisplayName: "!!!###",
			itemType:       "component",
			wantErr:        false, // Gets converted to "unnamed", which is valid
		},
		{
			name:           "same name allowed",
			oldPath:        "components/test.md",
			newDisplayName: "Test",
			itemType:       "component",
			wantErr:        false, // Same name is allowed (no-op)
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to temp directory for file checks
			oldCwd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldCwd)
			
			err := ValidateRename(tt.oldPath, tt.newDisplayName, tt.itemType)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid component path",
			path:    "components/contexts/test.md",
			wantErr: false,
		},
		{
			name:    "valid pipeline path",
			path:    "test-pipeline.yaml",
			wantErr: false,
		},
		{
			name:    "path with ..",
			path:    "../components/test.md",
			wantErr: true,
		},
		{
			name:    "absolute path is allowed",
			path:    "/absolute/path/test.md",
			wantErr: false, // validatePath doesn't reject absolute paths
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}