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

// Doc returns the assembled, cleaned Manifest renders as a single document.
// It adds a source header comment followed by all the yaml renders separated by '---'.
func (m *Manifest) Doc() string {
	header := fmt.Sprintf("#:manifestus render{appName=%s, srcName=%s, srcType=%s}", m.AppName, m.SrcName, m.SrcType)
	docs := make([]string, 0)
	for _, render := range m.Renders {
		docs = append(docs, render.Doc())
	}
	return fmt.Sprintf("%s\n%s", header, strings.Join(docs, "\n---\n"))
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
