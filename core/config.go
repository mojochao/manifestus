package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// configDir is the directory where the config was loaded from.
// It is used to resolve relative paths in the config, output files, and the
// Helmfile used by the 'helmfile template' command to render Helm releases.
var configDir string

// LoadConfig loads a YAML config file from the given path into a Config struct and returns it.
func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed to close config: %v", err)
		}
	}(file)

	// Cache the directory of the config file for resolving relative paths used elsewhere.
	configDir = filepath.Dir(filePath)

	// Decode the YAML config file into a Config and return it.
	decoder := yaml.NewDecoder(file)
	config := Config{}
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode YAML from config: %w", err)
	}
	config.Path = filePath
	return &config, nil
}

// Config represents the structure of the YAML config file, typically named 'renderfile.yaml'.
type Config struct {
	Path       string     `yaml:"-"`
	Renderfile Renderfile `yaml:"renderfile"`
}

// EnabledApps returns the App objects in the config not disabled sorted by
// name for consistent ordering of output and processing.
func (c *Config) EnabledApps() []*App {
	enabled := make([]*App, 0)
	for _, app := range c.Renderfile.Apps {
		if app.Disabled {
			continue
		}
		enabled = append(enabled, &app)
	}
	sort.Slice(enabled, func(i, j int) bool {
		return enabled[i].Name < enabled[j].Name
	})
	return enabled
}

// EnabledAppNames returns the names of the enabled apps in the config.
func (c *Config) EnabledAppNames() []string {
	enabled := c.EnabledApps()
	names := make([]string, len(enabled))
	for i, app := range enabled {
		names[i] = app.Name
	}
	sort.Strings(names)
	return names
}

// FindApp returns the enabled app in the config with the given name.
func (c *Config) FindApp(appName string) *App {
	for _, app := range c.Renderfile.Apps {
		if app.Name == appName {
			return &app
		}
	}
	return nil
}

// MissingApps returns a list of app names that are not in the config.
func (c *Config) MissingApps(appNames []string) []string {
	missing := make([]string, 0)
	for _, appName := range appNames {
		if c.FindApp(appName) == nil {
			missing = append(missing, appName)
		}
	}
	return missing
}

// Renderfile represents the structure of the top-level '.renderfile' section of the config.
type Renderfile struct {
	Schema string `yaml:"schema"`
	Apps   []App  `yaml:"apps"`
}

// App represents the structure of an app in '.manifestus.apps' section of the config.
type App struct {
	Name           string          `yaml:"name"`
	Disabled       bool            `yaml:"disabled"`
	Releases       []Release       `yaml:"releases"`
	Kustomizations []Kustomization `yaml:"kustomizations"`
	Bundles        []Bundle        `yaml:"bundles"`
	CRDs           []CRD           `yaml:"crds"`
}

// Release represents the structure of a Helm chart release in '.manifestus.apps.*.releases' section of the config.
// The only required field is 'name'.
//
// If only a 'name' is provided, the release will be rendered with 'helmfile template'
// command using the Helmfile in the 'helmfile' field if provided. If not provided 'the helmfile.yaml'
// in the same directory as the config file is used.
//
// If 'chart' is provided, the release will be rendered with 'helm template' command using the
// Helm chart in the 'chart' field at 'version' level into the namespace in the 'namespace' field
// with the 'values' file in the 'values' field.
type Release struct {
	// Name is the Helm release name.
	Name string `yaml:"name"`

	// Namespace is the Kubernetes namespace the release will be installed into.
	Namespace string `yaml:"namespace"`

	// Helmfile is the path to the Helmfile used to render the release with 'helmfile template' command.
	Helmfile string `yaml:"helmfile"`

	// Chart is the Helm chart name used to render the release with 'helm template' command.
	Chart string `yaml:"chart"`

	// Version is the Helm chart version used to render the release with 'helm template' command.
	Version string `yaml:"version"`

	// Values is the path to the values file used to render the release with 'helm template' command.
	Values string `yaml:"values"`
}

// Kustomization represents the structure of a kustomization in '.manifestus.apps.*.kustomizations' section of the config.
type Kustomization struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

// Bundle represents the structure of a bundle in '.manifestus.apps.*.bundles' section of the config.
type Bundle struct {
	Name    string            `yaml:"name"`
	Data    map[string]string `yaml:"data"`
	Sources []string          `yaml:"sources"`
}

// Paths returns filesystem paths in a bundle with {placeholders} replaced by values from the bundle's data.
func (b Bundle) Paths() ([]string, error) {
	paths := make([]string, 0)
	for _, source := range b.Sources {
		if isURL(source) {
			continue
		}
		path, err := expandTemplate(source, b.Data)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// URLs returns filesystem paths in a bundle with {placeholders} replaced by values from the bundle's data.
func (b Bundle) URLs() ([]string, error) {
	urls := make([]string, 0)
	for _, source := range b.Sources {
		if !isURL(source) {
			continue
		}
		url, err := expandTemplate(source, b.Data)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

// CRD represents the structure of a CRD in '.manifestus.apps.*.crds' section of the config.
type CRD struct {
	Name    string            `yaml:"name"`
	Data    map[string]string `yaml:"data"`
	Sources []string          `yaml:"sources"`
}

// Paths returns filesystem paths in a CRD with {placeholders} replaced by values from the CRD's data.
func (c CRD) Paths() ([]string, error) {
	paths := make([]string, 0)
	for _, source := range c.Sources {
		if isURL(source) {
			continue
		}
		path, err := expandTemplate(source, c.Data)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// URLs returns filesystem paths in a CRD with {placeholders} replaced by values from the CRD's data.
func (c CRD) URLs() ([]string, error) {
	urls := make([]string, 0)
	for _, source := range c.Sources {
		if !isURL(source) {
			continue
		}
		url, err := expandTemplate(source, c.Data)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}
