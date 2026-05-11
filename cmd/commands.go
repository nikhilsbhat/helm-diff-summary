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

func GetRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:     "helm-diff-summary [command]",
		Short:   "A Terraform-style summarizer for helm diff output",
		Long:    `A utility that reads the helm diff plugin's output and summarizes its output in a Terraform style`,
		PreRunE: setCLIClient,
		Args:    cobra.NoArgs,
		Example: `helm diff ... --output diff | helm-diff-summary
helm diff upgrade sample ../helm-images/example/chart/sample  | ./helm-diff-summary -o yaml
helm diff upgrade sample ../helm-images/example/chart/sample  | ./helm-diff-summary -o json`,
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
			input := renderer.New(resources, summary, cliCfg.noColor)

			switch cliCfg.outputFormat {
			case "json", "j":
				return input.JSON()
			case "yaml", "y":
				return input.YAML()
			default:
				if err = input.Table(); err != nil {
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
