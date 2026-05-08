package renderer

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
)

func Render(resources []parser.ResourceDiff) {
	tableWriter := table.NewWriter()
	tableWriter.SetOutputMirror(os.Stdout)

	tableWriter.AppendHeader(table.Row{
		"KIND",
		"NAME",
		"NAMESPACE",
		"ACTION",
		"CHANGES",
	})

	for _, resource := range resources {
		tableWriter.AppendRow(table.Row{
			resource.Kind,
			resource.Name,
			resource.Namespace,
			resource.ChangeType,
			resource.ChangedLines,
		})
	}

	tableWriter.Render()
}
