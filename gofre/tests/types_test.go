package tests

import (
	"testing"

	"github.com/NoRaincheck/gofre/internal/bindings"
)

func TestMapTypePrimitives(t *testing.T) {
	tests := []struct {
		goType   string
		wantC    string
		wantFFI  string
		wantGo   string
		isSlice  bool
	}{
		{"int8", "int8_t", "int8_t", "int8", false},
		{"int16", "int16_t", "int16_t", "int16", false},
		{"int32", "int32_t", "int32_t", "int32", false},
		{"int64", "int64_t", "int64_t", "int64", false},
		{"uint8", "uint8_t", "uint8_t", "uint8", false},
		{"uint16", "uint16_t", "uint16_t", "uint16", false},
		{"uint32", "uint32_t", "uint32_t", "uint32", false},
		{"uint64", "uint64_t", "uint64_t", "uint64", false},
		{"float32", "float", "float", "float32", false},
		{"float64", "double", "double", "float64", false},
		{"bool", "int", "int", "bool", false},
		{"string", "char*", "char*", "string", false},
		{"byte", "uint8_t", "uint8_t", "byte", false},
	}

	for _, tt := range tests {
		t.Run(tt.goType, func(t *testing.T) {
			m := bindings.MapType(tt.goType)
			if m.CType != tt.wantC {
				t.Errorf("CType: got %q, want %q", m.CType, tt.wantC)
			}
			if m.CFFIType != tt.wantFFI {
				t.Errorf("CFFIType: got %q, want %q", m.CFFIType, tt.wantFFI)
			}
			if m.GoType != tt.wantGo {
				t.Errorf("GoType: got %q, want %q", m.GoType, tt.wantGo)
			}
			if m.IsSlice != tt.isSlice {
				t.Errorf("IsSlice: got %v, want %v", m.IsSlice, tt.isSlice)
			}
		})
	}
}

func TestMapTypeSlices(t *testing.T) {
	tests := []struct {
		goType    string
		wantC     string
		wantElem  string
		wantGo    string
		goElem    string
	}{
		{"[]int64", "int64_t*", "int64_t", "[]int64", "int64"},
		{"[]float64", "double*", "double", "[]float64", "float64"},
		{"[]string", "char**", "char*", "[]string", "string"},
		{"[]bool", "int*", "int", "[]bool", "bool"},
		{"[]byte", "uint8_t*", "uint8_t", "[]byte", "byte"},
		{"[]float32", "float*", "float", "[]float32", "float32"},
	}

	for _, tt := range tests {
		t.Run(tt.goType, func(t *testing.T) {
			m := bindings.MapType(tt.goType)
			if m.CType != tt.wantC {
				t.Errorf("CType: got %q, want %q", m.CType, tt.wantC)
			}
			if m.ElemCType != tt.wantElem {
				t.Errorf("ElemCType: got %q, want %q", m.ElemCType, tt.wantElem)
			}
			if m.GoType != tt.wantGo {
				t.Errorf("GoType: got %q, want %q", m.GoType, tt.wantGo)
			}
			if !m.IsSlice {
				t.Error("expected IsSlice to be true")
			}
			if m.GoElemType != tt.goElem {
				t.Errorf("GoElemType: got %q, want %q", m.GoElemType, tt.goElem)
			}
		})
	}
}

func TestMapTypeUnknown(t *testing.T) {
	m := bindings.MapType("MyStruct")
	if m.CType != "void*" {
		t.Errorf("expected void* for unknown type, got %q", m.CType)
	}
	if m.CFFIType != "void*" {
		t.Errorf("expected void* CFFIType for unknown type, got %q", m.CFFIType)
	}
	if m.IsSlice {
		t.Error("expected IsSlice false for unknown type")
	}
}

func TestMapTypeUnknownSlice(t *testing.T) {
	m := bindings.MapType("[]MyStruct")
	// Unknown element type means it won't match in the slice handler
	// It falls through to the void* default
	if m.CType != "void*" {
		t.Errorf("expected void* for unknown slice element type, got %q", m.CType)
	}
}
