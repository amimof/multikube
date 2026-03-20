package config

import (
	"encoding/json"
	"fmt"
	"os"

	types "github.com/amimof/multikube/api/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

// LoadFromFile reads a YAML configuration file and parses it into
// the protobuf Config type.
func LoadFromFile(path string) (*types.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	return Load(data)
}

// Load parses raw YAML bytes into the protobuf Config type.
// It uses an intermediate YAML→JSON→protojson pipeline because
// proto-generated structs lack yaml struct tags.
func Load(data []byte) (*types.Config, error) {
	// Step 1: YAML → generic map (preserves snake_case keys)
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Step 2: map → JSON bytes
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("converting to JSON: %w", err)
	}

	// Step 3: JSON → protobuf Config via protojson
	// protojson handles Duration strings ("30s"), snake_case field names,
	// and nested message structures.
	cfg := &types.Config{}
	opts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := opts.Unmarshal(jsonBytes, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}
