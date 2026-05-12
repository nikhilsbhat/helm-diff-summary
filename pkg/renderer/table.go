package renderer

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
)

// Input holds the information that are required for rendering the output.
type Input struct {
	resources []parser.ResourceDiff
	summary   Summary
	noColor   bool
}

// Render implements the methods that renders output in various format.
type Render interface {
	Table() error
	JSON() error
	YAML() error
}

// Table renders the output in table format.
func (input *Input) Table() error {
	sort.Slice(input.resources, func(i, j int) bool {
		return priority(input.resources[i]) < priority(input.resources[j])
	})

	tableWriter := table.NewWriter()
	tableWriter.SetOutputMirror(os.Stdout)

	tableWriter.AppendHeader(table.Row{
		"KIND",
		"NAME",
		"NAMESPACE",
		"ACTION",
		"SEVERITY",
		"CATEGORY",
		"CHANGES",
	})

	for _, resource := range input.resources {
		tableWriter.AppendRow(table.Row{
			resource.Kind,
			resource.Name,
			resource.Namespace,
			coloredAction(resource.ChangeType),
			coloredSeverity(resource.Severity),
			resource.Category,
			resource.ChangedLines,
		})
	}

	tableWriter.Render()

	return printSummary(input.summary)
}

// New returns new instance of Input when invoked.
func New(resources []parser.ResourceDiff, summary Summary, noColor bool) *Input {
	return &Input{
		resources: resources,
		summary:   summary,
		noColor:   noColor,
	}
}

func printSummary(summary Summary) error {
	var builder strings.Builder

	builder.WriteString("\n")

	if _, err := fmt.Fprintf(
		&builder, "Plan: %d to create, %d to update, %d to delete.\n\n",
		summary.Creates, summary.Updates, summary.Deletes); err != nil {
		return err
	}

	builder.WriteString("Resource Summary:\n")

	for kind, count := range summary.ByKind {
		if _, err := fmt.Fprintf(&builder, "  %s: %d\n", kind, count); err != nil {
			return err
		}
	}

	fmt.Print(builder.String())

	return nil
}

func priority(resource parser.ResourceDiff) int {
	const (
		deleteCode  = 0
		updateCode  = 1
		createCode  = 2
		defaultCode = 99
	)

	switch resource.ChangeType {
	case parser.Delete:
		return deleteCode

	case parser.Update:
		return updateCode

	case parser.Create:
		return createCode
	}

	return defaultCode
}

func coloredAction(action parser.ChangeType) string {
	switch action {
	case parser.Create:
		return color.New(color.FgGreen).Sprint(action)

	case parser.Update:
		return color.New(color.FgYellow).Sprint(action)

	case parser.Delete:
		return color.New(color.FgRed).Sprint(action)
	}

	return string(action)
}

func coloredSeverity(severity parser.Severity) string {
	switch severity {
	case parser.Low:
		return color.New(color.FgGreen).Sprint(severity)

	case parser.Medium:
		return color.New(color.FgYellow).Sprint(severity)

	case parser.High:
		return color.New(color.FgHiRed).Sprint(severity)

	case parser.Critical:
		return color.New(color.BgRed, color.FgWhite, color.Bold).Sprint(severity)
	}

	return string(severity)
}
