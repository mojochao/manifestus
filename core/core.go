package core

import (
	"fmt"
	"path"
	"sort"
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
func GetOutputFiles(cfg *Config, appNames, srcNames, srcTypes []string, flatten bool) []string {
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
				paths = append(paths, getOutputFilePath(app.Name, release.Name, "release", flatten))
			}
		}
		if contains(srcTypes, "kustomization") {
			for _, kustomization := range app.Kustomizations {
				if len(srcNames) > 0 && !contains(srcNames, kustomization.Name) {
					continue
				}
				paths = append(paths, getOutputFilePath(app.Name, kustomization.Name, "kustomization", flatten))
			}
		}
		if contains(srcTypes, "bundle") {
			for _, bundle := range app.Bundles {
				if len(srcNames) > 0 && !contains(srcNames, bundle.Name) {
					continue
				}
				paths = append(paths, getOutputFilePath(app.Name, bundle.Name, "bundle", flatten))
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
	keys := make([]descriptor, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].appName != keys[j].appName {
			return keys[i].appName < keys[j].appName
		}
		if keys[i].srcName != keys[j].srcName {
			return keys[i].srcName < keys[j].srcName
		}
		return keys[i].srcType < keys[j].srcType
	})

	for _, key := range keys {
		appRenders := seen[key]
		manifest := &Manifest{
			AppName: key.appName,
			SrcName: key.srcName,
			SrcType: key.srcType,
			Renders: appRenders,
		}
		manifestList = append(manifestList, manifest)
	}
	return manifestList
}

// getOutputFilePath returns the path of an output file of an app rendered manifest.
func getOutputFilePath(appName, srcName, srcType string, flatten bool) string {
	if !flatten {
		return path.Join(appName, fmt.Sprintf("%s.%s.manifest.yaml", srcName, srcType))
	}
	return fmt.Sprintf("%s-%s.%s.manifest.yaml", appName, srcName, srcType)
}

// Chart represents metadata of a Helm chart defined in a Renderfile or Helmfile.
type Chart struct {
	Name    string
	Version string
	App     string
}

// LatestVersion returns the latest version of the Helm chart.
func (c Chart) LatestVersion() (string, error) {
	cmdline := fmt.Sprintf("helm search repo %s --max 1 | awk 'NR==2 {print $2}'", c.Name)
	_, stdout, stderr, exit, err := execCmd(cmdline, "")
	if err != nil {
		return string(stderr), err
	}
	if exit != 0 {
		return string(stderr), fmt.Errorf("failed to get latest version for chart '%s'", c.Name)
	}
	return string(stdout), nil
}

func GetCharts(cfg *Config, appNames []string) ([]*Chart, error) {
	results := make([]*Chart, 0)
	for _, appName := range appNames {
		app := cfg.FindApp(appName)
		var chartInfo *Chart
		for _, release := range app.Releases {
			if release.Chart != "" {
				chartInfo = &Chart{
					Name:    release.Chart,
					Version: release.Version,
					App:     appName,
				}
			} else {
				chart, version, err := getHelmfileHelmChartAndVersion(release.Helmfile, release.Name)
				if err != nil {
					return nil, err
				}
				chartInfo = &Chart{
					Name:    chart,
					Version: version,
					App:     appName,
				}
			}
			results = append(results, chartInfo)
		}
	}
	return results, nil
}

// ExecHelmRepoUpdate executes the 'helm repo update' command.
func ExecHelmRepoUpdate() error {
	cmdline := "helm repo update"
	_, _, _, exit, err := execCmd(cmdline, "")
	if err != nil {
		return fmt.Errorf("failed to update Helm repositories: error=%w", err)
	}
	if exit != 0 {
		return fmt.Errorf("failed to update Helm repositories: exit=%d", exit)
	}
	return nil
}
