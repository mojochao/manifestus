package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"

	"github.com/mojochao/manifestus/core"
)

const (
	defaultRenderfileName = "renderfile.yaml"
	defaultOutputDir      = "manifests"
)

// New returns a new CLI app.
func New() *cli.App {
	return &cli.App{
		Name:  "manifestus",
		Usage: "Render Kubernetes manifests from a declarative configuration",
		Commands: []*cli.Command{
			appsCommand,
			chartsCommand,
			outputsCommand,
			renderCommand,
			writeCommand,
			checkCommand,
			versionCommand,
		},
	}
}

var appsCommand = &cli.Command{
	Name:  "apps",
	Usage: "Show list of all apps",
	Flags: []cli.Flag{
		&renderfileFlag,
		&appNamesFlag,
	},
	Action: func(c *cli.Context) error {
		// Load the config file from disk.
		cfg, err := core.LoadConfig(flags.RenderFile)
		exitOnError(err, 1)

		// Print the names of all apps in the config that are not explicitly disabled to stdout.
		for _, app := range cfg.EnabledApps() {
			fmt.Println(app.Name)
		}
		return nil
	},
}

var chartsCommand = &cli.Command{
	Name:  "charts",
	Usage: "Show table of charts used",
	Flags: []cli.Flag{
		&renderfileFlag,
		&appNamesFlag,
		&latestFlag,
		&outdatedFlag,
	},
	Action: func(c *cli.Context) error {
		// Load the config file from disk.
		cfg, err := core.LoadConfig(flags.RenderFile)
		exitOnError(err, -1)

		// Ensure that we have app names and the app names exist in the config.
		appNames := flags.AppNames.Value()
		if len(appNames) == 0 {
			appNames = cfg.EnabledAppNames()
		} else {
			err = core.EnsureAppNamesExist(cfg, flags.AppNames.Value())
			exitOnError(err, -1)
		}

		charts, err := core.GetCharts(cfg, appNames)
		exitOnError(err, -1)

		latest := latestFlag.Value
		outdated := outdatedFlag.Value
		if latest || outdated {
			err := core.ExecHelmRepoUpdate()
			exitOnError(err, -1)
		}

		table, err := getChartsTable(charts, latest, outdated)
		exitOnError(err, -1)

		table.Print()
		return nil
	},
}

var outputsCommand = &cli.Command{
	Name:  "outputs",
	Usage: "Show list of output files of rendered manifests",
	Flags: []cli.Flag{
		&renderfileFlag,
		&appNamesFlag,
		&srcNamesFlag,
		&srcTypesFlag,
	},
	Action: func(c *cli.Context) error {
		// Load the config file from disk.
		cfg, err := core.LoadConfig(flags.RenderFile)
		exitOnError(err, -1)

		// Ensure that we have app names and the app names exist in the config.
		appNames := flags.AppNames.Value()
		if len(appNames) == 0 {
			appNames = cfg.EnabledAppNames()
		} else {
			err = core.EnsureAppNamesExist(cfg, flags.AppNames.Value())
			exitOnError(err, -1)
		}

		// Ensure that we have src types and they are valid.
		srcTypes := flags.SrcTypes.Value()
		if len(srcTypes) == 0 {
			srcTypes = core.ValidSrcTypes
		} else {
			err = core.EnsureSrcTypesValid(flags.SrcTypes.Value())
			exitOnError(err, -1)
		}

		// Print the output files of the rendered manifests for the apps to stdout.
		outputs := core.GetOutputFiles(cfg, appNames, flags.SrcNames.Value(), srcTypes)
		for _, output := range outputs {
			fmt.Println(output)
		}
		return nil
	},
}

var renderCommand = &cli.Command{
	Name:  "render",
	Usage: "Render manifests to standard output",
	Flags: []cli.Flag{
		&renderfileFlag,
		&appNamesFlag,
		&srcNamesFlag,
		&srcTypesFlag,
		&dryRunFlag,
		&debugFlag,
	},
	Action: func(c *cli.Context) error {
		// Load the config file from disk.
		cfg, err := core.LoadConfig(flags.RenderFile)
		exitOnError(err, -1)

		// Ensure that we have app names and the app names exist in the config.
		appNames := flags.AppNames.Value()
		if len(appNames) == 0 {
			appNames = cfg.EnabledAppNames()
		} else {
			err = core.EnsureAppNamesExist(cfg, appNames)
			exitOnError(err, -1)
		}

		// Ensure that we have src types and they are valid.
		srcTypes := flags.SrcTypes.Value()
		if len(srcTypes) == 0 {
			srcTypes = core.ValidSrcTypes
		} else {
			err = core.EnsureSrcTypesValid(flags.SrcTypes.Value())
			exitOnError(err, -1)
		}

		// Get the renders for the apps and ensure that they are OK.
		renders, err := core.GetRenders(cfg, appNames, flags.SrcNames.Value(), srcTypes, flags.Debug, flags.DryRun)
		exitOnError(err, -1)

		// If dry-run is enabled, just print the command lines to stdout and return.
		if flags.DryRun {
			for _, render := range renders {
				if render.CmdLine != "" { // Skip static manifests as they aren't rendered with a command line.
					fmt.Println(render.CmdLine)
				}
			}
			return nil
		}

		// Otherwise, print the rendered manifests to stdout and return.
		manifests := core.GetManifests(renders)
		for _, manifest := range manifests {
			fmt.Println(manifest.Doc())
		}
		return nil
	},
}

