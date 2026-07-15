package tests

import (
	"os"
	"runtime"
	"testing"

	"github.com/NoRaincheck/gofre/internal/build"
)

func TestGetLibExtension(t *testing.T) {
	ext := build.GetLibExtension()

	switch runtime.GOOS {
	case "windows":
		if ext != ".dll" {
			t.Errorf("expected '.dll' on Windows, got '%s'", ext)
		}
	case "darwin":
		if ext != ".dylib" {
			t.Errorf("expected '.dylib' on macOS, got '%s'", ext)
		}
	default:
		if ext != ".so" {
			t.Errorf("expected '.so' on Linux, got '%s'", ext)
		}
	}
}

func TestGetPlatformTag(t *testing.T) {
	tag := build.GetPlatformTag()

	// Verify it's not empty
	if tag == "" {
		t.Error("GetPlatformTag returned empty string")
	}

	// Verify expected tag for current platform
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			if tag != "manylinux_2_17_x86_64" {
				t.Errorf("expected 'manylinux_2_17_x86_64', got '%s'", tag)
			}
		case "arm64":
			if tag != "manylinux_2_17_aarch64" {
				t.Errorf("expected 'manylinux_2_17_aarch64', got '%s'", tag)
			}
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			if tag != "macosx_10_9_x86_64" {
				t.Errorf("expected 'macosx_10_9_x86_64', got '%s'", tag)
			}
		case "arm64":
			if tag != "macosx_11_0_arm64" {
				t.Errorf("expected 'macosx_11_0_arm64', got '%s'", tag)
			}
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			if tag != "win_amd64" {
				t.Errorf("expected 'win_amd64', got '%s'", tag)
			}
		case "386":
			if tag != "win32" {
				t.Errorf("expected 'win32', got '%s'", tag)
			}
		}
	}
}

func TestGetPythonTag(t *testing.T) {
	tag := build.GetPythonTag()
	if tag == "" {
		t.Error("GetPythonTag returned empty string")
	}
	// Validate format: should be 'cp' followed by major.minor version digits
	if len(tag) < 4 || tag[:2] != "cp" {
		t.Errorf("expected tag starting with 'cp' followed by version digits, got '%s'", tag)
	}
}

func TestEnsureGoInstalled(t *testing.T) {
	if err := build.EnsureGoInstalled(); err != nil {
		t.Skipf("Go not installed, skipping: %v", err)
	}
}

func TestEnsureGoModule(t *testing.T) {
	// Test with existing go.mod
	err := build.EnsureGoModule("..")
	if err != nil {
		t.Errorf("expected no error for project root with go.mod, got %v", err)
	}

	// Test without go.mod
	tmpDir := t.TempDir()
	err = build.EnsureGoModule(tmpDir)
	if err == nil {
		t.Error("expected error for directory without go.mod")
	}
}

func TestNewGoBuilder(t *testing.T) {
	builder := build.NewGoBuilder("github.com/user/test", "/tmp/test")
	if builder == nil {
		t.Fatal("NewGoBuilder returned nil")
	}
}

func TestGetPythonVersion(t *testing.T) {
	ver := build.GetPythonVersion()
	if ver == "" {
		t.Error("GetPythonVersion returned empty string")
	}
	// Should start with "python"
	if len(ver) < 7 || ver[:6] != "python" {
		t.Errorf("expected version starting with 'python', got '%s'", ver)
	}
}

func TestFindVenvDirNoVenv(t *testing.T) {
	tmpDir := t.TempDir()
	// Save and restore working directory and env
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	origVenv := os.Getenv("VIRTUAL_ENV")
	os.Unsetenv("VIRTUAL_ENV")
	defer os.Setenv("VIRTUAL_ENV", origVenv)

	os.Chdir(tmpDir)
	_, err := build.FindVenvDir()
	if err == nil {
		t.Error("expected error when no venv exists")
	}
}

func TestGetPlatformTagCoverage(t *testing.T) {
	// Ensure GetPlatformTag doesn't panic and returns non-empty
	tag := build.GetPlatformTag()
	if tag == "" {
		t.Error("GetPlatformTag returned empty for current platform")
	}
}
