//go:build !no_pocketpy

// Package pocketpy provides a Go wrapper for the pocketpy C API.
//
// The amalgamated pocketpy.c and pocketpy.h files are generated from
// the pocketpy submodule (third_party/pocketpy). To regenerate:
//
//	cd third_party/pocketpy && python3 amalgamate.py
//	cp amalgamated/*.c amalgamated/*.h ../../goforge/internal/pocketpy/
package pocketpy

//go:generate sh -c "cd ../../../third_party/pocketpy && python3 amalgamate.py"
//go:generate cp ../../../third_party/pocketpy/amalgamated/pocketpy.c .
//go:generate cp ../../../third_party/pocketpy/amalgamated/pocketpy.h .

/*
#cgo CFLAGS: -DPK_ENABLE_OS=1
#cgo LDFLAGS: -lm
#include <stdlib.h>
#include "pocketpy.h"

// pk_bridge_ptr is defined in bridge.c. It holds a pointer to the C bridge
// function that pocketpy calls, which then dispatches to Go.
extern py_CFunction pk_bridge_ptr;
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

var (
	funcs  sync.Map
	initOnce sync.Once
)

type Interpreter struct {
	mu sync.Mutex
}

func New() *Interpreter {
	return newInterpreter()
}

// SwitchToVM switches the current goroutine to the given VM index.
// This is necessary when calling pocketpy functions from goroutines
// that were not the one that called New().
func SwitchToVM(index int) {
	C.py_switchvm(C.int(index))
}

func newInterpreter() *Interpreter {
	initOnce.Do(func() {
		C.py_initialize()
	})
	return &Interpreter{}
}

func (vm *Interpreter) Close() {
	if unsafe.Pointer(C.pk_bridge_ptr) == nil {
		return
	}
	C.py_switchvm(0)
	C.py_resetvm()
}

// Finalize shuts down all pocketpy VMs. This is irreversible.
// Call this once at the end of your application.
func Finalize() {
	C.py_switchvm(0)
	C.py_finalize()
}

func (vm *Interpreter) Exec(source, filename string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	return exec(source, filename, int(C.EXEC_MODE))
}

func (vm *Interpreter) Eval(expr string) (string, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	err := exec(expr, "<eval>", int(C.EVAL_MODE))
	if err != nil {
		return "", err
	}
	// Copy py_retval to a safe location (py_retval cannot be used as input)
	*C.py_getreg(0) = *C.py_retval()
	return pyObjectToString(C.py_getreg(0)), nil
}

func (vm *Interpreter) RegisterFunc(moduleName, funcName, sig string, fn GoFunc) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Verify bridge pointer is initialized
	if unsafe.Pointer(C.pk_bridge_ptr) == nil {
		return fmt.Errorf("bridge pointer is nil — C constructor didn't run")
	}

	cModule := C.CString(moduleName)
	mod := C.py_getmodule(cModule)
	if mod == nil {
		mod = C.py_newmodule(cModule)
	}
	C.free(unsafe.Pointer(cModule))

	funcs.Store(funcName, fn)

	cSig := C.CString(sig)
	C.py_bind(mod, cSig, C.pk_bridge_ptr)
	C.free(unsafe.Pointer(cSig))
	return nil
}

func (vm *Interpreter) CallFunc(funcName string, args ...Value) (Value, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	cName := C.CString(funcName)
	fn := C.py_getglobal(C.py_name(cName))
	C.free(unsafe.Pointer(cName))

	if fn == nil {
		return Value{}, fmt.Errorf("function '%s' not found", funcName)
	}

	argc := len(args)
	var argv []C.struct_py_TValue
	if argc > 0 {
		argv = make([]C.struct_py_TValue, argc)
		for i, a := range args {
			valueToPyRef(a, &argv[i])
		}
	}

	var argvPtr *C.struct_py_TValue
	if argc > 0 {
		argvPtr = &argv[0]
	}

	if !C.py_call(fn, C.int(argc), argvPtr) {
		return Value{}, extractError()
	}

	return pyRefToValue(C.py_retval()), nil
}

func (vm *Interpreter) SetGlobal(name string, val int64) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	cName := C.CString(name)
	C.py_newint(C.py_getreg(0), C.py_i64(val))
	C.py_setglobal(C.py_name(cName), C.py_getreg(0))
	C.free(unsafe.Pointer(cName))
}

func (vm *Interpreter) GetGlobal(name string) (Value, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	cName := C.CString(name)
	item := C.py_getglobal(C.py_name(cName))
	C.free(unsafe.Pointer(cName))
	if item == nil {
		return Value{}, fmt.Errorf("name '%s' is not defined", name)
	}
	return pyRefToValue(item), nil
}

//export goDispatch
func goDispatch(name *C.char, argc C.int, argv *C.struct_py_TValue) C.bool {
	funcName := C.GoString(name)
	fn, ok := funcs.Load(funcName)
	if !ok {
		return C.bool(false)
	}
	goFn := fn.(GoFunc)

	args := make([]Value, argc)
	elemSize := unsafe.Sizeof(C.struct_py_TValue{})
	base := unsafe.Pointer(argv)
	for i := 0; i < int(argc); i++ {
		argPtr := (*C.struct_py_TValue)(unsafe.Pointer(uintptr(base) + uintptr(i)*elemSize))
		args[i] = pyRefToValue(argPtr)
	}

	result, err := goFn(args)
	if err != nil {
		raiseError(err.Error())
		return C.bool(false)
	}

	valueToPyRef(result, C.py_retval())
	return C.bool(true)
}

func raiseError(msg string) {
	cMsg := C.CString(msg)
	C.py_newstr(C.py_getreg(0), cMsg)
	C.free(unsafe.Pointer(cMsg))
	C.py_tpcall(C.tp_RuntimeError, 1, C.py_getreg(0))
	C.py_raise(C.py_retval())
}

func exec(source, filename string, mode int) error {
	cSource := C.CString(source)
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cSource))
	defer C.free(unsafe.Pointer(cFilename))

	if C.py_exec(cSource, cFilename, C.enum_py_CompileMode(mode), nil) {
		return nil
	}
	return extractError()
}

func extractError() error {
	exc := C.py_formatexc()
	if exc == nil {
		return fmt.Errorf("python error (unknown)")
	}
	defer C.free(unsafe.Pointer(exc))
	return fmt.Errorf("python error: %s", C.GoString(exc))
}

func pyObjectToString(ref *C.struct_py_TValue) string {
	if C.py_istype(ref, C.tp_str) {
		return C.GoString(C.py_tostr(ref))
	}
	// Copy ref to a temporary register before calling py_str on it
	// because py_str writes to py_retval and py_retval may overlap with ref
	*C.py_getreg(1) = *ref
	if !C.py_str(C.py_getreg(1)) {
		return "<error>"
	}
	return C.GoString(C.py_tostr(C.py_retval()))
}

func valueToPyRef(v Value, out *C.struct_py_TValue) {
	switch v.Type {
	case TypeNone:
		C.py_newnone(out)
	case TypeInt:
		C.py_newint(out, C.py_i64(v.Int))
	case TypeFloat:
		C.py_newfloat(out, C.py_f64(v.Float))
	case TypeBool:
		C.py_newbool(out, C.bool(v.Bool))
	case TypeStr:
		s := C.CString(v.Str)
		C.py_newstr(out, s)
		C.free(unsafe.Pointer(s))
	case TypeList:
		C.py_newlist(out)
		for _, item := range v.Items {
			valueToPyRef(item, C.py_getreg(0))
			C.py_list_append(out, C.py_getreg(0))
		}
	}
}

func pyRefToValue(ref *C.struct_py_TValue) Value {
	v := Value{}
	if ref == nil {
		return v
	}
	if C.py_istype(ref, C.tp_int) {
		v.Type = TypeInt
		v.Int = int64(C.py_toint(ref))
	} else if C.py_istype(ref, C.tp_float) {
		v.Type = TypeFloat
		v.Float = float64(C.py_tofloat(ref))
	} else if C.py_istype(ref, C.tp_bool) {
		v.Type = TypeBool
		v.Bool = bool(C.py_tobool(ref))
	} else if C.py_istype(ref, C.tp_str) {
		v.Type = TypeStr
		v.Str = C.GoString(C.py_tostr(ref))
	} else if C.py_istype(ref, C.tp_NoneType) {
		v.Type = TypeNone
	}
	return v
}
