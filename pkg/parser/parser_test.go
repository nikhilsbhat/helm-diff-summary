package parser

import (
	"strings"
	"testing"
)

func TestParseCountsAndClassifiesResources(t *testing.T) {
	input := strings.NewReader(`
ignored preface
default, sample-api, Deployment (apps) has been added:
+ apiVersion: apps/v1
+ kind: Deployment
+ metadata:
+   name: sample-api
+ # ignored comment
+++ ignored file header
production, sample-config, ConfigMap (v1) has changed:
- old: value
+ new: value
+ @@ ignored hunk
+ # ignored comment
kube-system, sample-service, Service (v1) has been removed:
- apiVersion: v1
- kind: Service
--- ignored file header
`)

	resources, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d: %#v", len(resources), resources)
	}

	assertResource(t, resources[0], ResourceDiff{
		Kind:         "Deployment",
		Name:         "sample-api",
		Namespace:    "default",
		ChangeType:   Create,
		Category:     Workload,
		Additions:    4,
		ChangedLines: 4,
		Severity:     Low,
	})

	assertResource(t, resources[1], ResourceDiff{
		Kind:         "ConfigMap",
		Name:         "sample-config",
		Namespace:    "production",
		ChangeType:   Update,
		Category:     Config,
		Additions:    1,
		Deletions:    1,
		ChangedLines: 1,
		Severity:     Medium,
	})

	assertResource(t, resources[2], ResourceDiff{
		Kind:         "Service",
		Name:         "sample-service",
		Namespace:    "kube-system",
		ChangeType:   Delete,
		Category:     Networking,
		Deletions:    2,
		ChangedLines: 2,
		Severity:     Critical,
	})
}

func TestParseIgnoresNonMatchingInput(t *testing.T) {
	resources, err := Parse(strings.NewReader("plain text\n+ value\n- value\n"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(resources) != 0 {
		t.Fatalf("expected no resources, got %#v", resources)
	}
}

func TestParseSeverity(t *testing.T) {
	tests := map[string]Severity{
		"low":      Low,
		"MEDIUM":   Medium,
		"High":     High,
		"critical": Critical,
	}

	for input, expected := range tests {
		actual, err := ParseSeverity(input)
		if err != nil {
			t.Fatalf("ParseSeverity(%q) returned error: %v", input, err)
		}

		if actual != expected {
			t.Fatalf("ParseSeverity(%q) = %s, want %s", input, actual, expected)
		}
	}

	if _, err := ParseSeverity("urgent"); err == nil {
		t.Fatal("expected invalid severity error")
	}
}

func TestDetectSeverityThresholds(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceDiff
		want     Severity
	}{
		{
			name: "unknown update is medium",
			resource: ResourceDiff{
				Kind:         "Widget",
				ChangeType:   Update,
				Category:     Unknown,
				ChangedLines: 1,
			},
			want: Medium,
		},
		{
			name: "storage update is medium",
			resource: ResourceDiff{
				Kind:         "PersistentVolumeClaim",
				ChangeType:   Update,
				Category:     Storage,
				ChangedLines: 20,
			},
			want: Medium,
		},
		{
			name: "large platform delete is critical",
			resource: ResourceDiff{
				Kind:         "CustomResourceDefinition",
				Namespace:    "crossplane-system",
				ChangeType:   Delete,
				Category:     Platform,
				ChangedLines: 250,
			},
			want: Critical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectSeverity(&tt.resource)
			if got != tt.want {
				t.Fatalf("detectSeverity() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDetectCategory(t *testing.T) {
	tests := map[string]Category{
		"Deployment":               Workload,
		"Service":                  Networking,
		"ClusterRole":              Security,
		"PersistentVolumeClaim":    Storage,
		"CustomResourceDefinition": Platform,
		"ConfigMap":                Config,
		"Widget":                   Unknown,
	}

	for kind, expected := range tests {
		if actual := detectCategory(kind); actual != expected {
			t.Fatalf("detectCategory(%q) = %s, want %s", kind, actual, expected)
		}
	}
}

func assertResource(t *testing.T, actual ResourceDiff, expected ResourceDiff) {
	t.Helper()

	if actual.Kind != expected.Kind ||
		actual.Name != expected.Name ||
		actual.Namespace != expected.Namespace ||
		actual.ChangeType != expected.ChangeType ||
		actual.Category != expected.Category ||
		actual.Severity != expected.Severity ||
		actual.Additions != expected.Additions ||
		actual.Deletions != expected.Deletions ||
		actual.ChangedLines != expected.ChangedLines {
		t.Fatalf("resource mismatch\nactual:   %#v\nexpected: %#v", actual, expected)
	}
}
