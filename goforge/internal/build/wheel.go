package build

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WheelBuilder struct {
	Config     *Config
	BuildDir   string
	OutputDir  string
}

type Config struct {
	Name        string
	Version     string
	PkgName     string
	LibName     string
	PythonTag   string
	AbiTag      string
	PlatformTag string
}

func NewWheelBuilder(config *Config, buildDir, outputDir string) *WheelBuilder {
	return &WheelBuilder{
		Config:    config,
		BuildDir:  buildDir,
		OutputDir: outputDir,
	}
}

func (w *WheelBuilder) Build() error {
	wheelName := w.getWheelName()
	wheelPath := filepath.Join(w.OutputDir, wheelName)
	
	fmt.Printf("Building wheel: %s\n", wheelName)
	
	zipFile, err := os.Create(wheelPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	if err := w.addMetadata(zipWriter); err != nil {
		return err
	}
	
	if err := w.addWHEEL(zipWriter); err != nil {
		return err
	}
	
	if err := w.addFiles(zipWriter); err != nil {
		return err
	}
	
	if err := w.addRecord(zipWriter); err != nil {
		return err
	}
	
	fmt.Printf("Created wheel: %s\n", wheelPath)
	return nil
}

func (w *WheelBuilder) getWheelName() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s.whl",
		w.Config.PkgName,
		w.Config.Version,
		w.Config.PythonTag,
		w.Config.AbiTag,
		w.Config.PlatformTag,
	)
}

func (w *WheelBuilder) addMetadata(zw *zip.Writer) error {
	metadata := fmt.Sprintf(`Metadata-Version: 2.1
Name: %s
Version: %s
Summary: A Python package with Go extensions
Requires-Python: >=3.8
Requires-Dist: cffi>=1.0.0
`,
		w.Config.Name,
		w.Config.Version,
	)
	
	f, err := zw.Create(fmt.Sprintf("%s-%s.dist-info/METADATA", w.Config.PkgName, w.Config.Version))
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(metadata))
	return err
}

func (w *WheelBuilder) addWHEEL(zw *zip.Writer) error {
	wheel := fmt.Sprintf(`Wheel-Version: 1.0
Generator: goforge
Root-Is-Purelib: false
Tag: %s-%s-%s
`,
		w.Config.PythonTag,
		w.Config.AbiTag,
		w.Config.PlatformTag,
	)
	
	f, err := zw.Create(fmt.Sprintf("%s-%s.dist-info/WHEEL", w.Config.PkgName, w.Config.Version))
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(wheel))
	return err
}

func (w *WheelBuilder) addFiles(zw *zip.Writer) error {
	return filepath.Walk(w.BuildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		
		relPath, err := filepath.Rel(w.BuildDir, path)
		if err != nil {
			return err
		}
		
		archivePath := filepath.Join(w.Config.PkgName, relPath)
		archivePath = filepath.ToSlash(archivePath)
		
		f, err := zw.Create(archivePath)
		if err != nil {
			return err
		}
		
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()
		
		_, err = io.Copy(f, src)
		return err
	})
}

func (w *WheelBuilder) addRecord(zw *zip.Writer) error {
	f, err := zw.Create(fmt.Sprintf("%s-%s.dist-info/RECORD", w.Config.PkgName, w.Config.Version))
	if err != nil {
		return err
	}
	
	_, err = f.Write([]byte(",,\n"))
	return err
}

func (w *WheelBuilder) Install(venvDir string) error {
	libDir := filepath.Join(venvDir, "lib")
	
	pythonVersion := "python3.9"
	targetDir := filepath.Join(libDir, pythonVersion, "site-packages", w.Config.PkgName)
	
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	
	return filepath.Walk(w.BuildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		
		relPath, err := filepath.Rel(w.BuildDir, path)
		if err != nil {
			return err
		}
		
		destPath := filepath.Join(targetDir, relPath)
		destDir := filepath.Dir(destPath)
		
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}
		
		return copyFile(path, destPath)
	})
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

func GetPythonVersion() string {
	cmd := exec.Command("python3", "--version")
	out, err := cmd.Output()
	if err != nil {
		return "python3.9"
	}
	
	version := strings.TrimSpace(string(out))
	version = strings.TrimPrefix(version, "Python ")
	
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return fmt.Sprintf("python%s.%s", parts[0], parts[1])
	}
	
	return "python3.9"
}

func FindVenvDir() (string, error) {
	venv := os.Getenv("VIRTUAL_ENV")
	if venv != "" {
		return venv, nil
	}
	
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	for {
		venvPath := filepath.Join(dir, ".venv")
		if _, err := os.Stat(venvPath); err == nil {
			return venvPath, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	
	return "", fmt.Errorf("no virtual environment found")
}
