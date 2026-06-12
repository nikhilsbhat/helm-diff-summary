package policy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
)

func TestPoliciesEvaluateMatchesAllConfiguredFields(t *testing.T) {
	policies := Policies{
		{
			Name:       "prod-deployment-update",
			Kind:       "Deployment",
			Category:   parser.Workload,
			Action:     parser.Update,
			Namespace:  "production",
			Severity:   parser.High,
			Message:    "deployment updated in production",
			MinChanges: 3,
		},
		{
			Name:       "too-large",
			MinChanges: 10,
			Severity:   parser.Medium,
			Message:    "too large",
		},
	}

	violations := policies.Evaluate([]parser.ResourceDiff{
		{
			Kind:         "Deployment",
			Name:         "api",
			Namespace:    "production",
			Category:     parser.Workload,
			ChangeType:   parser.Update,
			ChangedLines: 4,
		},
		{
			Kind:         "Deployment",
			Name:         "worker",
			Namespace:    "staging",
			Category:     parser.Workload,
			ChangeType:   parser.Update,
			ChangedLines: 12,
		},
	})

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %#v", violations)
	}

	if violations[0].Name != "prod-deployment-update" || violations[0].Resource != "api" {
		t.Fatalf("unexpected first violation: %#v", violations[0])
	}

	if violations[1].Name != "too-large" || violations[1].Resource != "worker" {
		t.Fatalf("unexpected second violation: %#v", violations[1])
	}
}

func TestHasViolationsAtOrAbove(t *testing.T) {
	violations := []Violation{
		{Severity: parser.Low},
		{Severity: parser.High},
	}

	if !HasViolationsAtOrAbove(violations, parser.Medium) {
		t.Fatal("expected high violation to satisfy medium threshold")
	}

	if HasViolationsAtOrAbove(violations, parser.Critical) {
		t.Fatal("did not expect critical threshold to be satisfied")
	}

	if HasViolationsAtOrAbove([]Violation{{Severity: parser.Severity("UNKNOWN")}}, parser.Low) {
		t.Fatal("unknown severity should not satisfy low threshold")
	}
}

func TestNewLoadsDefaultAndCustomPolicies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "helm-diff-summary.yaml")

	err := os.WriteFile(path, []byte(`
policies:
  - name: custom-policy
    kind: Deployment
    action: UPDATE
    severity: HIGH
    message: custom message
`), 0o600)
	if err != nil {
		t.Fatalf("failed to write policy file: %v", err)
	}

	policies, err := New(path)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if len(*policies) <= len(defaultPolicies()) {
		t.Fatalf("expected custom policy to be appended, got %d policies", len(*policies))
	}

	found := false
	for _, policy := range *policies {
		if policy.Name == "custom-policy" {
			found = true

			break
		}
	}

	if !found {
		t.Fatal("custom policy was not loaded")
	}
}

func TestNewMissingPolicyFileUsesDefaults(t *testing.T) {
	policies, err := New(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("New returned error for missing file: %v", err)
	}

	if len(*policies) != len(defaultPolicies()) {
		t.Fatalf("expected defaults only, got %d policies", len(*policies))
	}
}

func TestNewReturnsErrorForInvalidPolicyYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "helm-diff-summary.yaml")

	if err := os.WriteFile(path, []byte("policies: ["), 0o600); err != nil {
		t.Fatalf("failed to write policy file: %v", err)
	}

	if _, err := New(path); err == nil {
		t.Fatal("expected invalid YAML error")
	}
}

func TestDefaultPolicyGroups(t *testing.T) {
	if len(storagePolicies()) != 2 {
		t.Fatal("expected storage policies")
	}
	if len(platformPolicies()) != 2 {
		t.Fatal("expected platform policies")
	}
	if len(networkPolicies()) != 2 {
		t.Fatal("expected network policies")
	}
	if len(sensitiveNamespacePolicies()) == 0 {
		t.Fatal("expected sensitive namespace policies")
	}
	if len(criticalPlatformPolicies()) == 0 {
		t.Fatal("expected critical platform policies")
	}
}

func TestMatchesPolicyRejectsMismatchedFields(t *testing.T) {
	resource := parser.ResourceDiff{
		Kind:         "Deployment",
		Namespace:    "default",
		Category:     parser.Workload,
		ChangeType:   parser.Update,
		ChangedLines: 4,
	}

	tests := []Config{
		{Kind: "Service"},
		{Category: parser.Networking},
		{Action: parser.Delete},
		{Namespace: "production"},
		{MinChanges: 5},
	}

	for _, policy := range tests {
		if matchesPolicy(resource, policy) {
			t.Fatalf("expected policy not to match: %#v", policy)
		}
	}
}
