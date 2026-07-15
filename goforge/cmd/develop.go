package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grackin/goforge/internal/bindings"
	"github.com/grackin/goforge/internal/build"
	"github.com/grackin/goforge/internal/config"
	"github.com/spf13/cobra"
)

var developCmd = &cobra.Command{
	Use:   "develop",
	Short: "Build and install in current virtualenv",
	Long:  `Build the package and install it in the current virtual environment for development`,
	RunE:  runDevelop,
}

func init() {
	developCmd.Flags().BoolVar(&release, "release", false, "Build in release mode")
	developCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for build artifacts (default: ./build and ./dist)")
	rootCmd.AddCommand(developCmd)
}

func runDevelop(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	
	fmt.Println("Developing GoForge project...")
	
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
	
	venvDir, err := build.FindVenvDir()
	if err != nil {
		return fmt.Errorf("no virtual environment found: %w", err)
	}
	
	fmt.Printf("Using virtual environment: %s\n", venvDir)
	
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
	
	wheelBuilder := build.NewWheelBuilder(wheelConfig, pythonPkgDir, "")
	if err := wheelBuilder.Install(venvDir); err != nil {
		return fmt.Errorf("failed to install in virtualenv: %w", err)
	}
	
	fmt.Println("Development install complete!")
	fmt.Printf("You can now use: python -c \"import %s\"\n", pkgName)
	
	return nil
}
