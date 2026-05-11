package renderer

import "github.com/nikhilsbhat/helm-diff-summary/pkg/parser"

// Summary holds information on the diff summary.
type Summary struct {
	Creates int            `yaml:"creates,omitempty" json:"creates,omitempty"`
	Updates int            `yaml:"updates,omitempty" json:"updates,omitempty"`
	Deletes int            `yaml:"deletes,omitempty" json:"deletes,omitempty"`
	ByKind  map[string]int `yaml:"byKind,omitempty"  json:"byKind,omitempty"`
}

// BuildSummary builds summary on the diff resources.
func BuildSummary(resources []parser.ResourceDiff) Summary {
	summary := Summary{
		ByKind: map[string]int{},
	}

	for _, r := range resources {
		summary.ByKind[r.Kind]++

		switch r.ChangeType {
		case parser.Create:
			summary.Creates++

		case parser.Update:
			summary.Updates++

		case parser.Delete:
			summary.Deletes++
		}
	}

	return summary
}
