package renderer

import (
	"fmt"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
)

// Summary holds information on the diff summary.
type Summary struct {
	Creates int            `yaml:"creates,omitempty" json:"creates,omitempty"`
	Updates int            `yaml:"updates,omitempty" json:"updates,omitempty"`
	Deletes int            `yaml:"deletes,omitempty" json:"deletes,omitempty"`
	ByKind  map[string]int `yaml:"byKind,omitempty"  json:"byKind,omitempty"`
}

// BuildSummary builds summary on the diff resources.
func BuildSummary(resources []parser.ResourceDiff) Summary {
	summary := Summary{
		ByKind: map[string]int{},
	}

	for _, r := range resources {
		summary.ByKind[r.Kind]++

		switch r.ChangeType {
		case parser.Create:
			summary.Creates++

		case parser.Update:
			summary.Updates++

		case parser.Delete:
			summary.Deletes++
		}
	}

	return summary
}

// Text renders output in plain text format.
func (input *Input) Text() error {
	input.writer.WriteString("\n🚀 Helm Deployment Summary\n\n")

	if err := input.Table(); err != nil {
		return err
	}

	if err := input.renderSummary(); err != nil {
		return err
	}

	if err := input.renderViolations(); err != nil {
		return err
	}

	return nil
}

func (input *Input) GetText() string {
	return input.writer.String()
}

func (input *Input) renderSummary() error {
	input.writer.WriteString("\n")

	if _, err := fmt.Fprintf(
		input.writer, "Plan: %d to create, %d to update, %d to delete.\n\n",
		input.summary.Creates, input.summary.Updates, input.summary.Deletes); err != nil {
		return err
	}

	input.writer.WriteString("Resource Summary:\n")

	for kind, count := range input.summary.ByKind {
		if _, err := fmt.Fprintf(input.writer, "  %s: %d\n", kind, count); err != nil {
			return err
		}
	}

	return nil
}

func (input *Input) renderViolations() error {
	if len(input.violations) == 0 {
		return nil
	}

	var previousSeverity parser.Severity

	for index, violation := range input.violations {
		if index == 0 || violation.Severity != previousSeverity {
			input.writer.WriteString(violationHeader(violation.Severity))

			previousSeverity = violation.Severity
		}

		if _, err := fmt.Fprintf(
			input.writer,
			"  [%s] %s: %s (%s)\n",
			violation.Severity,
			violation.Name,
			violation.Message,
			violation.Resource,
		); err != nil {
			return err
		}
	}

	return nil
}

func violationHeader(severity parser.Severity) string {
	switch severity {
	case parser.Critical:
		return "\n🚨 Critical Violations\n\n"
	case parser.High:
		return "\n⚠ High Severity Violations\n\n"
	case parser.Medium:
		return "\n⚠ Policy Violations\n\n"
	case parser.Low:
		return "\nℹ Violations\n\n"
	default:
		return "\nℹ Violations\n\n"
	}
}
