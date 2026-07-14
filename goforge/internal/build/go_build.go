package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type GoBuilder struct {
	Module string
	Dir    string
}

func NewGoBuilder(module, dir string) *GoBuilder {
	return &GoBuilder{
		Module: module,
		Dir:    dir,
	}
}

func (b *GoBuilder) BuildSharedLib(output string) error {
	args := []string{
		"build",
		"-buildmode=c-shared",
		"-o", output,
		"./cmd/",
	}
	
	cmd := exec.Command("go", args...)
	cmd.Dir = b.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Printf("Building shared library: go %s\n", strings.Join(args, " "))
	return cmd.Run()
}

func (b *GoBuilder) BuildBinary(output string) error {
	args := []string{
		"build",
		"-o", output,
		"./cmd/...",
	}
	
	cmd := exec.Command("go", args...)
	cmd.Dir = b.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Printf("Building binary: go %s\n", strings.Join(args, " "))
	return cmd.Run()
}

func (b *GoBuilder) BuildForPlatform(output string, goos, goarch string) error {
	env := os.Environ()
	env = append(env, "GOOS="+goos, "GOARCH="+goarch)
	
	args := []string{
		"build",
		"-buildmode=c-shared",
		"-o", output,
		"./cmd/",
	}
	
	cmd := exec.Command("go", args...)
	cmd.Dir = b.Dir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Printf("Cross-compiling for %s/%s\n", goos, goarch)
	return cmd.Run()
}

func GetLibExtension() string {
	switch runtime.GOOS {
	case "windows":
		return ".dll"
	case "darwin":
		return ".dylib"
	default:
		return ".so"
	}
}

func GetPlatformTag() string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	
	switch os {
	case "linux":
		if arch == "amd64" {
			return "manylinux_2_17_x86_64"
		}
		if arch == "arm64" {
			return "manylinux_2_17_aarch64"
		}
	case "darwin":
		if arch == "amd64" {
			return "macosx_10_9_x86_64"
		}
		if arch == "arm64" {
			return "macosx_11_0_arm64"
		}
	case "windows":
		if arch == "amd64" {
			return "win_amd64"
		}
		if arch == "386" {
			return "win32"
		}
	}
	
	return fmt.Sprintf("%s_%s", os, arch)
}

func GetPythonTag() string {
	cmd := exec.Command("python3", "-c", "import sys; print(f'cp{sys.version_info.major}{sys.version_info.minor}')")
	out, err := cmd.Output()
	if err != nil {
		return "cp39"
	}
	tag := strings.TrimSpace(string(out))
	if len(tag) > 0 {
		return tag
	}
	return "cp39"
}

func EnsureGoInstalled() error {
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Go is not installed or not in PATH")
	}
	return nil
}

func EnsureGoModule(dir string) error {
	goMod := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goMod); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found in %s", dir)
	}
	return nil
}
