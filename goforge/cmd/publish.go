package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish the package to PyPI",
	Long:  `Build and publish the package to PyPI or TestPyPI`,
	RunE:  runPublish,
}

var test bool

func init() {
	publishCmd.Flags().BoolVar(&test, "test", false, "Publish to TestPyPI")
	rootCmd.AddCommand(publishCmd)
}

func runPublish(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	
	fmt.Println("Publishing GoForge project...")
	
	buildCmd := exec.Command("go", "forge", "build", "--release")
	buildCmd.Dir = dir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build package: %w", err)
	}
	
	distDir := filepath.Join(dir, "dist")
	wheels, err := filepath.Glob(filepath.Join(distDir, "*.whl"))
	if err != nil {
		return err
	}
	
	if len(wheels) == 0 {
		return fmt.Errorf("no wheels found in %s", distDir)
	}
	
	repo := "pypi"
	if test {
		repo = "testpypi"
	}
	
	for _, wheel := range wheels {
		fmt.Printf("Uploading %s to %s...\n", filepath.Base(wheel), repo)
		
		args := []string{"-m", "twine", "upload", "--repository", repo, wheel}
		if test {
			args = append(args, "--repository-url", "https://test.pypi.org/legacy/")
		}
		
		uploadCmd := exec.Command("python3", args...)
		uploadCmd.Dir = dir
		uploadCmd.Stdout = os.Stdout
		uploadCmd.Stderr = os.Stderr
		
		if err := uploadCmd.Run(); err != nil {
			return fmt.Errorf("failed to upload wheel: %w", err)
		}
	}
	
	fmt.Println("Publish complete!")
	
	if test {
		name := getPackageName()
		version := getPackageVersion()
		fmt.Printf("\nInstall with:\n  pip install %s==%s --index-url https://test.pypi.org/simple/\n", name, version)
	}
	
	return nil
}

func getPackageVersion() string {
	pyproject := filepath.Join(".", "pyproject.toml")
	if _, err := os.Stat(pyproject); os.IsNotExist(err) {
		return "0.1.0"
	}
	
	content, err := os.ReadFile(pyproject)
	if err != nil {
		return "0.1.0"
	}
	
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if line == "[project]" {
			for j := i + 1; j < len(lines); j++ {
				if strings.HasPrefix(lines[j], "version") {
					parts := strings.SplitN(lines[j], "=", 2)
					if len(parts) == 2 {
						return strings.Trim(strings.TrimSpace(parts[1]), "\"")
					}
				}
				if len(lines[j]) > 0 && lines[j][0] == '[' {
					break
				}
			}
		}
	}
	
	return "0.1.0"
}
