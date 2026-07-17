// Pure Go stdlib benchmark server with TFB-style endpoints.
//
// Routes:
//   GET  /plaintext   — Plain text response
//   GET  /json        — JSON serialization
//   GET  /db          — Single random row query
//   GET  /queries     — Multiple random row queries
//   POST /updates     — Update random rows
//
// Usage: go run server_pure_go.go [port]

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"

	_ "modernc.org/sqlite"
)

var (
	db          *sql.DB
	dbQueryRow  *sql.Stmt
	dbQueryRows *sql.Stmt
)

func init() {
	var err error
	db, err = sql.Open("sqlite", "benchmark_go.db?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_txlock=immediate")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(runtime.NumCPU())
	db.SetMaxIdleConns(runtime.NumCPU())
	seedDB()
	dbQueryRow, err = db.Prepare("SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}
	dbQueryRows, err = db.Prepare("SELECT id, randomNumber FROM world ORDER BY RANDOM() LIMIT ?")
	if err != nil {
		log.Fatal(err)
	}
}

func seedDB() {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM world").Scan(&count); err == nil && count > 0 {
		return
	}
	db.Exec("CREATE TABLE IF NOT EXISTS world (id INTEGER PRIMARY KEY, randomNumber INTEGER NOT NULL DEFAULT 0)")
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO world (id, randomNumber) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 1; i <= 10000; i++ {
		stmt.Exec(i, rand.Intn(10000)+1)
	}
	tx.Commit()
}

func main() {
	port := "8085"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/plaintext", handlePlaintext)
	http.HandleFunc("/json", handleJSON)
	http.HandleFunc("/db", handleDB)
	http.HandleFunc("/queries", handleQueries)
	http.HandleFunc("/updates", handleUpdates)

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

// ── DB Single Query ───────────────────────────────────────────────
func handleDB(w http.ResponseWriter, r *http.Request) {
	var id, randomNumber int64
	err := dbQueryRow.QueryRow().Scan(&id, &randomNumber)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           id,
		"randomNumber": randomNumber,
	})
}

// ── DB Multiple Queries ──────────────────────────────────────────
func handleQueries(w http.ResponseWriter, r *http.Request) {
	n := 1
	if q := r.URL.Query().Get("N"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 && v <= 500 {
			n = v
		}
	}
	rows, err := dbQueryRows.Query(n)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var id, randomNumber int64
		if err := rows.Scan(&id, &randomNumber); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
			return
		}
		results = append(results, map[string]interface{}{
			"id":           id,
			"randomNumber": randomNumber,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// ── DB Updates ───────────────────────────────────────────────────
func handleUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var updates []struct {
		ID           int64 `json:"id"`
		RandomNumber int64 `json:"randomNumber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 400)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}
	stmt, err := tx.Prepare("UPDATE world SET randomNumber = ? WHERE id = ?")
	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}
	defer stmt.Close()
	for _, u := range updates {
		stmt.Exec(u.RandomNumber, u.ID)
	}
	tx.Commit()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updates)
}
