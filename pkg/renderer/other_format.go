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
	Plan      Summary               `yaml:"plan"                json:"plan"`
	Resources []parser.ResourceDiff `yaml:"resources,omitempty" json:"resources,omitempty"`
}

// JSON renders output in JSON format.
func (input *Input) JSON() error {
	output := Output{
		Plan:      input.summary,
		Resources: input.resources,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to render json: %w", err)
	}

	return nil
}

// YAML renders output in YAML format.
func (input *Input) YAML() error {
	output := Output{
		Plan:      input.summary,
		Resources: input.resources,
	}

	encoder := yaml.NewEncoder(os.Stdout)

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to render json: %w", err)
	}

	return nil
}
