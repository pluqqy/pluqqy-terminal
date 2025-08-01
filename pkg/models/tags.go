package models

import (
	"errors"
	"hash/fnv"
	"strings"
)

// Tag-related errors
var (
	ErrEmptyTagName        = errors.New("tag name cannot be empty")
	ErrTagNameTooLong      = errors.New("tag name cannot exceed 50 characters")
	ErrInvalidTagCharacter = errors.New("tag name contains invalid characters")
)

// Tag represents a tag with metadata
type Tag struct {
	Name        string `yaml:"name"`
	Color       string `yaml:"color,omitempty"`
	Description string `yaml:"description,omitempty"`
	Parent      string `yaml:"parent,omitempty"`
}

// TagRegistry holds all tag metadata
type TagRegistry struct {
	Tags []Tag `yaml:"tags"`
}

// DefaultColorPalette provides a curated set of colors for tags
// These colors are chosen for good contrast and accessibility
var DefaultColorPalette = []string{
	"#e74c3c", // red
	"#3498db", // blue
	"#2ecc71", // green
	"#f39c12", // orange
	"#9b59b6", // purple
	"#1abc9c", // turquoise
	"#34495e", // dark gray
	"#e67e22", // dark orange
	"#16a085", // dark turquoise
	"#8e44ad", // dark purple
	"#f1c40f", // yellow
	"#d35400", // pumpkin
	"#27ae60", // nephritis
	"#2980b9", // belize hole
	"#c0392b", // pomegranate
}

// GetColor returns the color for a tag, using the registry color if available
// or generating a consistent color from the tag name
func GetTagColor(tagName string, registryColor string) string {
	if registryColor != "" {
		return registryColor
	}
	
	// Generate consistent color from tag name using hash
	h := fnv.New32a()
	h.Write([]byte(strings.ToLower(tagName)))
	hash := h.Sum32()
	
	return DefaultColorPalette[int(hash)%len(DefaultColorPalette)]
}

// NormalizeTagName normalizes a tag name for consistency
func NormalizeTagName(name string) string {
	// Convert to lowercase and trim spaces
	normalized := strings.ToLower(strings.TrimSpace(name))
	
	// Replace spaces with hyphens
	normalized = strings.ReplaceAll(normalized, " ", "-")
	
	// Remove any invalid characters (keep only alphanumeric, hyphens, slashes for hierarchy)
	var result strings.Builder
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '/' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// ValidateTagName checks if a tag name is valid
func ValidateTagName(name string) error {
	if name == "" {
		return ErrEmptyTagName
	}
	
	if len(name) > 50 {
		return ErrTagNameTooLong
	}
	
	// Check for valid characters
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			(r >= '0' && r <= '9') || r == '-' || r == '/' || r == ' ') {
			return ErrInvalidTagCharacter
		}
	}
	
	return nil
}

// IsHierarchicalTag checks if a tag has a parent (contains /)
func IsHierarchicalTag(tagName string) bool {
	return strings.Contains(tagName, "/")
}

// GetTagParent returns the parent part of a hierarchical tag
func GetTagParent(tagName string) string {
	parts := strings.Split(tagName, "/")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "/")
	}
	return ""
}

// GetTagLeaf returns the leaf part of a hierarchical tag
func GetTagLeaf(tagName string) string {
	parts := strings.Split(tagName, "/")
	return parts[len(parts)-1]
}