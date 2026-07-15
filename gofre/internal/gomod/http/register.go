//go:build !cffi && !no_pocketpy

package httpbridge

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"

	"github.com/NoRaincheck/gofre/internal/pocketpy"
)

type pyRequest struct {
	funcName string
	body     string
	resultCh chan pyResponse
}

type pyResponse struct {
	body string
	err  error
}

var (
	requestCh chan pyRequest
	startOnce sync.Once
)

func initDispatcher(vm *pocketpy.Interpreter) {
	startOnce.Do(func() {
		requestCh = make(chan pyRequest, 1000)
		go func() {
			runtime.LockOSThread()
			pocketpy.SwitchToVM(0)
			for req := range requestCh {
				result, err := vm.CallFunc(req.funcName, pocketpy.Value{Type: pocketpy.TypeStr, Str: req.body})
				if err != nil {
					req.resultCh <- pyResponse{err: err}
				} else {
					req.resultCh <- pyResponse{body: result.Str}
				}
			}
		}()
	})
}

func dispatchToPython(funcName, body string) (string, error) {
	resultCh := make(chan pyResponse, 1)
	requestCh <- pyRequest{funcName: funcName, body: body, resultCh: resultCh}
	resp := <-resultCh
	return resp.body, resp.err
}

func Register(vm *pocketpy.Interpreter) error {
	initDispatcher(vm)

	err := vm.RegisterFunc("gohttp", "create_server", "create_server()", func(args []pocketpy.Value) (pocketpy.Value, error) {
		id := NewServerHandle()
		return pocketpy.Value{Type: pocketpy.TypeInt, Int: int64(id)}, nil
	})
	if err != nil {
		return err
	}

	err = vm.RegisterFunc("gohttp", "add_route", "add_route(server_id, method, path, handler_id)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		serverID := int(args[0].Int)
		method := args[1].Str
		path := args[2].Str
		handlerID := int(args[3].Int)

		s := GetServer(serverID)
		if s == nil {
			return pocketpy.Value{}, fmt.Errorf("server %d not found", serverID)
		}

		s.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
			body := ""
			if r.Body != nil {
				buf, _ := io.ReadAll(r.Body)
				body = string(buf)
			}

			funcName := fmt.Sprintf("_handler_%d", handlerID)
			respBody, err := dispatchToPython(funcName, body)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), 500)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, respBody)
		})
		return pocketpy.Value{Type: pocketpy.TypeNone}, nil
	})
	if err != nil {
		return err
	}

	err = vm.RegisterFunc("gohttp", "start_server", "start_server(server_id, addr)", func(args []pocketpy.Value) (pocketpy.Value, error) {
		serverID := int(args[0].Int)
		addr := args[1].Str

		s := GetServer(serverID)
		if s == nil {
			return pocketpy.Value{}, fmt.Errorf("server %d not found", serverID)
		}

		go func() {
			runtime.LockOSThread()
			if err := s.ListenAndServe(addr); err != nil {
				panic(err)
			}
		}()
		return pocketpy.Value{Type: pocketpy.TypeNone}, nil
	})
	if err != nil {
		return err
	}

	return nil
}
