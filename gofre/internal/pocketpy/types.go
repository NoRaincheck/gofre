//go:build !no_pocketpy

package pocketpy

import "sync"

// ValueType represents the type of a Python value.
type ValueType int

const (
	TypeNone   ValueType = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeStr
	TypeList
	TypeDict
	TypeTuple
	TypeBytes
	TypeUnknown
)

// Value holds a Python value returned from the interpreter.
type Value struct {
	Type  ValueType
	Int   int64
	Float float64
	Bool  bool
	Str   string
	Items []Value // for lists/tuples
}

// GoFunc is a Go function that can be called from Python.
type GoFunc func(args []Value) (Value, error)

var goFuncs sync.Map
