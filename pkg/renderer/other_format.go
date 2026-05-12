package renderer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/parser"
	"github.com/nikhilsbhat/helm-diff-summary/pkg/policy"
)

// Output holds the necessary info to be printed in YAML/JSON format.
type Output struct {
	Plan       Summary               `yaml:"plan"                 json:"plan"`
	Resources  []parser.ResourceDiff `yaml:"resources,omitempty"  json:"resources,omitempty"`
	Violations []policy.Violation    `yaml:"violations,omitempty" json:"violations,omitempty"`
}

// JSON renders output in JSON format.
func (input *Input) JSON() error {
	output := Output{
		Plan:       input.summary,
		Resources:  input.resources,
		Violations: input.violations,
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
		Plan:       input.summary,
		Resources:  input.resources,
		Violations: input.violations,
	}

	encoder := yaml.NewEncoder(os.Stdout)

	defer func(encoder *yaml.Encoder) {
		if err := encoder.Close(); err != nil {

		}
	}(encoder)

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to render json: %w", err)
	}

	return nil
}
