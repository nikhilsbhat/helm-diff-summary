package cmd

import (
	"github.com/nikhilsbhat/helm-diff-summary/pkg/log"
	"github.com/spf13/cobra"
)

func setCLIClient(_ *cobra.Command, _ []string) error {
	logger = log.SetLogger(cliCfg.logLevel)

	return nil
}
