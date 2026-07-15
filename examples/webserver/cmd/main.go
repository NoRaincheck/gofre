package main

/*
#include <stdlib.h>
#include <string.h>

typedef void (*dispatch_fn)(int, char*, char*);

static void call_dispatch(dispatch_fn fn, int id, char* body, char* out) {
    if (fn == NULL) { strcpy(out, "{\"error\":\"no dispatch registered\"}"); return; }
    fn(id, body, out);
}
*/
import "C"
import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"unsafe"

	"github.com/grackin/examples/webserver/pkg/core"
)

func main() {}

// ── JSON ─────────────────────────────────────────────────────────

//export Dumps
func Dumps(obj *C.char) *C.char {
	result := core.Dumps(C.GoString(obj))
	return C.CString(result)
}

//export Loads
func Loads(s *C.char) *C.char {
	result := core.Loads(C.GoString(s))
	return C.CString(result)
}

// ── HTTP Server ──────────────────────────────────────────────────

type serverEntry struct {
	mux    *http.ServeMux
	server *http.Server
}

var (
	dispatchFunc C.dispatch_fn
	servers      = make(map[int]*serverEntry)
	serversMu    sync.Mutex
	serverSeq    int
)

//export HTTPSetDispatch
func HTTPSetDispatch(fn C.dispatch_fn) {
	dispatchFunc = fn
}

//export HTTPCreateServer
func HTTPCreateServer() C.int {
	serversMu.Lock()
	defer serversMu.Unlock()
	serverSeq++
	servers[serverSeq] = &serverEntry{mux: http.NewServeMux()}
	return C.int(serverSeq)
}

//export HTTPAddRoute
func HTTPAddRoute(serverID C.int, method *C.char, path *C.char, handlerID C.int) {
	m := C.GoString(method)
	p := C.GoString(path)

	serversMu.Lock()
	s := servers[int(serverID)]
	serversMu.Unlock()

	if s == nil {
		return
	}

	s.mux.HandleFunc(m+" "+p, func(w http.ResponseWriter, r *http.Request) {
		body := ""
		if r.Body != nil {
			buf, _ := io.ReadAll(r.Body)
			body = string(buf)
		}
		if r.URL.RawQuery != "" {
			if body != "" {
				body += "\n"
			}
			body += r.URL.RawQuery
		}

		cBody := C.CString(body)
		defer C.free(unsafe.Pointer(cBody))

		outBuf := (*C.char)(C.malloc(65536))
		defer C.free(unsafe.Pointer(outBuf))
		C.memset(unsafe.Pointer(outBuf), 0, 65536)

		C.call_dispatch(dispatchFunc, C.int(handlerID), cBody, outBuf)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(C.GoString(outBuf)))
	})
}

//export HTTPStartServer
func HTTPStartServer(serverID C.int, addr *C.char) {
	a := C.GoString(addr)

	serversMu.Lock()
	s := servers[int(serverID)]
	serversMu.Unlock()

	if s == nil {
		return
	}

	s.server = &http.Server{
		Addr:    a,
		Handler: s.mux,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
}

func init() {
	_ = unsafe.Pointer(nil)
}
