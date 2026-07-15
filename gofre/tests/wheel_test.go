package tests

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grackin/gofre/internal/build"
)

func TestWheelBuilderBuild(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	outputDir := filepath.Join(tmpDir, "dist")

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Create a fake file directly in the build directory
	if err := os.WriteFile(filepath.Join(buildDir, "__init__.py"), []byte("# test"), 0644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	config := &build.Config{
		Name:        "test-project",
		Version:     "1.0.0",
		PkgName:     "test_pkg",
		LibName:     "_binding.so",
		PythonTag:   "cp39",
		AbiTag:      "cp39",
		PlatformTag: "manylinux_2_17_x86_64",
	}

	wb := build.NewWheelBuilder(config, buildDir, outputDir)
	if err := wb.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify wheel file exists
	wheelName := "test_pkg-1.0.0-cp39-cp39-manylinux_2_17_x86_64.whl"
	wheelPath := filepath.Join(outputDir, wheelName)
	if _, err := os.Stat(wheelPath); os.IsNotExist(err) {
		t.Fatalf("wheel file not created: %s", wheelPath)
	}

	// Open and verify wheel contents
	r, err := zip.OpenReader(wheelPath)
	if err != nil {
		t.Fatalf("failed to open wheel: %v", err)
	}
	defer r.Close()

	expectedFiles := []string{
		"test_pkg-1.0.0.dist-info/METADATA",
		"test_pkg-1.0.0.dist-info/WHEEL",
		"test_pkg-1.0.0.dist-info/RECORD",
		"test_pkg/__init__.py",
	}

	fileNames := make(map[string]bool)
	for _, f := range r.File {
		fileNames[f.Name] = true
	}

	for _, expected := range expectedFiles {
		if !fileNames[expected] {
			t.Errorf("wheel missing expected file: %s", expected)
		}
	}
}

func TestWheelMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	outputDir := filepath.Join(tmpDir, "dist")

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	config := &build.Config{
		Name:        "my-package",
		Version:     "2.3.4",
		PkgName:     "my_package",
		LibName:     "_binding.so",
		PythonTag:   "cp39",
		AbiTag:      "cp39",
		PlatformTag: "manylinux_2_17_x86_64",
	}

	wb := build.NewWheelBuilder(config, buildDir, outputDir)
	if err := wb.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	wheelPath := filepath.Join(outputDir, "my_package-2.3.4-cp39-cp39-manylinux_2_17_x86_64.whl")
	r, err := zip.OpenReader(wheelPath)
	if err != nil {
		t.Fatalf("failed to open wheel: %v", err)
	}
	defer r.Close()

	// Check METADATA
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "METADATA") {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open METADATA: %v", err)
			}
			buf := make([]byte, 1024)
			n, _ := rc.Read(buf)
			rc.Close()

			metadata := string(buf[:n])
			if !strings.Contains(metadata, "Name: my-package") {
				t.Error("METADATA missing correct Name")
			}
			if !strings.Contains(metadata, "Version: 2.3.4") {
				t.Error("METADATA missing correct Version")
			}
			if !strings.Contains(metadata, "Requires-Python: >=3.8") {
				t.Error("METADATA missing Requires-Python")
			}
			break
		}
	}
}

func TestWheelFileContents(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	outputDir := filepath.Join(tmpDir, "dist")

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatalf("failed to create build dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Create test files
	pkgDir := filepath.Join(buildDir, "test_pkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create pkg dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	config := &build.Config{
		Name:        "test",
		Version:     "1.0.0",
		PkgName:     "test_pkg",
		LibName:     "_binding.so",
		PythonTag:   "cp39",
		AbiTag:      "cp39",
		PlatformTag: "manylinux_2_17_x86_64",
	}

	wb := build.NewWheelBuilder(config, buildDir, outputDir)
	if err := wb.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	wheelPath := filepath.Join(outputDir, "test_pkg-1.0.0-cp39-cp39-manylinux_2_17_x86_64.whl")
	r, err := zip.OpenReader(wheelPath)
	if err != nil {
		t.Fatalf("failed to open wheel: %v", err)
	}
	defer r.Close()

	// Check __init__.py content
	for _, f := range r.File {
		if f.Name == "test_pkg/__init__.py" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open __init__.py: %v", err)
			}
			buf := make([]byte, 1024)
			n, _ := rc.Read(buf)
			rc.Close()

			content := string(buf[:n])
			if content != "test content" {
				t.Errorf("expected 'test content', got '%s'", content)
			}
			break
		}
	}
}

func TestWheelBuilderWheelFilename(t *testing.T) {
	config := &build.Config{
		PkgName:     "my_pkg",
		Version:     "1.2.3",
		PythonTag:   "cp39",
		AbiTag:      "cp39",
		PlatformTag: "macosx_11_0_arm64",
	}

	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	outputDir := filepath.Join(tmpDir, "dist")
	os.MkdirAll(buildDir, 0755)
	os.MkdirAll(outputDir, 0755)

	wb := build.NewWheelBuilder(config, buildDir, outputDir)
	if err := wb.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "my_pkg-1.2.3-cp39-cp39-macosx_11_0_arm64.whl"
	if _, err := os.Stat(filepath.Join(outputDir, expected)); os.IsNotExist(err) {
		t.Errorf("expected wheel file '%s' not found", expected)
	}
}
