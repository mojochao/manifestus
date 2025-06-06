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

// Doc returns the assembled, cleaned Manifest renders as a single document.
// It adds a source header comment followed by all the yaml renders separated by '---'.
func (m *Manifest) Doc(noBanner bool) string {
	var header string
	if !noBanner {
		header = fmt.Sprintf("#:manifestus render{appName=%s, srcName=%s, srcType=%s}", m.AppName, m.SrcName, m.SrcType)
	}

	docs := make([]string, 0)
	for _, render := range m.Renders {
		docs = append(docs, render.Doc())
	}
	return fmt.Sprintf("%s\n%s", header, strings.Join(docs, "\n---\n"))
}

// Write writes the manifest to a file in the output directory.
func (m *Manifest) Write(outputDir string, flatten bool, noBanner bool) (string, error) {
	p := path.Join(outputDir, getOutputFilePath(m.AppName, m.SrcName, m.SrcType, flatten))
	if p, err := filepath.Abs(p); err != nil {
		return p, err
	}
	if err := os.MkdirAll(path.Dir(p), 0755); err != nil {
		return p, err
	}
	err := os.WriteFile(p, []byte(m.Doc(noBanner)), 0644)
	return p, err
}
