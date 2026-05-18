package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/notifier"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/policy"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/renderer"
	"github.com/spf13/cobra"
)

func run(_ *cobra.Command, _ []string) error {
	if cliCfg.showVersion {
		versionInfo(os.Stdout)

		return nil
	}

	if err := validateInput(); err != nil {
		return err
	}

	resources, err := parser.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	if len(resources) == 0 {
		return handleNoResources()
	}

	policies, err := policy.New("helm-diff-summary.yaml")
	if err != nil {
		return err
	}

	violations := policies.Evaluate(resources)

	summary := renderer.BuildSummary(resources)

	var builder strings.Builder

	input := renderer.New(resources, violations, summary, &builder, logger, cliCfg.noColor)

	if err = renderOutput(input); err != nil {
		return err
	}

	if err = handleExitConditions(violations, summary); err != nil {
		return err
	}

	return nil
}

func validateInput() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("failed to inspect stdin: %v", err)
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		if _, err = fmt.Fprintln(os.Stderr, "no helm diff input detected"); err != nil {
			return err
		}

		if _, err = fmt.Fprintln(os.Stderr); err != nil {
			return err
		}

		os.Exit(1)
	}

	return nil
}

func handleNoResources() error {
	if _, err := fmt.Fprintln(os.Stderr, "no resources detected in helm diff output"); err != nil {
		return err
	}

	os.Exit(0)

	return nil
}

func renderOutput(input *renderer.Input) error {
	switch cliCfg.outputFormat {
	case "json", "j":
		if err := input.JSON(); err != nil {
			return err
		}

		text := input.GetText()

		fmt.Printf("%s", text)

		return nil
	case "yaml", "y":
		if err := input.YAML(); err != nil {
			return err
		}

		text := input.GetText()

		fmt.Printf("%s", text)

		return nil
	default:
		if err := input.Text(); err != nil {
			return err
		}

		text := input.GetText()

		fmt.Printf("%s", text)

		return sendNotifications(text)
	}
}

func sendNotifications(message string) error {
	return notifier.Notify(message, cliCfg.notifiers)
}

func handleExitConditions(violations []policy.Violation, summary renderer.Summary) error {
	const exitCode = 2

	if cliCfg.failOnSeverity != "" {
		severity, err := parser.ParseSeverity(cliCfg.failOnSeverity)
		if err != nil {
			log.Fatal(err)
		}

		if policy.HasViolationsAtOrAbove(violations, severity) {
			os.Exit(exitCode)
		}
	}

	if cliCfg.failOnDelete && summary.Deletes > 0 {
		os.Exit(exitCode)
	}

	return nil
}
