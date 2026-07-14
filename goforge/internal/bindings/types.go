package bindings

import "strings"

type TypeMapping struct {
	CType       string
	CFFIType    string
	GoType      string
	IsSlice     bool
	ElemCType   string
	ElemCFFI    string
	GoElemType  string
}

var typeMap = map[string]TypeMapping{
	"int8":     {CType: "int8_t", CFFIType: "int8_t", GoType: "int8"},
	"int16":    {CType: "int16_t", CFFIType: "int16_t", GoType: "int16"},
	"int32":    {CType: "int32_t", CFFIType: "int32_t", GoType: "int32"},
	"int64":    {CType: "int64_t", CFFIType: "int64_t", GoType: "int64"},
	"uint8":    {CType: "uint8_t", CFFIType: "uint8_t", GoType: "uint8"},
	"uint16":   {CType: "uint16_t", CFFIType: "uint16_t", GoType: "uint16"},
	"uint32":   {CType: "uint32_t", CFFIType: "uint32_t", GoType: "uint32"},
	"uint64":   {CType: "uint64_t", CFFIType: "uint64_t", GoType: "uint64"},
	"float32":  {CType: "float", CFFIType: "float", GoType: "float32"},
	"float64":  {CType: "double", CFFIType: "double", GoType: "float64"},
	"bool":     {CType: "int", CFFIType: "int", GoType: "bool"},
	"string":   {CType: "char*", CFFIType: "char*", GoType: "string"},
	"byte":     {CType: "uint8_t", CFFIType: "uint8_t", GoType: "byte"},
}

func MapType(goType string) TypeMapping {
	if m, ok := typeMap[goType]; ok {
		return m
	}

	if strings.HasPrefix(goType, "[]") {
		elemType := goType[2:]
		if m, ok := typeMap[elemType]; ok {
			return TypeMapping{
				CType:      m.CType + "*",
				CFFIType:   m.CFFIType + "*",
				GoType:     goType,
				IsSlice:    true,
				ElemCType:  m.CType,
				ElemCFFI:   m.CFFIType,
				GoElemType: elemType,
			}
		}
	}

	return TypeMapping{
		CType:    "void*",
		CFFIType: "void*",
		GoType:   goType,
	}
}
