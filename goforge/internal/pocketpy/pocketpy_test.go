//go:build !no_pocketpy

package pocketpy

import (
	"fmt"
	"testing"
)

func TestExec(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.Exec("x = 1 + 2", "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	val, err := vm.GetGlobal("x")
	if err != nil {
		t.Fatalf("GetGlobal failed: %v", err)
	}
	if val.Type != TypeInt || val.Int != 3 {
		t.Fatalf("expected x=3, got %+v", val)
	}
}

func TestEval(t *testing.T) {
	vm := New()
	defer vm.Close()

	result, err := vm.Eval("1 + 2")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}
	if result != "3" {
		t.Fatalf("expected '3', got '%s'", result)
	}
}

func TestEvalString(t *testing.T) {
	vm := New()
	defer vm.Close()

	result, err := vm.Eval(`"hello" + " " + "world"`)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}
	if result != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", result)
	}
}

func TestEvalFloat(t *testing.T) {
	vm := New()
	defer vm.Close()

	result, err := vm.Eval("3.14 * 2")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}
	if result != "6.28" {
		t.Logf("float result: %s", result)
	}
}

func TestExecWithError(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.Exec("1 / 0", "<test>")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	t.Logf("expected error: %v", err)
}

func TestRegisterFunc(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.RegisterFunc("testmod", "add", "add(x, y)", func(args []Value) (Value, error) {
		return Value{Type: TypeInt, Int: args[0].Int + args[1].Int}, nil
	})
	if err != nil {
		t.Fatalf("RegisterFunc failed: %v", err)
	}

	err = vm.Exec(`
import testmod
result = testmod.add(3, 4)
`, "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	val, err := vm.GetGlobal("result")
	if err != nil {
		t.Fatalf("GetGlobal failed: %v", err)
	}
	if val.Type != TypeInt || val.Int != 7 {
		t.Fatalf("expected result=7, got %+v", val)
	}
}

func TestRegisterFuncWithString(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.RegisterFunc("testmod", "greet", "greet(name)", func(args []Value) (Value, error) {
		return Value{Type: TypeStr, Str: "Hello, " + args[0].Str + "!"}, nil
	})
	if err != nil {
		t.Fatalf("RegisterFunc failed: %v", err)
	}

	err = vm.Exec(`
import testmod
result = testmod.greet("World")
`, "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	val, err := vm.GetGlobal("result")
	if err != nil {
		t.Fatalf("GetGlobal failed: %v", err)
	}
	if val.Type != TypeStr || val.Str != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got %+v", val)
	}
}

func TestFibonacci(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.RegisterFunc("math", "fib", "fib(n)", func(args []Value) (Value, error) {
		n := args[0].Int
		if n <= 1 {
			return Value{Type: TypeInt, Int: n}, nil
		}
		a, b := int64(0), int64(1)
		for i := int64(2); i <= n; i++ {
			a, b = b, a+b
		}
		return Value{Type: TypeInt, Int: b}, nil
	})
	if err != nil {
		t.Fatalf("RegisterFunc failed: %v", err)
	}

	err = vm.Exec(`
import math
result = math.fib(20)
`, "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	val, err := vm.GetGlobal("result")
	if err != nil {
		t.Fatalf("GetGlobal failed: %v", err)
	}
	if val.Type != TypeInt || val.Int != 6765 {
		t.Fatalf("expected fib(20)=6765, got %+v", val)
	}
}

func TestPythonFibonacci(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.Exec(`
def fib(n):
    a, b = 0, 1
    for _ in range(2, n+1):
        a, b = b, a + b
    return b

result = fib(20)
`, "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	val, err := vm.GetGlobal("result")
	if err != nil {
		t.Fatalf("GetGlobal failed: %v", err)
	}
	if val.Type != TypeInt || val.Int != 6765 {
		t.Fatalf("expected fib(20)=6765, got %+v", val)
	}
}

func TestSetGlobal(t *testing.T) {
	vm := New()
	defer vm.Close()

	vm.SetGlobal("x", 42)

	err := vm.Exec("y = x * 2", "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	val, err := vm.GetGlobal("y")
	if err != nil {
		t.Fatalf("GetGlobal failed: %v", err)
	}
	if val.Type != TypeInt || val.Int != 84 {
		t.Fatalf("expected y=84, got %+v", val)
	}
}

func TestPythonPrint(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.Exec("print('hello from pocketpy!')", "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
}

func TestPythonListComp(t *testing.T) {
	vm := New()
	defer vm.Close()

	result, err := vm.Eval("[x*2 for x in range(5)]")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}
	t.Logf("list comprehension result: %s", result)
}

func TestCallFunc(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.Exec(`
def add(a, b):
    return a + b

def greet(name):
    return "Hello, " + name + "!"
`, "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	result, err := vm.CallFunc("add", Value{Type: TypeInt, Int: 3}, Value{Type: TypeInt, Int: 4})
	if err != nil {
		t.Fatalf("CallFunc failed: %v", err)
	}
	if result.Type != TypeInt || result.Int != 7 {
		t.Fatalf("expected 7, got %+v", result)
	}

	result, err = vm.CallFunc("greet", Value{Type: TypeStr, Str: "World"})
	if err != nil {
		t.Fatalf("CallFunc greet failed: %v", err)
	}
	if result.Type != TypeStr || result.Str != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got %+v", result)
	}

	_, err = vm.CallFunc("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent function")
	}
}

func TestRegisterFuncError(t *testing.T) {
	vm := New()
	defer vm.Close()

	err := vm.RegisterFunc("testmod", "div", "div(x, y)", func(args []Value) (Value, error) {
		if args[1].Int == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}
		return Value{Type: TypeInt, Int: args[0].Int / args[1].Int}, nil
	})
	if err != nil {
		t.Fatalf("RegisterFunc failed: %v", err)
	}

	err = vm.Exec(`
import testmod
try:
    result = testmod.div(5, 0)
    success = False
except RuntimeError as e:
    success = True
    error_msg = str(e)
`, "<test>")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	val, err := vm.GetGlobal("success")
	if err != nil {
		t.Fatalf("GetGlobal failed: %v", err)
	}
	if val.Type != TypeBool || !val.Bool {
		t.Fatal("expected success=True (RuntimeError caught)")
	}

	msgVal, err := vm.GetGlobal("error_msg")
	if err != nil {
		t.Fatalf("GetGlobal error_msg failed: %v", err)
	}
	t.Logf("error message: %s", msgVal.Str)
}
