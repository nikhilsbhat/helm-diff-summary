package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/renderer"
	"github.com/nikhilsbhat/helm-diff-summary/version"
)

func versionInfo(writer io.Writer) {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		log.Fatalf("failed to fetch version info: %v", err)
	}

	if _, err = fmt.Fprintf(writer, "helm-diff-summary version: %s\n", buildInfo); err != nil {
		log.Fatal(err)
	}
}

func usage(writer io.Writer) {
	if _, err := fmt.Fprintf(writer, `helm-diff-summary

A Terraform-style summarizer for helm diff output.

USAGE:
helm diff ... --output diff | helm-diff-summary

EXAMPLES:
helm diff upgrade sample ./chart \
--allow-unreleased \
--output diff | helm-diff-summary

helm diff upgrade sample ./chart \
-n production \
--output diff | helm-diff-summary

cat diff.txt | helm-diff-summary

FLAGS:
-h, --help       Show help
-v, --version    Show version

DESCRIPTION:
helm-diff-summary reads helm diff output from stdin
and renders a concise deployment summary table.

`); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var showHelp, showVersion bool

	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.BoolVar(&showHelp, "help", false, "show help")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")

	flag.Usage = func() { usage(os.Stdout) }

	flag.Parse()

	switch {
	case showHelp:
		usage(os.Stdout)

		return
	case showVersion:
		versionInfo(os.Stdout)

		return
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
			return
		}

		if _, err = fmt.Fprintln(os.Stderr); err != nil {
			return
		}

		usage(os.Stderr)

		os.Exit(1)
	}

	resources, err := parser.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	if len(resources) == 0 {
		if _, err = fmt.Fprintln(os.Stderr, "no resources detected in helm diff output"); err != nil {
			return
		}

		os.Exit(0)
	}

	renderer.Render(resources)
}
