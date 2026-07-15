// Pure Go stdlib benchmark server with TFB-style endpoints.
//
// Routes:
//   GET  /plaintext   — Plain text response
//   GET  /json        — JSON serialization
//
// Usage: go run server_pure_go.go [port]

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	port := "8085"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/plaintext", handlePlaintext)
	http.HandleFunc("/json", handleJSON)

	log.Printf("Pure Go benchmark server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// ── Plaintext ─────────────────────────────────────────────────────
func handlePlaintext(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

// ── JSON ──────────────────────────────────────────────────────────
func handleJSON(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"message":   "Hello, World!",
		"timestamp": float64(1234567890),
		"random":    float64(42),
		"data": map[string]interface{}{
			"name":     "benchmark",
			"version":  "1.0.0",
			"features": []string{"json", "db", "template"},
			"metadata": map[string]interface{}{
				"host": "localhost",
				"port": float64(8080),
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
