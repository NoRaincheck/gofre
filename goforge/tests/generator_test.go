package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grackin/goforge/internal/bindings"
)

func TestGenerateBasic(t *testing.T) {
	tmpDir := t.TempDir()

	files := []*bindings.GoFile{
		{
			Path:    "test.go",
			Package: "testpkg",
			Functions: []bindings.Function{
				{
					Name: "Add",
					Params: []bindings.Param{
						{Name: "a", Type: "int64_t", GoType: "int64"},
						{Name: "b", Type: "int64_t", GoType: "int64"},
					},
					Returns: []bindings.Return{
						{Type: "int64_t", GoType: "int64"},
					},
				},
			},
		},
	}

	gen := bindings.NewGenerator(tmpDir, "testpkg")
	if err := gen.Generate(files); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check _binding.py was created
	bindingPath := filepath.Join(tmpDir, "_binding.py")
	if _, err := os.Stat(bindingPath); os.IsNotExist(err) {
		t.Fatal("_binding.py was not created")
	}

	bindingContent, err := os.ReadFile(bindingPath)
	if err != nil {
		t.Fatalf("failed to read _binding.py: %v", err)
	}

	bindingStr := string(bindingContent)
	if !strings.Contains(bindingStr, "def Add(a, b):") {
		t.Error("_binding.py missing Add function definition")
	}
	if !strings.Contains(bindingStr, "ffi.cdef") {
		t.Error("_binding.py missing ffi.cdef")
	}
	if !strings.Contains(bindingStr, "int64_t Add(int64_t a, int64_t b)") {
		t.Error("_binding.py missing correct cffi declaration")
	}

	// Check __init__.py was created
	initPath := filepath.Join(tmpDir, "__init__.py")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		t.Fatal("__init__.py was not created")
	}

	initContent, err := os.ReadFile(initPath)
	if err != nil {
		t.Fatalf("failed to read __init__.py: %v", err)
	}

	initStr := string(initContent)
	if !strings.Contains(initStr, "from ._binding import Add") {
		t.Error("__init__.py missing Add import")
	}
	if !strings.Contains(initStr, `__all__ = ["Add"]`) {
		t.Error("__init__.py missing __all__")
	}
}

func TestGenerateMultipleFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	files := []*bindings.GoFile{
		{
			Path:    "math.go",
			Package: "math",
			Functions: []bindings.Function{
				{
					Name: "Add",
					Params: []bindings.Param{
						{Name: "a", Type: "int64_t"},
						{Name: "b", Type: "int64_t"},
					},
					Returns: []bindings.Return{{Type: "int64_t"}},
				},
				{
					Name: "Multiply",
					Params: []bindings.Param{
						{Name: "a", Type: "int64_t"},
						{Name: "b", Type: "int64_t"},
					},
					Returns: []bindings.Return{{Type: "int64_t"}},
				},
				{
					Name:    "GetHello",
					Params:  []bindings.Param{},
					Returns: []bindings.Return{{Type: "char*"}},
				},
			},
		},
	}

	gen := bindings.NewGenerator(tmpDir, "math")
	if err := gen.Generate(files); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	bindingContent, err := os.ReadFile(filepath.Join(tmpDir, "_binding.py"))
	if err != nil {
		t.Fatalf("failed to read _binding.py: %v", err)
	}

	bindingStr := string(bindingContent)
	for _, name := range []string{"Add", "Multiply", "GetHello"} {
		if !strings.Contains(bindingStr, "def "+name+"(") {
			t.Errorf("_binding.py missing %s function", name)
		}
	}

	initContent, err := os.ReadFile(filepath.Join(tmpDir, "__init__.py"))
	if err != nil {
		t.Fatalf("failed to read __init__.py: %v", err)
	}

	initStr := string(initContent)
	for _, name := range []string{"Add", "Multiply", "GetHello"} {
		if !strings.Contains(initStr, name) {
			t.Errorf("__init__.py missing %s", name)
		}
	}
}

func TestGenerateNoFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	gen := bindings.NewGenerator(tmpDir, "empty")
	err := gen.Generate([]*bindings.GoFile{})
	if err == nil {
		t.Error("expected error when no functions provided")
	}
	if !strings.Contains(err.Error(), "no exported functions") {
		t.Errorf("expected 'no exported functions' error, got: %v", err)
	}
}

func TestGenerateSliceParams(t *testing.T) {
	tmpDir := t.TempDir()

	files := []*bindings.GoFile{
		{
			Path:    "slices.go",
			Package: "math",
			Functions: []bindings.Function{
				{
					Name: "SumSlice",
					Params: []bindings.Param{
						{
							Name:     "data",
							Type:     "double*",
							IsSlice:  true,
							ElemType: "double",
						},
					},
					Returns: []bindings.Return{
						{Type: "double"},
					},
				},
			},
		},
	}

	gen := bindings.NewGenerator(tmpDir, "math")
	if err := gen.Generate(files); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	bindingContent, err := os.ReadFile(filepath.Join(tmpDir, "_binding.py"))
	if err != nil {
		t.Fatalf("failed to read _binding.py: %v", err)
	}

	bindingStr := string(bindingContent)
	// Should have cffi declaration with pointer param
	if !strings.Contains(bindingStr, "double* data") {
		t.Error("_binding.py missing slice pointer declaration")
	}
	// Should have length parameter
	if !strings.Contains(bindingStr, "int64_t data_len") {
		t.Error("_binding.py missing slice length parameter")
	}
	// Should have ffi.new for converting list to C array
	if !strings.Contains(bindingStr, `ffi.new("double[]"`) {
		t.Error("_binding.py missing ffi.new for slice conversion")
	}
}

func TestCffiDeclParams(t *testing.T) {
	params := []bindings.Param{
		{Name: "x", Type: "int64_t"},
		{Name: "y", Type: "double"},
	}
	result := bindings.CffiDeclParams(params)
	if result != "int64_t x, double y" {
		t.Errorf("expected 'int64_t x, double y', got '%s'", result)
	}
}

func TestCffiDeclParamsSlice(t *testing.T) {
	params := []bindings.Param{
		{Name: "data", Type: "int64_t*", IsSlice: true, ElemType: "int64_t"},
	}
	result := bindings.CffiDeclParams(params)
	if result != "int64_t* data, int64_t data_len" {
		t.Errorf("expected 'int64_t* data, int64_t data_len', got '%s'", result)
	}
}

func TestCffiCallArgs(t *testing.T) {
	params := []bindings.Param{
		{Name: "a", Type: "int64_t"},
		{Name: "b", Type: "int64_t"},
	}
	result := bindings.CffiCallArgs(params)
	if result != "a, b" {
		t.Errorf("expected 'a, b', got '%s'", result)
	}
}

func TestCffiNewArg(t *testing.T) {
	params := []bindings.Param{
		{Name: "x", Type: "int64_t"},
		{Name: "data", Type: "int64_t*", IsSlice: true},
	}
	result := bindings.CffiNewArg(params)
	expected := "x, _c_data, len(data)"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestHasSliceParam(t *testing.T) {
	paramsNoSlice := []bindings.Param{
		{Name: "a", Type: "int64_t"},
		{Name: "b", Type: "int64_t"},
	}
	if bindings.HasSliceParam(paramsNoSlice) {
		t.Error("expected HasSliceParam false for non-slice params")
	}

	paramsWithSlice := []bindings.Param{
		{Name: "data", Type: "int64_t*", IsSlice: true},
	}
	if !bindings.HasSliceParam(paramsWithSlice) {
		t.Error("expected HasSliceParam true for slice params")
	}
}
