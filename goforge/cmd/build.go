package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/grackin/goforge/internal/bindings"
	"github.com/grackin/goforge/internal/build"
	"github.com/grackin/goforge/internal/config"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the Python package",
	Long:  `Build the Python package with Go extensions`,
	RunE:  runBuild,
}

var release bool
var outputDir string

func init() {
	buildCmd.Flags().BoolVar(&release, "release", false, "Build in release mode")
	buildCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for build artifacts (default: ./build and ./dist)")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	
	fmt.Println("Building GoForge project...")
	
	if err := build.EnsureGoInstalled(); err != nil {
		return err
	}
	
	if err := build.EnsureGoModule(dir); err != nil {
		return err
	}
	
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	
	if err := cfg.Validate(); err != nil {
		return err
	}
	
	baseDir := dir
	if outputDir != "" {
		baseDir = outputDir
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return err
		}
	}
	
	buildDir := filepath.Join(baseDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return err
	}
	
	libName := "_binding" + build.GetLibExtension()
	libPath := filepath.Join(buildDir, libName)
	
	goBuilder := build.NewGoBuilder(cfg.Tool.GoForge.Module, dir)
	if err := goBuilder.BuildSharedLib(libPath, cfg.Tool.GoForge.BuildTags); err != nil {
		return fmt.Errorf("failed to build Go shared library: %w", err)
	}
	
	pkgDir := filepath.Join(dir, cfg.Tool.GoForge.PkgDir)
	files, err := bindings.ParseDir(pkgDir)
	if err != nil {
		return fmt.Errorf("failed to parse Go files: %w", err)
	}
	
	if len(files) == 0 {
		return fmt.Errorf("no Go files with exported functions found in %s", pkgDir)
	}
	
	pkgName := strings.ReplaceAll(cfg.Project.Name, "-", "_")
	pythonPkgDir := filepath.Join(buildDir, pkgName)
	if err := os.MkdirAll(pythonPkgDir, 0755); err != nil {
		return err
	}
	
	gen := bindings.NewGenerator(pythonPkgDir, pkgName)
	if err := gen.Generate(files); err != nil {
		return fmt.Errorf("failed to generate bindings: %w", err)
	}
	
	// Copy the shared library to the Python package directory
	srcLib := libPath
	dstLib := filepath.Join(pythonPkgDir, libName)
	if err := copyFile(srcLib, dstLib); err != nil {
		return fmt.Errorf("failed to copy shared library: %w", err)
	}
	
	wheelConfig := &build.Config{
		Name:        cfg.Project.Name,
		Version:     cfg.Project.Version,
		PkgName:     pkgName,
		LibName:     libName,
		PythonTag:   build.GetPythonTag(),
		AbiTag:      build.GetPythonTag(),
		PlatformTag: build.GetPlatformTag(),
	}
	
	distDir := filepath.Join(baseDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	
	wheelBuilder := build.NewWheelBuilder(wheelConfig, pythonPkgDir, distDir)
	if err := wheelBuilder.Build(); err != nil {
		return fmt.Errorf("failed to build wheel: %w", err)
	}
	
	fmt.Println("Build complete!")
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}
