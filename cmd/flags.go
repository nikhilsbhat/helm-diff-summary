package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// Config holds the information of the cli config.
type config struct {
	logLevel, outputFormat, failOnSeverity string
	noColor, showVersion, failOnDelete     bool
}

var (
	cliCfg = new(config)
	logger *slog.Logger
	_      = logger
)

// Registers all global flags to utility.
func registerCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cliCfg.logLevel, "log-level", "l", "INFO",
		"log level for the helm-diff-summary")
	cmd.PersistentFlags().StringVarP(&cliCfg.outputFormat, "output", "o", "",
		"output format to which the output needs to be rendered")
	cmd.PersistentFlags().BoolVarP(&cliCfg.failOnDelete, "fail-on-delete", "", false,
		"when set, fail if deletes are detected")
	cmd.PersistentFlags().StringVarP(&cliCfg.failOnSeverity, "fail-on", "", "",
		"fail when violations reach severity: low|medium|high|critical")
	cmd.PersistentFlags().BoolVarP(&cliCfg.showVersion, "version", "v", false,
		"when set, prints the version of the utility")
	cmd.PersistentFlags().BoolVarP(&cliCfg.noColor, "no-color", "", false,
		"when set, renders the output with no color")
}
