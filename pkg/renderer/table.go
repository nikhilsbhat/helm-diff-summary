package renderer

import (
	"fmt"
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

	var (
		creates int
		updates int
		deletes int
	)

	for _, resource := range resources {
		tableWriter.AppendRow(table.Row{
			resource.Kind,
			resource.Name,
			resource.Namespace,
			resource.ChangeType,
			resource.ChangedLines,
		})

		switch resource.ChangeType {
		case parser.Create:
			creates++
		case parser.Update:
			updates++
		case parser.Delete:
			deletes++
		}
	}

	tableWriter.Render()

	fmt.Println()

	fmt.Printf("Plan: %d to create, %d to update, %d to delete.\n", creates, updates, deletes)
}
