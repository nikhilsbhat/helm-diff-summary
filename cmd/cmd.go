package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var cmd *cobra.Command

//nolint:gochecknoinits
func init() {
	cmd = getRootCommand()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Main will take the workload of executing/starting the cli, when the command is passed to it.
func Main() {
	if err := execute(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

// execute will actually execute the cli by taking the arguments passed to cli.
func execute(args []string) error {
	cmd.SetArgs(args)

	if _, err := cmd.ExecuteC(); err != nil {
		return err
	}

	return nil
}

func getUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}{{printf "\n" }}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}{{printf "\n" }}
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{printf "\n"}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{printf "\n"}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}{{printf "\n"}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}{{printf "\n"}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
{{if .HasAvailableSubCommands}}{{printf "\n"}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
{{printf "\n"}}`
}