var writeCommand = &cli.Command{
	Name:  "write",
	Usage: "Write rendered manifests to output directory",
	Flags: []cli.Flag{
		&renderfileFlag,
		&outputDirFlag,
		&appNamesFlag,
		&srcNamesFlag,
		&srcTypesFlag,
		&cleanFlag,
		&debugFlag,
		&verboseFlag,
	},
	Action: func(c *cli.Context) error {
		// Load the config file from disk.
		cfg, err := core.LoadConfig(flags.RenderFile)
		exitOnError(err, -1)

		// Ensure that we have app names and the app names exist in the config.
		appNames := flags.AppNames.Value()
		if len(appNames) == 0 {
			appNames = cfg.EnabledAppNames()
		} else {
			err = core.EnsureAppNamesExist(cfg, appNames)
			exitOnError(err, -1)
		}

		// Ensure that we have src types and they are valid.
		srcTypes := flags.SrcTypes.Value()
		if len(srcTypes) == 0 {
			srcTypes = core.ValidSrcTypes
		} else {
			err = core.EnsureSrcTypesValid(flags.SrcTypes.Value())
			exitOnError(err, -1)
		}

		// Get the renders for the apps and ensure that they are OK.
		// Unlike the 'render' command, we won't allow dry-run here as we want to
		// update the rendered manifests in the output directory.
		renders, err := core.GetRenders(cfg, appNames, flags.SrcNames.Value(), srcTypes, flags.Debug, flags.DryRun)
		exitOnError(err, -1)

		// Clean output directories if the clean flag is set.
		if flags.Clean {
			for _, appName := range appNames {
				appDir := path.Join(flags.OutputDir, appName)
				printMsg(fmt.Sprintf("Cleaning up %s\n", appDir), true)
				if err := os.RemoveAll(appDir); err != nil {
					exitOnError(err, -1)
				}
			}
		}

		// Write the rendered manifests to the output directory.
		manifests := core.GetManifests(renders)
		for _, manifest := range manifests {
			path, err := manifest.Write(flags.OutputDir)
			exitOnError(err, -1)
			printMsg(fmt.Sprintf("Wrote %s\n", path), true)
		}
		return nil
	},
}

var checkCommand = &cli.Command{
	Name:  "check",
	Usage: "Check that rendered manifests are up-to-date with their sources.\n\nExit with status code 1 if differences are found.",
	Flags: []cli.Flag{
		&renderfileFlag,
		&outputDirFlag,
		&appNamesFlag,
		&debugFlag,
		&quietFlag,
		&verboseFlag,
	},
	Action: func(c *cli.Context) error {
		// Load the config file from disk.
		cfg, err := core.LoadConfig(flags.RenderFile)
		exitOnError(err, -1)

		// Ensure that we have app names and the app names exist in the config.
		appNames := flags.AppNames.Value()
		if len(appNames) == 0 {
			appNames = cfg.EnabledAppNames()
		} else {
			err = core.EnsureAppNamesExist(cfg, appNames)
			exitOnError(err, -1)
		}

		// Get the renders for the apps and ensure that they are OK.
		// Unlike the 'render' command, we won't allow dry-run here as we want to
		// update the rendered manifests in the output directory.
		renders, err := core.GetRenders(cfg, appNames, nil, core.ValidSrcTypes, flags.Debug, flags.DryRun)
		exitOnError(err, -1)

		// Ensure that we're starting with a clean temp directory.
		tempDir, err := os.MkdirTemp("", "manifestus")
		exitOnError(err, -1)

		// Ensure thaw we're cleaning up the temp directory when we're done.
		defer func(path string) {
			if flags.Verbose {
				printMsg(fmt.Sprintf("Cleaning up manifest output directory: %s\n", path), true)
			}
			err := os.RemoveAll(path)
			exitOnError(err, -1)
		}(tempDir)

		// Write the rendered manifests to the output directory.
		manifests := core.GetManifests(renders)
		for _, manifest := range manifests {
			printMsg(fmt.Sprintf("Writing %s", manifest.AppName), true)
			_, err := manifest.Write(tempDir)
			exitOnError(err, -1)
		}

		// Test if the contents of the output dir and the temp dir are the same.
		diff, err := diffDirs(flags.OutputDir, tempDir)
		exitOnError(err, -1)

		// If there are differences, show them and exit with a non-zero exit code to indicate differences found.
		if len(diff) > 0 {
			printMsg("Rendered manifests are not up-to-date with their sources", false)
			os.Exit(1)
		}
		printMsg("Rendered manifests are up-to-date with their sources", false)
		return nil
	},
}

