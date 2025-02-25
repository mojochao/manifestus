package core

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
)

// ValidSrcTypes is a list of valid source types.
var ValidSrcTypes = []string{"release", "kustomization", "bundle"}

// Renders is a collection of Render objects containing states of all attempted
// renders for an App from its configured sources and their renderers.
type Renders []*Render

// Render represents a rendered document for an App from some source.
// The source can be a Helm or Helmfile Release, a Kustomization, or a Bundle.
type Render struct {
	// AppName is the name of the App that the Render is for.
	AppName string

	// SrcName is the name of the source of the Render.
	// For Release objects, this is the Release name.
	// For Kustomization objects, this is the Kustomization name.
	// For Bundle objects, this is the Bundle name.
	SrcName string

	// SrcType is the type of the source of the Render. One of "release", "kustomization", or "bundle".
	SrcType string

	// CmdLine is the command line that was executed to render the document.
	CmdLine string

	// Cmd is the command that was executed to render the document, if any.
	// Static manifests are copied directly to the output directory using just
	// the standard library, so this field will be nil for those, however those
	// standard library calls may still return an error in the Err field.
	Cmd *exec.Cmd

	// Stdout is the standard output, if any, of the command that was executed to render the document.
	// For static manifests, this is the contents of the file.
	Stdout []byte

	// Stderr is the standard error output, if any,of the command that was executed to render the document.
	// For static manifests, this is nil.
	Stderr []byte

	// Err is the error that occurred during rendering, if any.
	Err error
}

// Doc returns the normalized, rendered manifest output as a string trimming leading and trailing whitespace.
func (r Render) Doc() string {
	return strings.TrimSpace(string(r.Stdout))
}

// Docs returns normalized, rendered manifests documents from render command output in stdout.
func (r Render) Docs() []string {
	docs := make([]string, 0)
	for _, doc := range docs {
		docs = append(docs, strings.TrimSpace(doc))
	}
	return docs
}

// Msg returns any render command output in stderr, which may or may not be related to an error.
// The stderr stream is also used for info output about the data being written to stdout by the render command.
func (r Render) Msg() string {
	return strings.TrimSpace(string(r.Stderr))
}

// helmfileName is the default name of a Helmfile.
const helmfileName = "helmfile.yaml"

// getRendersForApp returns a list of rendered manifests for a named app in the Config.
func getRendersForApp(app *App, srcNames, srcTypes []string, debug, dryRun bool) (Renders, error) {
	results := make([]*Render, 0)
	if contains(srcTypes, "release") {
		for _, release := range app.Releases {
			if len(srcNames) > 0 && !contains(srcNames, release.Name) {
				continue
			}
			render, err := renderRelease(app.Name, release, debug, dryRun)
			if err != nil {
				return nil, err
			}
			results = append(results, render)
		}
	}
	if contains(srcTypes, "kustomization") {
		for _, kustomization := range app.Kustomizations {
			if len(srcNames) > 0 && !contains(srcNames, kustomization.Name) {
				continue
			}
			render, err := renderKustomization(app.Name, kustomization, dryRun)
			if err != nil {
				return nil, err
			}
			results = append(results, render)
		}
	}
	if contains(srcTypes, "bundle") {
		for _, bundle := range app.Bundles {
			if len(srcNames) > 0 && !contains(srcNames, bundle.Name) {
				continue
			}
			renders, err := renderBundle(app.Name, bundle)
			if err != nil {
				return nil, err
			}
			results = append(results, renders...)
		}
	}
	return results, nil
}

// renderRelease returns render of a Helm chart release.
func renderRelease(appName string, release Release, debug, dryRun bool) (*Render, error) {
	// If the release has a chart, render it with 'helm template'.
	if release.Chart != "" {
		cmdLine, cmd, stdout, stderr, err := execHelmTemplateCmdline(release.Name, release.Chart, release.Version, release.Values, debug, dryRun)
		return &Render{
			AppName: appName,
			SrcName: release.Name,
			SrcType: "release",
			CmdLine: cmdLine,
			Cmd:     cmd,
			Stdout:  stdout,
			Stderr:  stderr,
			Err:     err,
		}, err
	}

	// Otherwise, render the release with 'helmfile template'.
	helmfile := release.Helmfile
	if helmfile == "" {
		helmfile = path.Join(configDir, helmfileName)
	}
	cmdLine, cmd, stdout, stderr, err := execHelmfileTemplateCmd(release.Name, helmfile, debug, dryRun)
	return &Render{
		AppName: appName,
		SrcName: release.Name,
		SrcType: "release",
		CmdLine: cmdLine,
		Cmd:     cmd,
		Stdout:  stdout,
		Stderr:  stderr,
		Err:     err,
	}, err
}

