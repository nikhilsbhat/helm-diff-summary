package cmd

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/policy"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/renderer"
)

func TestGetRootCommand(t *testing.T) {
	root := GetRootCommand()

	if root.Use != "helm-diff-summary [command]" {
		t.Fatalf("unexpected use: %s", root.Use)
	}

	if !strings.Contains(root.Example, "helm diff upgrade my-release ./chart") {
		t.Fatalf("expected updated examples, got %q", root.Example)
	}
}

func TestVersionInfo(t *testing.T) {
	var writer bytes.Buffer
	versionInfo(&writer)

	if !strings.Contains(writer.String(), "helm-diff-summary version:") {
		t.Fatalf("unexpected version output: %q", writer.String())
	}
}

func TestSetCLIClient(t *testing.T) {
	cliCfg.logLevel = "debug"

	if err := setCLIClient(nil, nil); err != nil {
		t.Fatalf("setCLIClient returned error: %v", err)
	}

	if logger == nil {
		t.Fatal("expected logger to be configured")
	}
}

func TestRunParsesAndRendersInput(t *testing.T) {
	var out strings.Builder

	restoreCmdGlobals(t)
	stdin = strings.NewReader(`
default, api, Deployment (apps) has been added:
+ apiVersion: apps/v1
+ kind: Deployment
`)
	stdout = &out
	stderr = io.Discard
	statStdin = func() (os.FileInfo, error) {
		return fakeFileInfo{mode: 0}, nil
	}
	exitProcess = func(code int) {
		t.Fatalf("unexpected exit with code %d", code)
	}
	cliCfg.outputFormat = "json"
	cliCfg.showVersion = false
	cliCfg.failOnSeverity = ""
	cliCfg.failOnDelete = false
	cliCfg.noColor = true
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := run(nil, nil); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(out.String(), `"creates": 1`) {
		t.Fatalf("expected rendered JSON output, got %q", out.String())
	}
}

func TestRunHandlesNoResources(t *testing.T) {
	exitCode := -1
	var errOut strings.Builder

	restoreCmdGlobals(t)
	stdin = strings.NewReader("no matching resources")
	stderr = &errOut
	statStdin = func() (os.FileInfo, error) {
		return fakeFileInfo{mode: 0}, nil
	}
	exitProcess = func(code int) {
		exitCode = code
	}
	cliCfg.showVersion = false

	if err := run(nil, nil); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(errOut.String(), "no resources detected") {
		t.Fatalf("unexpected stderr: %q", errOut.String())
	}
}

