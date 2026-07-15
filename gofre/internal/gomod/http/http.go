package httpbridge

import (
	"net/http"
	"sync"
)

type route struct {
	method string
	path   string
	h      http.HandlerFunc
}

type Server struct {
	mu     sync.Mutex
	mux    *http.ServeMux
	routes []route
	server *http.Server
}

func NewServer() *Server {
	return &Server{
		mux: http.NewServeMux(),
	}
}

func (s *Server) Handle(method, pattern string, handler http.HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes = append(s.routes, route{method, pattern, handler})
	fullPattern := pattern
	if method != "" {
		fullPattern = method + " " + pattern
	}
	s.mux.HandleFunc(fullPattern, handler)
}

func (s *Server) ListenAndServe(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}
	return s.server.ListenAndServe()
}

var servers = make(map[int]*Server)
var serversMu sync.Mutex
var serverSeq int

func NewServerHandle() int {
	serversMu.Lock()
	defer serversMu.Unlock()
	serverSeq++
	s := NewServer()
	servers[serverSeq] = s
	return serverSeq
}

func GetServer(handle int) *Server {
	serversMu.Lock()
	defer serversMu.Unlock()
	return servers[handle]
}
