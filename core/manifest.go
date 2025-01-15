package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Manifest represents a manifest file for an app name and src type.
type Manifest struct {
	AppName string
	SrcName string
	SrcType string
	Renders Renders
}

// NewManifest returns a new Manifest for app name and renderer src type.
func NewManifest(appName, srcName, srcType string, renders Renders) *Manifest {
	return &Manifest{
		AppName: appName,
		SrcName: srcName,
		SrcType: srcType,
		Renders: renders,
	}
}

// Doc returns the assembled, cleaned Manifest renders as a string
func (m *Manifest) Doc() string {
	doc := fmt.Sprintf("# %s %s manifest rendered by manifestus\n", m.AppName, m.SrcType)
	parts := make([]string, 0)
	parts = append(parts, doc)
	for _, render := range m.Renders {
		parts = append(parts, render.Doc())
	}
	return strings.Join(parts, fmt.Sprintf("\n%s\n", yamlSep))
}

// Write writes the manifest to a file in the output directory.
func (m *Manifest) Write(outputDir string) (string, error) {
	p := path.Join(outputDir, getOutputFilePath(m.AppName, m.SrcName, m.SrcType))
	if p, err := filepath.Abs(p); err != nil {
		return p, err
	}
	if err := os.MkdirAll(path.Dir(p), 0755); err != nil {
		return p, err
	}
	err := os.WriteFile(p, []byte(m.Doc()), 0644)
	return p, err
}

// yamlSep is the delimiter for separating YAML documents in a multi-document YAML file.
const yamlSep = "---"
