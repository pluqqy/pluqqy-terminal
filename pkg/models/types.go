package models

import "time"

type Component struct {
	Path     string
	Type     string
	Content  string
	Modified time.Time
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
}