package config

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var defaultConfig []byte

type Config struct {
	Categories []Category `yaml:"categories" json:"categories"`
}

type Category struct {
	Name       string   `yaml:"name" json:"name"`
	Extensions []string `yaml:"extensions" json:"extensions"`
}

func LoadOrDefault(path string) (*Config, error) {
	if path == "" {
		return LoadDefault()
	}

	return Load(path)
}

func LoadDefault() (*Config, error) {
	return parse(defaultConfig)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg, err := parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, nil
}

func parse(data []byte) (*Config, error) {
	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Categories) == 0 {
		return fmt.Errorf("no categories configured")
	}

	names := make(map[string]struct{})
	extensions := make(map[string]string)

	for i := range c.Categories {
		category := &c.Categories[i]

		category.Name = strings.ToUpper(
			strings.TrimSpace(category.Name),
		)

		if category.Name == "" {
			return fmt.Errorf("category name cannot be empty")
		}

		if _, exists := names[category.Name]; exists {
			return fmt.Errorf(
				"duplicate category %q",
				category.Name,
			)
		}

		names[category.Name] = struct{}{}

		if len(category.Extensions) == 0 {
			return fmt.Errorf(
				"category %q has no extensions",
				category.Name,
			)
		}

		for j := range category.Extensions {
			ext := strings.ToLower(
				strings.TrimSpace(category.Extensions[j]),
			)

			if ext == "" {
				return fmt.Errorf(
					"empty extension in category %q",
					category.Name,
				)
			}

			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}

			if previous, exists := extensions[ext]; exists {
				return fmt.Errorf(
					"extension %q belongs to both %q and %q",
					ext,
					previous,
					category.Name,
				)
			}

			extensions[ext] = category.Name
			category.Extensions[j] = ext
		}
	}

	return nil
}

func (c *Config) ExtensionMap() map[string]string {
	result := make(map[string]string)

	for _, category := range c.Categories {
		for _, ext := range category.Extensions {
			result[ext] = category.Name
		}
	}

	return result
}
