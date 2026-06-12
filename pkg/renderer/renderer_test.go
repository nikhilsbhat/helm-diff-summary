package renderer

import (
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/policy"
)

func TestBuildSummary(t *testing.T) {
	summary := BuildSummary(sampleResources())

	if summary.Creates != 1 || summary.Updates != 1 || summary.Deletes != 1 {
		t.Fatalf("unexpected action counts: %#v", summary)
	}

	if summary.ByKind["Deployment"] != 1 || summary.ByKind["ConfigMap"] != 1 || summary.ByKind["Service"] != 1 {
		t.Fatalf("unexpected kind summary: %#v", summary.ByKind)
	}
}

func TestTableSortsByActionPriorityAndRendersNoColor(t *testing.T) {
	var writer strings.Builder
	input := New(sampleResources(), nil, BuildSummary(sampleResources()), &writer, testLogger(), true)

	if err := input.Table(); err != nil {
		t.Fatalf("Table returned error: %v", err)
	}

	output := writer.String()

	deleteIndex := strings.Index(output, "Service")
	updateIndex := strings.Index(output, "ConfigMap")
	createIndex := strings.Index(output, "Deployment")

	if deleteIndex == -1 || updateIndex == -1 || createIndex == -1 {
		t.Fatalf("expected rendered resources, got:\n%s", output)
	}

	if deleteIndex >= updateIndex || updateIndex >= createIndex {
		t.Fatalf("expected delete/update/create ordering, got:\n%s", output)
	}

	if !strings.Contains(output, "CRITICAL") || !strings.Contains(output, "UPDATE") {
		t.Fatalf("expected uncolored severity/action text, got:\n%s", output)
	}
}

func TestTextRendersSummaryAndViolations(t *testing.T) {
	violations := []policy.Violation{
		{Name: "critical", Severity: parser.Critical, Resource: "svc", Message: "delete"},
		{Name: "high", Severity: parser.High, Resource: "cm", Message: "update"},
		{Name: "medium", Severity: parser.Medium, Resource: "deploy", Message: "large"},
		{Name: "low", Severity: parser.Low, Resource: "deploy", Message: "minor"},
		{Name: "unknown", Severity: parser.Severity("UNKNOWN"), Resource: "deploy", Message: "unknown"},
	}

	var writer strings.Builder
	input := New(sampleResources(), violations, BuildSummary(sampleResources()), &writer, testLogger(), true)

	if err := input.Text(); err != nil {
		t.Fatalf("Text returned error: %v", err)
	}

	output := input.GetText()
	for _, expected := range []string{
		"Helm Deployment Summary",
		"Plan: 1 to create, 1 to update, 1 to delete.",
		"Resource Summary:",
		"Critical Violations",
		"High Severity Violations",
		"Policy Violations",
		"Violations",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}

func TestJSONRendersStructuredOutput(t *testing.T) {
	var writer strings.Builder
	input := New(sampleResources(), sampleViolations(), BuildSummary(sampleResources()), &writer, testLogger(), true)

	if err := input.JSON(); err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var output Output
	if err := json.Unmarshal([]byte(writer.String()), &output); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\n%s", err, writer.String())
	}

	if output.Plan.Deletes != 1 || len(output.Resources) != 3 || len(output.Violations) != 1 {
		t.Fatalf("unexpected JSON output: %#v", output)
	}
}

func TestYAMLRendersStructuredOutput(t *testing.T) {
	var writer strings.Builder
	input := New(sampleResources(), sampleViolations(), BuildSummary(sampleResources()), &writer, testLogger(), true)

	if err := input.YAML(); err != nil {
		t.Fatalf("YAML returned error: %v", err)
	}

	var output Output
	if err := yaml.Unmarshal([]byte(writer.String()), &output); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v\n%s", err, writer.String())
	}

	if output.Plan.Updates != 1 || len(output.Resources) != 3 || len(output.Violations) != 1 {
		t.Fatalf("unexpected YAML output: %#v", output)
	}
}

func TestPriorityAndColorFallbacks(t *testing.T) {
	var writer strings.Builder
	input := New(nil, nil, Summary{}, &writer, testLogger(), false)

	if priority(parser.ResourceDiff{ChangeType: parser.ChangeType("OTHER")}) != 99 {
		t.Fatal("expected default priority")
	}

	if got := input.coloredAction(parser.ChangeType("OTHER")); got != "OTHER" {
		t.Fatalf("unexpected fallback action: %q", got)
	}

	if got := input.coloredSeverity(parser.Severity("OTHER")); got != "OTHER" {
		t.Fatalf("unexpected fallback severity: %q", got)
	}

	for _, action := range []parser.ChangeType{parser.Create, parser.Update, parser.Delete} {
		if got := input.coloredAction(action); !strings.Contains(got, string(action)) {
			t.Fatalf("colored action %s did not contain action text: %q", action, got)
		}
	}

	for _, severity := range []parser.Severity{parser.Low, parser.Medium, parser.High, parser.Critical} {
		if got := input.coloredSeverity(severity); !strings.Contains(got, string(severity)) {
			t.Fatalf("colored severity %s did not contain severity text: %q", severity, got)
		}
	}
}

func sampleResources() []parser.ResourceDiff {
	return []parser.ResourceDiff{
		{
			Kind:         "Deployment",
			Name:         "api",
			Namespace:    "default",
			ChangeType:   parser.Create,
			Severity:     parser.Low,
			Category:     parser.Workload,
			ChangedLines: 3,
		},
		{
			Kind:         "ConfigMap",
			Name:         "config",
			Namespace:    "production",
			ChangeType:   parser.Update,
			Severity:     parser.Medium,
			Category:     parser.Config,
			ChangedLines: 2,
		},
		{
			Kind:         "Service",
			Name:         "api",
			Namespace:    "kube-system",
			ChangeType:   parser.Delete,
			Severity:     parser.Critical,
			Category:     parser.Networking,
			ChangedLines: 1,
		},
	}
}

func sampleViolations() []policy.Violation {
	return []policy.Violation{
		{Name: "resource-deletion", Severity: parser.Critical, Resource: "api", Message: "delete"},
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
}
