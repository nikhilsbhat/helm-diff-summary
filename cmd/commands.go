package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

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
helm diff upgrade sample ../helm-images/example/chart/sample  | ./helm-diff-summary -o json
helm diff upgrade sample ../helm-images/example/chart/sample  | ./helm-diff-summary --fail-on high
helm diff upgrade sample ../helm-images/example/chart/sample  | ./helm-diff-summary --fail-on-delete
helm diff upgrade sample ../helm-images/example/chart/sample  | ./helm-diff-summary --notify slack,gchat`,
		RunE: run,
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
