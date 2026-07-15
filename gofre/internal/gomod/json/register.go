//go:build !cffi && !no_pocketpy

package jsonbridge

import "github.com/grackin/gofre/internal/pocketpy"

func Register(vm *pocketpy.Interpreter) error {
	err := vm.RegisterFunc("gojson", "dumps", "dumps(obj)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		result := GoDumps(args[0].Str)
		return pocketpy.Value{Type: pocketpy.TypeStr, Str: result}, nil
	})
	if err != nil {
		return err
	}

	err = vm.RegisterFunc("gojson", "loads", "loads(s)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		result := GoLoads(args[0].Str)
		return pocketpy.Value{Type: pocketpy.TypeStr, Str: result}, nil
	})
	if err != nil {
		return err
	}

	return nil
}
