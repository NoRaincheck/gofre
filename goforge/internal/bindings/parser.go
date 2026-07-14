package bindings

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Function struct {
	Name    string
	Params  []Param
	Returns []Return
	Comment string
}

type Param struct {
	Name      string
	Type      string
	GoType    string
	IsSlice   bool
	ElemType  string
	CFFIElem  string
}

type Return struct {
	Type      string
	GoType    string
	IsSlice   bool
	ElemType  string
	CFFIElem  string
}

type GoFile struct {
	Path      string
	Package   string
	Functions []Function
}

func ParseDir(dir string) ([]*GoFile, error) {
	var files []*GoFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		gf, err := ParseFile(path)
		if err != nil {
			return err
		}
		if gf != nil && len(gf.Functions) > 0 {
			files = append(files, gf)
		}
		return nil
	})

	return files, err
}

func ParseFile(path string) (*GoFile, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	gf := &GoFile{
		Path:    path,
		Package: node.Name.Name,
	}

	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}

		if !fn.Name.IsExported() {
			continue
		}

		if fn.Doc != nil {
			for _, comment := range fn.Doc.List {
				if strings.Contains(comment.Text, "//export") {
					f := parseFunction(fn)
					gf.Functions = append(gf.Functions, f)
					break
				}
			}
		}
	}

	return gf, nil
}

func parseFunction(fn *ast.FuncDecl) Function {
	f := Function{
		Name: fn.Name.Name,
	}

	if fn.Doc != nil {
		f.Comment = fn.Doc.Text()
	}

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			goType := extractType(field.Type)
			mapping := MapType(goType)
			for _, name := range field.Names {
				p := Param{
					Name:     name.Name,
					Type:     mapping.CType,
					GoType:   mapping.GoType,
					IsSlice:  mapping.IsSlice,
					ElemType: mapping.ElemCType,
					CFFIElem: mapping.ElemCFFI,
				}
				f.Params = append(f.Params, p)
			}
		}
	}

	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			goType := extractType(field.Type)
			mapping := MapType(goType)
			r := Return{
				Type:     mapping.CType,
				GoType:   mapping.GoType,
				IsSlice:  mapping.IsSlice,
				ElemType: mapping.ElemCType,
				CFFIElem: mapping.ElemCFFI,
			}
			f.Returns = append(f.Returns, r)
		}
	}

	return f
}

func extractType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return extractType(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + extractType(t.X)
	case *ast.ArrayType:
		return "[]" + extractType(t.Elt)
	case *ast.SliceExpr:
		return "[]" + extractType(t.X)
	default:
		return "interface{}"
	}
}
