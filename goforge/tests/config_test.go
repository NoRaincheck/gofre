package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grackin/goforge/internal/config"
)

func TestLoadValidConfig(t *testing.T) {
	cfg, err := config.Load("testdata")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Project.Name != "test-project" {
		t.Errorf("expected name 'test-project', got '%s'", cfg.Project.Name)
	}
	if cfg.Project.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got '%s'", cfg.Project.Version)
	}
	if cfg.Project.Description != "A test project" {
		t.Errorf("expected description 'A test project', got '%s'", cfg.Project.Description)
	}
	if cfg.Project.RequiresPy != ">=3.8" {
		t.Errorf("expected requires-python '>=3.8', got '%s'", cfg.Project.RequiresPy)
	}
	if cfg.Tool.GoForge.Module != "github.com/user/test-project" {
		t.Errorf("expected module 'github.com/user/test-project', got '%s'", cfg.Tool.GoForge.Module)
	}
	if cfg.Tool.GoForge.Bindings != "cffi" {
		t.Errorf("expected bindings 'cffi', got '%s'", cfg.Tool.GoForge.Bindings)
	}
	if cfg.Tool.GoForge.PkgDir != "pkg/core" {
		t.Errorf("expected pkg-dir 'pkg/core', got '%s'", cfg.Tool.GoForge.PkgDir)
	}
}

func TestLoadDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	content := `[project]
name = "my-app"
version = "0.1.0"

[tool.goforge]
module = "github.com/user/my-app"
`
	if err := os.WriteFile(pyproject, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write pyproject.toml: %v", err)
	}

	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Tool.GoForge.Bindings != "cffi" {
		t.Errorf("expected default bindings 'cffi', got '%s'", cfg.Tool.GoForge.Bindings)
	}
	if cfg.Tool.GoForge.PkgDir != "pkg" {
		t.Errorf("expected default pkg-dir 'pkg', got '%s'", cfg.Tool.GoForge.PkgDir)
	}
}

func TestLoadDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	content := `[project]
name = "deps-test"
version = "1.0.0"
dependencies = ["numpy>=1.20", "pandas>=1.3"]

[tool.goforge]
module = "github.com/user/deps-test"
`
	if err := os.WriteFile(pyproject, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write pyproject.toml: %v", err)
	}

	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Project.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(cfg.Project.Dependencies))
	}
	if cfg.Project.Dependencies[0] != "numpy>=1.20" {
		t.Errorf("expected 'numpy>=1.20', got '%s'", cfg.Project.Dependencies[0])
	}
}

func TestValidateMissingName(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	content := `[project]
version = "1.0.0"
`
	if err := os.WriteFile(pyproject, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write pyproject.toml: %v", err)
	}

	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	err = cfg.Validate()
	if err == nil {
		t.Error("expected validation error for missing name")
	}

	validationErr, ok := err.(*config.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "name" {
		t.Errorf("expected field 'name', got '%s'", validationErr.Field)
	}
}

func TestValidateMissingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	content := `[project]
name = "test"
`
	if err := os.WriteFile(pyproject, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write pyproject.toml: %v", err)
	}

	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	err = cfg.Validate()
	if err == nil {
		t.Error("expected validation error for missing version")
	}

	validationErr, ok := err.(*config.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "version" {
		t.Errorf("expected field 'version', got '%s'", validationErr.Field)
	}
}

func TestValidateValid(t *testing.T) {
	cfg, err := config.Load("testdata")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no validation error, got %v", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := config.Load(tmpDir)
	if err == nil {
		t.Error("expected error for missing pyproject.toml")
	}
}

func TestLoadInvalidToml(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	if err := os.WriteFile(pyproject, []byte("not valid toml {{{"), 0644); err != nil {
		t.Fatalf("failed to write pyproject.toml: %v", err)
	}

	_, err := config.Load(tmpDir)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestLoadBinaries(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	content := `[project]
name = "multi-bin"
version = "1.0.0"

[tool.goforge]
module = "github.com/user/multi-bin"
binaries = ["cli", "server", "worker"]
`
	if err := os.WriteFile(pyproject, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write pyproject.toml: %v", err)
	}

	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Tool.GoForge.Binaries) != 3 {
		t.Fatalf("expected 3 binaries, got %d", len(cfg.Tool.GoForge.Binaries))
	}
	if cfg.Tool.GoForge.Binaries[0] != "cli" {
		t.Errorf("expected 'cli', got '%s'", cfg.Tool.GoForge.Binaries[0])
	}
}

func TestLoadBuildTags(t *testing.T) {
	cfg, err := config.Load("testdata")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Tool.GoForge.BuildTags) != 1 {
		t.Fatalf("expected 1 build tag, got %d", len(cfg.Tool.GoForge.BuildTags))
	}
	if cfg.Tool.GoForge.BuildTags[0] != "no_pocketpy" {
		t.Errorf("expected 'no_pocketpy', got '%s'", cfg.Tool.GoForge.BuildTags[0])
	}
}

func TestLoadBuildTagsDefault(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	content := `[project]
name = "no-tags"
version = "1.0.0"

[tool.goforge]
module = "github.com/user/no-tags"
`
	if err := os.WriteFile(pyproject, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write pyproject.toml: %v", err)
	}

	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Tool.GoForge.BuildTags) != 0 {
		t.Errorf("expected 0 build tags by default, got %d", len(cfg.Tool.GoForge.BuildTags))
	}
}

func TestValidationErrorString(t *testing.T) {
	err := &config.ValidationError{Field: "name", Message: "is required"}
	expected := "name: is required"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}
