package sqlbridge

import (
	"database/sql"
	"math/rand"
	"time"

	_ "modernc.org/sqlite"
)

// OpenDB opens a SQLite database with WAL mode and busy timeout.
func OpenDB(path string) (*sql.DB, error) {
	dsn := path + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_txlock=immediate"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// Single connection: the pocketpy Python VM is single-threaded (one goroutine
	// processes all requests via channel dispatch in http/register.go), so only
	// one db.Query() runs at a time. MaxOpenConns(1) matches this architecture.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

// SeedIfEmpty creates the world table and seeds 10k rows if empty.
func SeedIfEmpty(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM world").Scan(&count); err == nil && count > 0 {
		return nil
	}
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS world (id INTEGER PRIMARY KEY, randomNumber INTEGER NOT NULL DEFAULT 0)"); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO world (id, randomNumber) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for i := 1; i <= 10000; i++ {
		if _, err := stmt.Exec(i, rand.Intn(10000)+1); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
