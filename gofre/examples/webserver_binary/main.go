//go:build !no_pocketpy

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	httpbridge "github.com/NoRaincheck/gofre/internal/gomod/http"
	jsonbridge "github.com/NoRaincheck/gofre/internal/gomod/json"
	"github.com/NoRaincheck/gofre/internal/pocketpy"
)

//go:embed app.py
var appSource string

func main() {
	runtime.LockOSThread()
	vm := pocketpy.New()

	if err := httpbridge.Register(vm); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register gohttp: %v\n", err)
		os.Exit(1)
	}

	if err := jsonbridge.Register(vm); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register gojson: %v\n", err)
		os.Exit(1)
	}

	if err := vm.Exec(appSource, "app.py"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to exec app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Webserver binary started. Listening on :8080")
	fmt.Println("Press Ctrl+C to stop.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
}
