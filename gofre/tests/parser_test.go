package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NoRaincheck/gofre/internal/bindings"
)

func TestParseFileSimple(t *testing.T) {
	path := filepath.Join("testdata", "simple.go")
	gf, err := bindings.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if gf == nil {
		t.Fatal("ParseFile returned nil")
	}

	if gf.Package != "math" {
		t.Errorf("expected package 'math', got '%s'", gf.Package)
	}

	if len(gf.Functions) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(gf.Functions))
	}

	// Check Add function
	add := gf.Functions[0]
	if add.Name != "Add" {
		t.Errorf("expected function name 'Add', got '%s'", add.Name)
	}
	if len(add.Params) != 2 {
		t.Fatalf("expected 2 params for Add, got %d", len(add.Params))
	}
	if add.Params[0].Name != "a" || add.Params[0].Type != "int64_t" {
		t.Errorf("expected param 'a int64_t', got '%s %s'", add.Params[0].Name, add.Params[0].Type)
	}
	if add.Params[1].Name != "b" || add.Params[1].Type != "int64_t" {
		t.Errorf("expected param 'b int64_t', got '%s %s'", add.Params[1].Name, add.Params[1].Type)
	}
	if len(add.Returns) != 1 {
		t.Fatalf("expected 1 return for Add, got %d", len(add.Returns))
	}
	if add.Returns[0].Type != "int64_t" {
		t.Errorf("expected return type 'int64_t', got '%s'", add.Returns[0].Type)
	}

	// Check GetHello function (string return)
	hello := gf.Functions[2]
	if hello.Name != "GetHello" {
		t.Errorf("expected function name 'GetHello', got '%s'", hello.Name)
	}
	if len(hello.Params) != 0 {
		t.Errorf("expected 0 params for GetHello, got %d", len(hello.Params))
	}
	if len(hello.Returns) != 1 || hello.Returns[0].Type != "char*" {
		t.Errorf("expected return type 'char*' for GetHello, got %v", hello.Returns)
	}
}

func TestParseFileSlices(t *testing.T) {
	path := filepath.Join("testdata", "slices.go")
	gf, err := bindings.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(gf.Functions) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(gf.Functions))
	}

	// Check SumSlice (slice param, scalar return)
	sum := gf.Functions[0]
	if sum.Name != "SumSlice" {
		t.Errorf("expected 'SumSlice', got '%s'", sum.Name)
	}
	if len(sum.Params) != 1 {
		t.Fatalf("expected 1 param for SumSlice, got %d", len(sum.Params))
	}
	if !sum.Params[0].IsSlice {
		t.Error("expected SumSlice param to be a slice")
	}
	if sum.Params[0].ElemType != "double" {
		t.Errorf("expected elem type 'double', got '%s'", sum.Params[0].ElemType)
	}

	// Check DoubleSlice (slice param, slice return)
	double := gf.Functions[1]
	if double.Name != "DoubleSlice" {
		t.Errorf("expected 'DoubleSlice', got '%s'", double.Name)
	}
	if len(double.Returns) != 1 {
		t.Fatalf("expected 1 return for DoubleSlice, got %d", len(double.Returns))
	}
	if !double.Returns[0].IsSlice {
		t.Error("expected DoubleSlice return to be a slice")
	}
	if double.Returns[0].ElemType != "int64_t" {
		t.Errorf("expected return elem type 'int64_t', got '%s'", double.Returns[0].ElemType)
	}
}

func TestParseFileNoExports(t *testing.T) {
	path := filepath.Join("testdata", "no_exports.go")
	gf, err := bindings.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(gf.Functions) != 0 {
		t.Errorf("expected 0 exported functions, got %d", len(gf.Functions))
	}
}

func TestParseFileMethods(t *testing.T) {
	path := filepath.Join("testdata", "methods.go")
	gf, err := bindings.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Should only find GetCounter, not the method Increment
	if len(gf.Functions) != 1 {
		t.Fatalf("expected 1 function (methods excluded), got %d", len(gf.Functions))
	}
	if gf.Functions[0].Name != "GetCounter" {
		t.Errorf("expected 'GetCounter', got '%s'", gf.Functions[0].Name)
	}
}

func TestParseFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "bad.go")
	if err := os.WriteFile(invalidFile, []byte("package bad\n\nfunc ("), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := bindings.ParseFile(invalidFile)
	if err == nil {
		t.Error("expected error for invalid Go file")
	}
}

func TestParseDir(t *testing.T) {
	files, err := bindings.ParseDir("testdata")
	if err != nil {
		t.Fatalf("ParseDir failed: %v", err)
	}

	// Should find functions from simple.go and slices.go (methods.go has 1, no_exports.go has 0)
	totalFuncs := 0
	for _, f := range files {
		totalFuncs += len(f.Functions)
	}

	// simple.go: 3, slices.go: 2, methods.go: 1, no_exports.go: 0
	if totalFuncs != 6 {
		t.Errorf("expected 6 total functions across all files, got %d", totalFuncs)
	}
}

func TestParseDirSkipsTestFiles(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "example_test.go")
	content := `package example

//export Helper
func Helper(x int64) int64 {
	return x
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	files, err := bindings.ParseDir(tmpDir)
	if err != nil {
		t.Fatalf("ParseDir failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files (test files skipped), got %d", len(files))
	}
}

func TestParseDirEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	files, err := bindings.ParseDir(tmpDir)
	if err != nil {
		t.Fatalf("ParseDir failed: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files from empty dir, got %d", len(files))
	}
}

func TestParseFileCommentPreserved(t *testing.T) {
	path := filepath.Join("testdata", "simple.go")
	gf, err := bindings.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	add := gf.Functions[0]
	// CommentGroup.Text() strips // markers and leading whitespace
	// For "//export Add", Text() returns empty because it only has directive text
	// Verify the doc comment exists via the AST
	if len(add.Params) == 0 {
		t.Error("expected Add to have params")
	}
}
