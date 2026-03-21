package config

import (
	"encoding/json"
	"fmt"

	types "github.com/amimof/multikube/api/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

// MarshalYAML serialises a protobuf Config to YAML bytes.
// It uses the inverse of the Load pipeline: protojson → JSON → YAML.
// This preserves snake_case field names and Duration formatting ("30s").
func MarshalYAML(cfg *types.Config) ([]byte, error) {
	// Step 1: protobuf → JSON bytes via protojson.
	// UseProtoNames keeps snake_case keys instead of camelCase.
	// EmitDefaultValues=false omits zero-value fields for cleaner output.
	opts := protojson.MarshalOptions{
		UseProtoNames: true,
	}
	jsonBytes, err := opts.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshalling config to JSON: %w", err)
	}

	// Step 2: JSON → generic map (for yaml.Marshal to produce clean YAML).
	var raw map[string]any
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		return nil, fmt.Errorf("parsing JSON intermediate: %w", err)
	}

	// Step 3: map → YAML bytes.
	yamlBytes, err := yaml.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshalling to YAML: %w", err)
	}

	return yamlBytes, nil
}
