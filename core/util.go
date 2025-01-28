package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// contains tests if a slice contains a given item.
func contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// execCmd executes a command and returns its result, including stdout, stderr, exit code, and error when executing the command.
func execCmd(cmdline, workingDir string) (*exec.Cmd, []byte, []byte, int, error) {
	// split command name and args out of command line
	parts := strings.Fields(cmdline)
	cmdName := parts[0]
	cmdArgs := parts[1:]
	cmd := exec.Command(cmdName, cmdArgs...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		err = nil
	}
	return cmd, stdout.Bytes(), stderr.Bytes(), exitCode, err
}

// expandTemplate replaces placeholders in a string with values from a map and returns an error if any placeholders are not expanded.
func expandTemplate(s string, data map[string]string) (string, error) {
	for key, value := range data {
		placeholder := "{" + key + "}"
		s = strings.ReplaceAll(s, placeholder, value)
	}
	if strings.Contains(s, "{") && strings.Contains(s, "}") {
		return s, fmt.Errorf("not all placeholders were expanded")
	}
	return s, nil
}

// isURL tests if the given string is a URL.
func isURL(s string) bool {
	return strings.HasPrefix(s, "https://")
}

// fetchDocument makes an HTTP GET request to the given URL and returns the document data and any error encountered.
func fetchDocument(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v", err)
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// readDocument reads the contents of a file at the given path and returns the document data and any error encountered.
func readDocument(path string) ([]byte, error) {
	return os.ReadFile(path)
}
