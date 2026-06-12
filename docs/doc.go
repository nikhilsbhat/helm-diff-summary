package main

import (
	"log"

	"github.com/nikhilsbhat/helm-diff-summary/cmd"
	"github.com/spf13/cobra/doc"
)

//go:generate go run github.com/nikhilsbhat/helm-diff-summary/docs
func main() {
	if err := generateDocs("doc"); err != nil {
		log.Fatal(err)
	}
}

func generateDocs(outputDir string) error {
	commands := cmd.GetRootCommand()

	return doc.GenMarkdownTree(commands, outputDir)
}