func TestRunReturnsPolicyLoadError(t *testing.T) {
	restoreCmdGlobals(t)

	dir := t.TempDir()
	originalWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkingDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	if err := os.WriteFile(filepath.Join(dir, "helm-diff-summary.yaml"), []byte("policies: ["), 0o600); err != nil {
		t.Fatalf("failed to write policy file: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	stdin = strings.NewReader(`
default, api, Deployment (apps) has been added:
+ apiVersion: apps/v1
`)
	stdout = io.Discard
	stderr = io.Discard
	statStdin = func() (os.FileInfo, error) {
		return fakeFileInfo{mode: 0}, nil
	}
	exitProcess = func(code int) {
		t.Fatalf("unexpected exit with code %d", code)
	}
	cliCfg.showVersion = false

	if err := run(nil, nil); err == nil {
		t.Fatal("expected policy load error")
	}
}

func TestValidateInputExitsForTerminalInput(t *testing.T) {
	var errOut strings.Builder
	var exitCode int

	restoreCmdGlobals(t)
	stderr = &errOut
	statStdin = func() (os.FileInfo, error) {
		return fakeFileInfo{mode: os.ModeCharDevice}, nil
	}
	exitProcess = func(code int) {
		exitCode = code
	}

	if err := validateInput(); err != nil {
		t.Fatalf("validateInput returned error: %v", err)
	}

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(errOut.String(), "no helm diff input detected") {
		t.Fatalf("unexpected stderr: %q", errOut.String())
	}
}

func TestHandleNoResourcesExitsZero(t *testing.T) {
	var errOut strings.Builder
	exitCode := -1

	restoreCmdGlobals(t)
	stderr = &errOut
	exitProcess = func(code int) {
		exitCode = code
	}

	if err := handleNoResources(); err != nil {
		t.Fatalf("handleNoResources returned error: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(errOut.String(), "no resources detected") {
		t.Fatalf("unexpected stderr: %q", errOut.String())
	}
}

func TestExecuteVersion(t *testing.T) {
	err := execute([]string{"--version"})
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
}

func TestMainExecutesCommand(t *testing.T) {
	restoreCmdGlobals(t)

	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })

	os.Args = []string{"helm-diff-summary", "--version"}
	stdout = io.Discard

	Main()
}

func TestRenderOutputFormats(t *testing.T) {
	restoreCmdGlobals(t)

	logger = slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	resources := []parser.ResourceDiff{
		{
			Kind:         "Deployment",
			Name:         "api",
			Namespace:    "default",
			ChangeType:   parser.Create,
			Severity:     parser.Low,
			Category:     parser.Workload,
			ChangedLines: 1,
		},
	}
	summary := renderer.BuildSummary(resources)

	tests := []string{"json", "j", "yaml", "y", "table"}
	for _, format := range tests {
		t.Run(format, func(t *testing.T) {
			cliCfg.outputFormat = format
			cliCfg.notifiers = nil
			stdout = io.Discard
			t.Cleanup(func() {
				cliCfg.outputFormat = ""
				cliCfg.notifiers = nil
			})

			var builder strings.Builder
			input := renderer.New(resources, nil, summary, &builder, logger, true)

			if err := renderOutput(input); err != nil {
				t.Fatalf("renderOutput returned error: %v", err)
			}

			if builder.Len() == 0 {
				t.Fatal("expected rendered output")
			}
		})
	}
}

func TestHandleExitConditionsNoExit(t *testing.T) {
	restoreCmdGlobals(t)

	cliCfg.failOnSeverity = "critical"
	cliCfg.failOnDelete = true
	t.Cleanup(func() {
		cliCfg.failOnSeverity = ""
		cliCfg.failOnDelete = false
	})

	err := handleExitConditions(
		[]policy.Violation{{Severity: parser.Low}},
		renderer.Summary{Deletes: 0},
	)
	if err != nil {
		t.Fatalf("handleExitConditions returned error: %v", err)
	}
}

func TestHandleExitConditionsSeverityExit(t *testing.T) {
	var exitCode int

	restoreCmdGlobals(t)
	cliCfg.failOnSeverity = "high"
	exitProcess = func(code int) {
		exitCode = code
	}

	err := handleExitConditions(
		[]policy.Violation{{Severity: parser.Critical}},
		renderer.Summary{},
	)
	if err != nil {
		t.Fatalf("handleExitConditions returned error: %v", err)
	}

	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
}

func TestHandleExitConditionsDeleteExit(t *testing.T) {
	var exitCode int

	restoreCmdGlobals(t)
	cliCfg.failOnDelete = true
	exitProcess = func(code int) {
		exitCode = code
	}

	err := handleExitConditions(nil, renderer.Summary{Deletes: 1})
	if err != nil {
		t.Fatalf("handleExitConditions returned error: %v", err)
	}

	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
}

func TestSendNotificationsWithNoTargets(t *testing.T) {
	cliCfg.notifiers = nil

	if err := sendNotifications("hello"); err != nil {
		t.Fatalf("sendNotifications returned error: %v", err)
	}
}

func restoreCmdGlobals(t *testing.T) {
	t.Helper()

	originalStdin := stdin
	originalStdout := stdout
	originalStderr := stderr
	originalExitProcess := exitProcess
	originalStatStdin := statStdin
	originalCfg := *cliCfg

	t.Cleanup(func() {
		stdin = originalStdin
		stdout = originalStdout
		stderr = originalStderr
		exitProcess = originalExitProcess
		statStdin = originalStatStdin
		*cliCfg = originalCfg
	})
}

type fakeFileInfo struct {
	mode os.FileMode
}

func (info fakeFileInfo) Name() string       { return "stdin" }
func (info fakeFileInfo) Size() int64        { return 0 }
func (info fakeFileInfo) Mode() os.FileMode  { return info.mode }
func (info fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (info fakeFileInfo) IsDir() bool        { return false }
func (info fakeFileInfo) Sys() any           { return nil }
