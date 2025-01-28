package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Helmfile represents the partial structure of the Helmfile.yaml file.
type Helmfile struct {
	Path     string    `yaml:"-"`
	Releases []Release `yaml:"releases"`
}

// getHelmfileHelmChartAndVersion return the Helm chart and version for a release in a Helmfile.
func getHelmfileHelmChartAndVersion(helmfile, releaseName string) (string, string, error) {
	// Load data from Helmfile.yaml
	helmfileData, err := loadHelmfile(helmfile)
	if err != nil {
		return "", "", err
	}
	// Find the release in the Helmfile.
	for _, release := range helmfileData.Releases {
		if release.Name == releaseName {
			return release.Chart, release.Version, nil
		}
	}
	return "", "", fmt.Errorf("release '%s' not found in Helmfile '%s'", releaseName, helmfile)
}

// loadHelmfile loads a Helmfile from local disk.
func loadHelmfile(path string) (*Helmfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	helmfile := &Helmfile{Path: path}
	if err := yaml.Unmarshal(data, &helmfile); err != nil {
		return nil, err
	}
	return helmfile, nil
}
