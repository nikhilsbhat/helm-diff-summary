package policy

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
	diffError "github.com/nikhilsbhat/helm-diff-summary/pkg/errors"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
)

// Policy implements the methods that evaluates the diffs against the policies.
type Policy interface {
	Evaluate(resources []parser.ResourceDiff) []Violation
}

type Config struct {
	Name       string            `yaml:"name,omitempty"        json:"name,omitempty"`
	Category   parser.Category   `yaml:"category,omitempty"    json:"category,omitempty"`
	Kind       string            `yaml:"kind,omitempty"        json:"kind,omitempty"`
	Action     parser.ChangeType `yaml:"action,omitempty"      json:"action,omitempty"`
	Namespace  string            `yaml:"namespace,omitempty"   json:"namespace,omitempty"`
	Severity   parser.Severity   `yaml:"severity,omitempty"    json:"severity,omitempty"`
	Message    string            `yaml:"message,omitempty"     json:"message,omitempty"`
	MinChanges int               `yaml:"min_changes,omitempty" json:"min_changes,omitempty"`
}

type Configs struct {
	Policies []Config `yaml:"policies,omitempty" json:"policies,omitempty"`
}

type Violation struct {
	Name     string          `yaml:"name,omitempty"     json:"name,omitempty"`
	Severity parser.Severity `yaml:"severity,omitempty" json:"severity,omitempty"`
	Resource string          `yaml:"resource,omitempty" json:"resource,omitempty"`
	Message  string          `yaml:"message,omitempty"  json:"message,omitempty"`
}

type Policies []Config

func (policies Policies) Evaluate(resources []parser.ResourceDiff) []Violation {
	violations := make([]Violation, 0, len(resources))

	for _, resource := range resources {
		for _, policy := range policies {
			if matchesPolicy(resource, policy) {
				violations = append(
					violations,
					Violation{
						Name:     policy.Name,
						Severity: policy.Severity,
						Resource: resource.Name,
						Message:  policy.Message,
					},
				)
			}
		}
	}

	return violations
}

func HasViolationsAtOrAbove(violations []Violation, severity parser.Severity) bool {
	for _, violation := range violations {
		if severityRank(violation.Severity) >= severityRank(severity) {
			return true
		}
	}

	return false
}

func New(policyFilepath string) (*Policies, error) {
	policies := defaultPolicies()

	if policyFilepath != "" {
		customPolicies, err := loadPolicies(policyFilepath)
		if err != nil {
			return nil, &diffError.DiffSummaryError{Message: err.Error()}
		}

		policies = append(policies, customPolicies...)
	}

	policy := Policies(policies)

	return &policy, nil
}

func loadPolicies(path string) ([]Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	var config Configs

	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.Policies, nil
}

func defaultPolicies() []Config {
	const (
		largeChangeLength   = 100
		massiveChangeLength = 500
	)

	network := networkPolicies()
	platform := platformPolicies()
	storage := storagePolicies()
	namespaces := sensitiveNamespacePolicies()
	critical := criticalPlatformPolicies()

	policies := make([]Config, 0,
		4+
			len(network)+
			len(platform)+
			len(storage)+
			len(namespaces)+
			len(critical),
	)

	policies = append(policies, []Config{
		// ------------------------------------------------------------
		// Any deletion is critical
		// ------------------------------------------------------------

		{
			Name:     "resource-deletion",
			Action:   parser.Delete,
			Severity: parser.Critical,
			Message:  "resource deletion detected",
		},

		// ------------------------------------------------------------
		// Security resources
		// ------------------------------------------------------------

		{
			Name:     "security-resource-change",
			Category: parser.Security,
			Severity: parser.High,
			Message:  "security-sensitive resource modified",
		},

		// ------------------------------------------------------------
		// Large changes
		// ------------------------------------------------------------

		{
			Name:       "large-change",
			MinChanges: largeChangeLength,
			Severity:   parser.Medium,
			Message:    "large resource change detected",
		},

		// ------------------------------------------------------------
		// Extremely large changes
		// ------------------------------------------------------------

		{
			Name:       "massive-change",
			MinChanges: massiveChangeLength,
			Severity:   parser.High,
			Message:    "extremely large resource change detected",
		},
	}...)

	policies = append(policies, network...)
	policies = append(policies, platform...)
	policies = append(policies, storage...)
	policies = append(policies, namespaces...)
	policies = append(policies, critical...)

	return policies
}

func severityRank(severity parser.Severity) int {
	const (
		defaultSeverityRank0 = 0
		defaultSeverityRank1 = 1
		defaultSeverityRank2 = 2
		defaultSeverityRank3 = 3
		defaultSeverityRank4 = 4
	)

	switch severity {
	case parser.Low:
		return defaultSeverityRank1
	case parser.Medium:
		return defaultSeverityRank2
	case parser.High:
		return defaultSeverityRank3
	case parser.Critical:
		return defaultSeverityRank4
	}

	return defaultSeverityRank0
}

func matchesPolicy(resource parser.ResourceDiff, policy Config) bool {
	if policy.Kind != "" &&
		policy.Kind != resource.Kind {
		return false
	}

	if policy.Category != "" &&
		policy.Category != resource.Category {
		return false
	}

	if policy.Action != "" &&
		policy.Action != resource.ChangeType {
		return false
	}

	if policy.Namespace != "" &&
		policy.Namespace != resource.Namespace {
		return false
	}

	if policy.MinChanges > 0 && resource.ChangedLines < policy.MinChanges {
		return false
	}

	return true
}
