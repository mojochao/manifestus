package core

import (
	"fmt"
	"path"
)

// EnsureAppNamesExist checks if the given app names exist in the Config.
func EnsureAppNamesExist(cfg *Config, appNames []string) error {
	for _, appName := range appNames {
		if cfg.FindApp(appName) == nil {
			return fmt.Errorf("app '%s' not found", appName)
		}
	}
	return nil
}

// EnsureSrcTypesValid checks if the given source types are valid.
func EnsureSrcTypesValid(srcTypes []string) error {
	for _, srcType := range srcTypes {
		if !contains(ValidSrcTypes, srcType) {
			return fmt.Errorf("invalid source type '%s'", srcType)
		}
	}
	return nil
}

// GetOutputFiles returns a list of output files of rendered manifests for named apps in the Config.
func GetOutputFiles(cfg *Config, appNames, srcNames, srcTypes []string) []string {
	paths := make([]string, 0)
	for _, appName := range appNames {
		app := cfg.FindApp(appName)
		if app.Disabled {
			continue
		}
		if contains(srcTypes, "release") {
			for _, release := range app.Releases {
				if len(srcNames) > 0 && !contains(srcNames, release.Name) {
					continue
				}
				paths = append(paths, getOutputFilePath(app.Name, release.Name, "release"))
			}
		}
		if contains(srcTypes, "kustomization") {
			for _, kustomization := range app.Kustomizations {
				if len(srcNames) > 0 && !contains(srcNames, kustomization.Name) {
					continue
				}
				paths = append(paths, getOutputFilePath(app.Name, kustomization.Name, "kustomization"))
			}
		}
		if contains(srcTypes, "bundle") {
			for _, bundle := range app.Bundles {
				if len(srcNames) > 0 && !contains(srcNames, bundle.Name) {
					continue
				}
				paths = append(paths, getOutputFilePath(app.Name, bundle.Name, "bundle"))
			}
		}
	}
	return paths
}

// GetRenders returns a list of rendered manifests for named apps in the Config.
func GetRenders(cfg *Config, appNames, srcNames, srcTypes []string, debug, dryRun bool) ([]*Render, error) {
	results := make([]*Render, 0)
	for _, appName := range appNames {
		app := cfg.FindApp(appName)
		renders, err := getRendersForApp(app, srcNames, srcTypes, debug, dryRun)
		if err != nil {
			return nil, err
		}
		results = append(results, renders...)
	}
	return results, nil
}

// GetManifests returns a list of manifests from a list of renders.
func GetManifests(renders []*Render) []*Manifest {
	// Map renders to manifests by app name, source name, and source type.
	type descriptor struct {
		appName string
		srcName string
		srcType string
	}
	seen := make(map[descriptor][]*Render)
	for _, render := range renders {
		key := descriptor{render.AppName, render.SrcName, render.SrcType}
		if _, ok := seen[key]; !ok {
			seen[key] = make([]*Render, 0)
		}
		seen[key] = append(seen[key], render)
	}

	// Return manifests from the mapped renders.
	manifestList := make([]*Manifest, 0)
	for key, renders := range seen {
		manifest := NewManifest(key.appName, key.srcName, key.srcType, renders)
		manifestList = append(manifestList, manifest)
	}
	return manifestList
}

// getOutputFilePath returns the path of an output file of an app rendered manifest.
func getOutputFilePath(appName, srcName, srcType string) string {
	return path.Join(appName, fmt.Sprintf("%s.%s.manifest.yaml", srcName, srcType))
}
