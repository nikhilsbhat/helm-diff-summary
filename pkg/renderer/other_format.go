package renderer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
	"gopkg.in/yaml.v3"
)

// Output holds the necessary info to be printed in YAML/JSON format.
type Output struct {
	Plan      Summary               `yaml:"plan,omitzero"       json:"plan,omitzero"`
	Resources []parser.ResourceDiff `yaml:"resources,omitempty" json:"resources,omitempty"`
}

// RenderJSON renders output in JSON format.
func RenderJSON(resources []parser.ResourceDiff, summary Summary) error {
	output := Output{
		Plan:      summary,
		Resources: resources,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to render json: %w", err)
	}

	return nil
}

// RenderYAML renders output in YAML format.
func RenderYAML(resources []parser.ResourceDiff, summary Summary) error {
	output := Output{
		Plan:      summary,
		Resources: resources,
	}

	encoder := yaml.NewEncoder(os.Stdout)

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to render json: %w", err)
	}

	return nil
}
