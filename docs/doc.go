package main

import (
	"log"

	"github.com/nikhilsbhat/helm-diff-summary/cmd"
	"github.com/spf13/cobra/doc"
)

//go:generate go run github.com/nikhilsbhat/helm-diff-summary/docs
func main() {
	commands := cmd.GetRootCommand()

	if err := doc.GenMarkdownTree(commands, "doc"); err != nil {
		log.Fatal(err)
	}
}
