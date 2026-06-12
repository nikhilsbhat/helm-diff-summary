package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateDocs(t *testing.T) {
	outputDir := t.TempDir()

	if err := generateDocs(outputDir); err != nil {
		t.Fatalf("generateDocs returned error: %v", err)
	}

	path := filepath.Join(outputDir, "helm-diff-summary.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected generated docs at %s: %v", path, err)
	}
}

func TestMainGeneratesDocs(t *testing.T) {
	originalWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkingDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "doc"), 0o700); err != nil {
		t.Fatalf("failed to create doc directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	main()

	if _, err := os.Stat(filepath.Join(dir, "doc", "helm-diff-summary.md")); err != nil {
		t.Fatalf("expected generated docs: %v", err)
	}
}
