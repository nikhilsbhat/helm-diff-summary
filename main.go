package main

import (
	"log"
	"os"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/renderer"
)

func main() {
	resources, err := parser.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	renderer.Render(resources)
}
