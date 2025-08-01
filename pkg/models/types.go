package models

import (
	"fmt"
	"time"
)

// Component type constants
const (
	ComponentTypeContext = "contexts"
	ComponentTypePrompt  = "prompts"
	ComponentTypeRules   = "rules"
)

type Component struct {
	Path     string
	Type     string
	Content  string
	Modified time.Time
	Tags     []string `yaml:"tags,omitempty"`
}

type ComponentRef struct {
	Type  string `yaml:"type"`
	Path  string `yaml:"path"`
	Order int    `yaml:"order"`
}

type Pipeline struct {
	Name       string         `yaml:"name"`
	Path       string         `yaml:"-"`
	Components []ComponentRef `yaml:"components"`
	OutputPath string         `yaml:"output_path,omitempty"`
	Tags       []string       `yaml:"tags,omitempty"`
}

// Validate checks if the pipeline is valid
func (p *Pipeline) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("pipeline name cannot be empty")
	}
	
	if len(p.Components) == 0 {
		return fmt.Errorf("pipeline must have at least one component")
	}
	
	// Validate each component reference
	for i, comp := range p.Components {
		if comp.Type == "" {
			return fmt.Errorf("component %d: type cannot be empty", i+1)
		}
		
		// Validate component type
		switch comp.Type {
		case ComponentTypeContext, ComponentTypePrompt, ComponentTypeRules:
			// Valid type
		default:
			return fmt.Errorf("component %d: invalid type '%s', must be one of: %s, %s, %s", 
				i+1, comp.Type, ComponentTypeContext, ComponentTypePrompt, ComponentTypeRules)
		}
		
		if comp.Path == "" {
			return fmt.Errorf("component %d: path cannot be empty", i+1)
		}
		
		if comp.Order <= 0 {
			return fmt.Errorf("component %d: order must be greater than 0", i+1)
		}
	}
	
	// Check for duplicate order values
	orderMap := make(map[int]bool)
	for i, comp := range p.Components {
		if orderMap[comp.Order] {
			return fmt.Errorf("component %d: duplicate order value %d", i+1, comp.Order)
		}
		orderMap[comp.Order] = true
	}
	
	return nil
}