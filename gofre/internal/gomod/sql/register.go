//go:build !cffi && !no_pocketpy

package sqlbridge

import (
	"encoding/json"
	"fmt"

	"github.com/NoRaincheck/gofre/internal/pocketpy"
)

func Register(vm *pocketpy.Interpreter) error {
	// gosql.open(path) -> handle
	err := vm.RegisterFunc("gosql", "open", "open(path)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		db, err := OpenDB(args[0].Str)
		if err != nil {
			return pocketpy.Value{}, err
		}
		id := NewDBHandle(db)
		return pocketpy.Value{Type: pocketpy.TypeInt, Int: id}, nil
	})
	if err != nil {
		return err
	}

	// gosql.close(handle)
	err = vm.RegisterFunc("gosql", "close", "close(handle)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		CloseDB(args[0].Int)
		return pocketpy.Value{Type: pocketpy.TypeNone}, nil
	})
	if err != nil {
		return err
	}

	// gosql.seed(handle) - creates table and seeds 10k rows
	err = vm.RegisterFunc("gosql", "seed", "seed(handle)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		db := GetDB(args[0].Int)
		if db == nil {
			return pocketpy.Value{}, fmt.Errorf("invalid db handle %d", args[0].Int)
		}
		if err := SeedIfEmpty(db); err != nil {
			return pocketpy.Value{}, err
		}
		return pocketpy.Value{Type: pocketpy.TypeNone}, nil
	})
	if err != nil {
		return err
	}

	// gosql.query_row(handle, sql, *args) -> JSON string
	err = vm.RegisterFunc("gosql", "query_row", "query_row(handle, sql, *args)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		handle := args[0].Int
		query := args[1].Str
		queryArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			queryArgs[i] = a.Int
		}
		result, err := QuerySingleRow(handle, query, queryArgs...)
		if err != nil {
			return pocketpy.Value{}, err
		}
		b, err := json.Marshal(result)
		if err != nil {
			return pocketpy.Value{}, err
		}
		return pocketpy.Value{Type: pocketpy.TypeStr, Str: string(b)}, nil
	})
	if err != nil {
		return err
	}

	// gosql.query_rows(handle, sql, *args) -> JSON string
	err = vm.RegisterFunc("gosql", "query_rows", "query_rows(handle, sql, *args)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		handle := args[0].Int
		query := args[1].Str
		queryArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			queryArgs[i] = a.Int
		}
		results, err := QueryRows(handle, query, queryArgs...)
		if err != nil {
			return pocketpy.Value{}, err
		}
		b, err := json.Marshal(results)
		if err != nil {
			return pocketpy.Value{}, err
		}
		return pocketpy.Value{Type: pocketpy.TypeStr, Str: string(b)}, nil
	})
	if err != nil {
		return err
	}

	// gosql.exec(handle, sql, *args)
	err = vm.RegisterFunc("gosql", "exec", "exec(handle, sql, *args)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		handle := args[0].Int
		query := args[1].Str
		queryArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			queryArgs[i] = a.Int
		}
		if err := Exec(handle, query, queryArgs...); err != nil {
			return pocketpy.Value{}, err
		}
		return pocketpy.Value{Type: pocketpy.TypeNone}, nil
	})
	if err != nil {
		return err
	}

	return nil
}
