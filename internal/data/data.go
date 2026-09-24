package data

import (
	"fmt"
	"os"

	"github.com/dragstor/awesome-reliability-tools/internal/schema"
	"gopkg.in/yaml.v3"
)

// LoadCatalog reads the central feature definitions from a YAML file.
func LoadCatalog(path string) (*schema.Catalog, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading catalog %q: %w", path, err)
	}
	var defs []schema.FeatureDef
	if err := yaml.Unmarshal(b, &defs); err != nil {
		return nil, fmt.Errorf("parsing catalog %q: %w", path, err)
	}
	return &schema.Catalog{Features: defs}, nil
}

// LoadTools reads the tool dataset from a YAML file.
func LoadTools(path string) ([]schema.Tool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading tools %q: %w", path, err)
	}
	var tools []schema.Tool
	if err := yaml.Unmarshal(b, &tools); err != nil {
		return nil, fmt.Errorf("parsing tools %q: %w", path, err)
	}
	return tools, nil
}
