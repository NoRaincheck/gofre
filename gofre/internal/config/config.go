package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Project struct {
		Name        string   `toml:"name"`
		Version     string   `toml:"version"`
		Description string   `toml:"description"`
		RequiresPy  string   `toml:"requires-python"`
		Dependencies []string `toml:"dependencies"`
	} `toml:"project"`
	
	Tool struct {
		GoFre struct {
			Module    string   `toml:"module"`
			Bindings  string   `toml:"bindings"`
			PkgDir    string   `toml:"pkg-dir"`
			Binaries  []string `toml:"binaries"`
			BuildTags []string `toml:"build-tags"`
		} `toml:"gofre"`
	} `toml:"tool"`
}

func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, "pyproject.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	
	if cfg.Tool.GoFre.Bindings == "" {
		cfg.Tool.GoFre.Bindings = "cffi"
	}
	if cfg.Tool.GoFre.PkgDir == "" {
		cfg.Tool.GoFre.PkgDir = "pkg"
	}
	
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Project.Name == "" {
		return &ValidationError{Field: "name", Message: "project name is required"}
	}
	if c.Project.Version == "" {
		return &ValidationError{Field: "version", Message: "project version is required"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