var versionCommand = &cli.Command{
	Name:  "version",
	Usage: "Show version",
	Action: func(c *cli.Context) error {
		fmt.Println(core.Version)
		return nil
	},
}

// flags is used to store the values of the flags passed to the CLI
var flags struct {
	RenderFile string
	OutputDir  string
	AppNames   cli.StringSlice
	SrcNames   cli.StringSlice
	SrcTypes   cli.StringSlice
	Clean      bool
	Debug      bool
	DryRun     bool
	Quiet      bool
	Verbose    bool
	Latest     bool
	Outdated   bool
}

var renderfileFlag = cli.StringFlag{
	Name:        "renderfile",
	Aliases:     []string{"r"},
	Usage:       "Specify the path to the Renderfile",
	Destination: &flags.RenderFile,
	Value:       defaultRenderfileName,
}

var outputDirFlag = cli.StringFlag{
	Name:        "output-dir",
	Aliases:     []string{"o"},
	Usage:       "Specify the output directory for rendered manifests",
	Destination: &flags.OutputDir,
	Value:       defaultOutputDir,
}

var appNamesFlag = cli.StringSliceFlag{
	Name:        "app",
	Aliases:     []string{"a"},
	Usage:       "Specify the name of an app to render manifests for",
	Destination: &flags.AppNames,
}

var srcNamesFlag = cli.StringSliceFlag{
	Name:        "name",
	Aliases:     []string{"n"},
	Usage:       "Specify the name of a source to render",
	Destination: &flags.SrcNames,
}

var validSrcTypes = strings.Join(core.ValidSrcTypes, " | ")

var srcTypesFlag = cli.StringSliceFlag{
	Name:        "type",
	Aliases:     []string{"t"},
	Usage:       fmt.Sprintf("Specify the type of source to render (valid: %s)", validSrcTypes),
	Destination: &flags.SrcTypes,
}

var cleanFlag = cli.BoolFlag{
	Name:        "clean",
	Usage:       "Clean the output directory before writing rendered manifests",
	Destination: &flags.Clean,
}

var debugFlag = cli.BoolFlag{
	Name:        "debug",
	Usage:       "Show debug output",
	Destination: &flags.Debug,
}

var dryRunFlag = cli.BoolFlag{
	Name:        "dry-run",
	Usage:       "Preview render commands without executing them",
	Destination: &flags.DryRun,
}

var quietFlag = cli.BoolFlag{
	Name:        "quiet",
	Aliases:     []string{"q"},
	Usage:       "Suppress output",
	Destination: &flags.Quiet,
}

var verboseFlag = cli.BoolFlag{
	Name:        "verbose",
	Aliases:     []string{"v"},
	Usage:       "Show verbose output",
	Destination: &flags.Verbose,
}

var latestFlag = cli.BoolFlag{
	Name:        "latest",
	Usage:       "Show the latest version of each chart",
	Destination: &flags.Latest,
}

var outdatedFlag = cli.BoolFlag{
	Name:        "outdated",
	Usage:       "Show only outdated charts",
	Destination: &flags.Outdated,
}

// diffDirs runs the `diff` command to compare the contents of two directories.
// An empty string is returned if the directories are the same.
// An error is returned if the `diff` command fails.
func diffDirs(dir1, dir2 string) (string, error) {
	cmd := exec.Command("diff", "-r", dir1, dir2)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

// exitOnError prints the error to stdout and exits with the given exit code.
func exitOnError(err error, exitCode int) {
	if err == nil {
		return
	}
	fmt.Printf("Error: %v\n", err)
	os.Exit(exitCode)
}

// printMsg prints a message to stderr if the quiet flag is not set and verbosity requirements met
func printMsg(msg string, verboseOnly bool) {
	if flags.Quiet {
		return
	}
	if verboseOnly && !flags.Verbose {
		return
	}
	// Trim newline from message to avoid double newlines.
	_, err := fmt.Fprintln(os.Stdout, strings.TrimSuffix(msg, "\n"))
	exitOnError(err, -1)
}

// getChartsTable returns a table of charts.
func getChartsTable(charts []*core.Chart, includeLatest, onlyOutdated bool) (table.Table, error) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("App", "Chart", "Version")
	if includeLatest {
		tbl = table.New("App", "Chart", "Version", "Latest")
	}
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, chart := range charts {
		if !includeLatest && !onlyOutdated {
			tbl.AddRow(chart.App, chart.Name, chart.Version)
			continue
		}
		if includeLatest {
			latestVersion, err := chart.LatestVersion()
			if err != nil {
				return tbl, err
			}
			if !onlyOutdated || chart.Version != latestVersion {
				tbl.AddRow(chart.App, chart.Name, chart.Version, latestVersion)
			}
		}
	}
	return tbl, nil
}
