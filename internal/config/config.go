package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var defaultConfig []byte

type Config struct {
	ToolsDir   string          `yaml:"tools_dir" json:"toolsDir"`
	Tools      map[string]Tool `yaml:"tools" json:"tools"`
	Categories []Category      `yaml:"categories" json:"categories"`

	baseDir string
}

type Tool struct {
	Type    string `yaml:"type" json:"type"`
	Command string `yaml:"command" json:"command"`
	Timeout int    `yaml:"timeout" json:"timeout"`
}

type Analyzer struct {
	Name string   `yaml:"name" json:"name"`
	Tool string   `yaml:"tool" json:"tool"`
	Args []string `yaml:"args" json:"args"`
}

type Category struct {
	Name       string     `yaml:"name" json:"name"`
	Extensions []string   `yaml:"extensions" json:"extensions"`
	Analyzers  []Analyzer `yaml:"analyzers" json:"analyzers"`
}

func defaultBaseDir() (string, error) {
	if value := os.Getenv("DOCSCAN_HOME"); value != "" {
		return filepath.Abs(value)
	}

	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(executable), nil
}

func LoadOrDefault(path string) (*Config, error) {
	if path == "" {
		return LoadDefault()
	}

	return Load(path)
}

func LoadDefault() (*Config, error) {
	baseDir, err := defaultBaseDir()
	if err != nil {
		return nil, fmt.Errorf(
			"resolve application directory: %w",
			err,
		)
	}

	return parse(
		defaultConfig,
		baseDir,
	)
}

func Load(path string) (*Config, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve config path %q: %w",
			path,
			err,
		)
	}

	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf(
			"read config %q: %w",
			absolutePath,
			err,
		)
	}

	cfg, err := parse(
		data,
		filepath.Dir(absolutePath),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse config %q: %w",
			absolutePath,
			err,
		)
	}

	return cfg, nil
}

func parse(
	data []byte,
	baseDir string,
) (*Config, error) {

	var cfg Config

	if err := yaml.Unmarshal(
		data,
		&cfg,
	); err != nil {
		return nil, fmt.Errorf(
			"parse YAML: %w",
			err,
		)
	}

	cfg.baseDir = baseDir

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf(
			"invalid config: %w",
			err,
		)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	c.ToolsDir = strings.TrimSpace(c.ToolsDir)

	if c.ToolsDir == "" {
		return fmt.Errorf("tools_dir cannot be empty")
	}

	if len(c.Categories) == 0 {
		return fmt.Errorf("no categories configured")
	}

	names := make(map[string]struct{})
	extensions := make(map[string]string)

	for name, tool := range c.Tools {
		name = strings.TrimSpace(name)

		if name == "" {
			return fmt.Errorf("tool name cannot be empty")
		}

		if tool.Type == "" {
			return fmt.Errorf(
				"tool %q has no type",
				name,
			)
		}

		if tool.Type != "command" {
			return fmt.Errorf(
				"tool %q: unsupported type %q",
				name,
				tool.Type,
			)
		}

		if strings.TrimSpace(tool.Command) == "" {
			return fmt.Errorf(
				"tool %q has no command",
				name,
			)
		}

		if tool.Timeout < 0 {
			return fmt.Errorf(
				"tool %q has invalid timeout %d",
				name,
				tool.Timeout,
			)
		}
	}

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

		for j, analyzer := range category.Analyzers {
			analyzer.Name = strings.TrimSpace(analyzer.Name)
			analyzer.Tool = strings.TrimSpace(analyzer.Tool)

			if analyzer.Name == "" {
				return fmt.Errorf(
					"category %q analyzer %d has no name",
					category.Name,
					j,
				)
			}

			if analyzer.Tool == "" {
				return fmt.Errorf(
					"category %q analyzer %q has no tool",
					category.Name,
					analyzer.Name,
				)
			}

			if _, exists := c.Tools[analyzer.Tool]; !exists {
				return fmt.Errorf(
					"category %q analyzer %q references unknown tool %q",
					category.Name,
					analyzer.Name,
					analyzer.Tool,
				)
			}

			category.Analyzers[j] = analyzer
		}
	}

	return nil
}

func (c *Config) ToolsPath() string {
	if filepath.IsAbs(c.ToolsDir) {
		return filepath.Clean(c.ToolsDir)
	}

	return filepath.Join(
		c.baseDir,
		c.ToolsDir,
	)
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

func (c *Config) Category(name string) (*Category, bool) {
	name = strings.ToUpper(strings.TrimSpace(name))

	for i := range c.Categories {
		if c.Categories[i].Name == name {
			return &c.Categories[i], true
		}
	}

	return nil, false
}
