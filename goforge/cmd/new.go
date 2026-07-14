package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new GoForge project",
	Long:  `Create a Python package with Go extensions`,
	Args:  cobra.ExactArgs(1),
	RunE:  runNew,
}

var templateType string

func init() {
	newCmd.Flags().StringVar(&templateType, "template", "extension", "Template type: extension, binary, both")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	fmt.Printf("Creating new GoForge project: %s\n", name)
	
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := createGoMod(name, name); err != nil {
		return err
	}
	
	if err := createPyproject(name, name); err != nil {
		return err
	}
	
	if err := createPkgDir(name, name); err != nil {
		return err
	}
	
	if err := createCmdDir(name, name); err != nil {
		return err
	}
	
	if err := createReadme(name); err != nil {
		return err
	}
	
	fmt.Printf("Created project structure in %s/\n", name)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  goforge develop")
	
	return nil
}

func createGoMod(dir, moduleName string) error {
	content := fmt.Sprintf(`module github.com/user/%s

go 1.22
`, moduleName)
	return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644)
}

func createPyproject(dir, moduleName string) error {
	content := fmt.Sprintf(`[build-system]
requires = ["goforge>=0.1.0"]
build-backend = "goforge.build"

[project]
name = "%s"
version = "0.1.0"
description = "A Python package with Go extensions"
requires-python = ">=3.8"
dependencies = ["cffi>=1.0.0"]

[tool.goforge]
module = "github.com/user/%s"
bindings = "cffi"
`, moduleName, moduleName)
	return os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(content), 0644)
}

func createPkgDir(dir, moduleName string) error {
	pkgDir := filepath.Join(dir, "pkg", "core")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return err
	}
	
	content := `package core

//export Add
func Add(a, b int64) int64 {
	return a + b
}

//export Multiply
func Multiply(a, b int64) int64 {
	return a * b
}
`
	return os.WriteFile(filepath.Join(pkgDir, "core.go"), []byte(content), 0644)
}

func createCmdDir(dir, moduleName string) error {
	cmdDir := filepath.Join(dir, "cmd")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}
	
	goModule := fmt.Sprintf("github.com/user/%s", moduleName)
	content := fmt.Sprintf(`package main

// #include <stdint.h>
// #include <stdlib.h>
import "C"
import "unsafe"

import (
	"%s/pkg/core"
)

func main() {}

//export Add
func Add(a C.int64_t, b C.int64_t) C.int64_t {
	return C.int64_t(core.Add(int64(a), int64(b)))
}

//export Multiply
func Multiply(a C.int64_t, b C.int64_t) C.int64_t {
	return C.int64_t(core.Multiply(int64(a), int64(b)))
}

var _ = unsafe.Pointer(nil)
`, goModule)
	return os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(content), 0644)
}

func createReadme(dir string) error {
	content := fmt.Sprintf(`# %s

A Python package with Go extensions built with GoForge.

## Installation

`+"```bash"+`
pip install %s
`+"```"+`

## Usage

`+"```python"+`
import %s

result = %s.Add(2, 3)
print(result)  # 5
`+"```"+`
`, dir, dir, strings.ReplaceAll(dir, "-", "_"), strings.ReplaceAll(dir, "-", "_"))
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(content), 0644)
}
