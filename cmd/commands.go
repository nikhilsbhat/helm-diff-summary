package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/renderer"
	"github.com/nikhilsbhat/helm-diff-summary/version"
	"github.com/spf13/cobra"
)

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:     "helm-diff-summary [command]",
		Short:   "A Terraform-style summarizer for helm diff output",
		Long:    `An utility that reads the helm diff plugin's output and summarizes its output in a Terraform style`,
		PreRunE: setCLIClient,
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if cliCfg.showVersion {
				versionInfo(os.Stdout)

				return nil
			}

			stat, err := os.Stdin.Stat()
			if err != nil {
				log.Fatalf("failed to inspect stdin: %v", err)
			}

			// No piped stdin
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				if _, err = fmt.Fprintln(
					os.Stderr,
					"no helm diff input detected"); err != nil {
					return err
				}

				if _, err = fmt.Fprintln(os.Stderr); err != nil {
					return err
				}

				os.Exit(1)
			}

			resources, err := parser.Parse(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}

			if len(resources) == 0 {
				if _, err = fmt.Fprintln(os.Stderr, "no resources detected in helm diff output"); err != nil {
					return err
				}

				os.Exit(0)
			}

			summary := renderer.BuildSummary(resources)

			switch cliCfg.outputFormat {
			case "json", "j":
				return renderer.RenderJSON(resources, summary)
			case "yaml", "y":
				return renderer.RenderYAML(resources, summary)
			default:
				if err = renderer.RenderTable(resources, summary); err != nil {
					return err
				}
			}

			const exitCode = 2

			if cliCfg.failOnDelete && summary.Deletes > 0 {
				os.Exit(exitCode)
			}

			return nil
		},
	}

	rootCommand.SetUsageTemplate(getUsageTemplate())

	registerCommonFlags(rootCommand)

	return rootCommand
}

func versionInfo(writer io.Writer) {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		log.Fatalf("failed to fetch version info: %v", err)
	}

	if _, err = fmt.Fprintf(writer, "helm-diff-summary version: %s\n", buildInfo); err != nil {
		log.Fatal(err)
	}
}