// renderKustomization renders an App Kustomization object.
func renderKustomization(appName string, kustomization Kustomization, dryRun bool) (*Render, error) {
	cmdLine, cmd, stdout, stderr, err := execKustomizeBuildCmd(kustomization.Source, dryRun)
	return &Render{
		AppName: appName,
		SrcName: kustomization.Name,
		SrcType: "kustomization",
		CmdLine: cmdLine,
		Cmd:     cmd,
		Stdout:  stdout,
		Stderr:  stderr,
		Err:     err,
	}, err
}

// renderBundle renders an App Bundle object.
func renderBundle(appName string, bundle Bundle) (Renders, error) {
	renders := make(Renders, 0)
	paths, err := bundle.Paths()
	if err != nil {
		return nil, err
	}
	for _, source := range paths {
		source = path.Join(configDir, source)
		data, err := readDocument(source)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", source, err)
		}
		renders = append(renders, &Render{
			AppName: appName,
			SrcName: bundle.Name,
			SrcType: "bundle",
			CmdLine: fmt.Sprintf("cat %s", source), // No command executed for static manifests. Diagnostic only.
			Stdout:  data,
			Err:     err,
		})
	}
	urls, err := bundle.URLs()
	if err != nil {
		return nil, err
	}
	for _, source := range urls {
		data, err := fetchDocument(source)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", source, err)
		}
		renders = append(renders, &Render{
			AppName: appName,
			SrcName: bundle.Name,
			SrcType: "bundle",
			CmdLine: fmt.Sprintf("curl %s", source), // No command executed for static manifests. Diagnostic only.
			Stdout:  data,
			Err:     err,
		})
	}
	return renders, nil
}

// getHelmfileTemplateCmdline returns a 'helmfile template' line command for a Release.
func getHelmfileTemplateCmdline(releaseName, helmfile string, debug bool) string {
	cmdline := fmt.Sprintf("helmfile template --file %s --selector name=%s --skip-deps", helmfile, releaseName)
	if debug {
		cmdline += " --debug"
	}
	return cmdline
}

// execHelmfileTemplateCmd executes a 'helmfile template' command for a Release and returns its command line, command, stdout, stderr and error.
func execHelmfileTemplateCmd(releaseName, helmfile string, debug, dryRun bool) (string, *exec.Cmd, []byte, []byte, error) {
	cmdline := getHelmfileTemplateCmdline(releaseName, helmfile, debug)
	if dryRun {
		return cmdline, nil, nil, nil, nil
	}
	cmd, stdout, stderr, exitCode, err := execCmd(cmdline, path.Dir(helmfile))
	if exitCode != 0 {
		err = fmt.Errorf("helmfile template failed with exit code %d: %s", exitCode, string(stderr))
	}
	return cmdline, cmd, stdout, stderr, err
}

// getHelmTemplateCmd returns a 'helm template' command line for a Release.
func getHelmTemplateCmd(releaseName, chart, version, values string, debug bool) string {
	cmdline := fmt.Sprintf("helm template %s %s --version %s --values %s", releaseName, chart, version, values)
	if debug {
		cmdline += " --debug"
	}
	return cmdline
}

// execHelmfileTemplateCmdline executes a 'helm template' command for a Release and returns its command line, command, stdout, stderr and error.
func execHelmTemplateCmdline(releaseName, chart, version, values string, debug, dryRun bool) (string, *exec.Cmd, []byte, []byte, error) {
	cmdline := getHelmTemplateCmd(releaseName, chart, version, values, debug)
	if dryRun {
		return cmdline, nil, nil, nil, nil
	}
	cmd, stdout, stderr, exitCode, err := execCmd(cmdline, "")
	if exitCode != 0 {
		err = fmt.Errorf("helm template failed with exit code %d: %s", exitCode, string(stderr))
	}
	return cmdline, cmd, stdout, stderr, err
}

// getKustomizeBuildCmdline returns a 'kustomize build' command line for a Kustomization source.
func getKustomizeBuildCmdline(kustomizationSource string) string {
	return fmt.Sprintf("kustomize build %s", kustomizationSource)
}

// execKustomizeBuildCmd executes a 'kustomize build' command for a Kustomization and returns its command line, command, stdout, stderr and error.
func execKustomizeBuildCmd(kustomizationSource string, dryRun bool) (string, *exec.Cmd, []byte, []byte, error) {
	cmdline := getKustomizeBuildCmdline(kustomizationSource)
	if dryRun {
		return cmdline, nil, nil, nil, nil
	}
	cmd, stdout, stderr, exitCode, err := execCmd(cmdline, "")
	if exitCode != 0 {
		err = fmt.Errorf("kustomize build failed with exit code %d: %s", exitCode, string(stderr))
	}
	return cmdline, cmd, stdout, stderr, err
}
