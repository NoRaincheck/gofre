package sqlbridge

import (
	"database/sql"
	"fmt"
	"sync"
)

var (
	dbs   sync.Map
	dbSeq int64
	dbMu  sync.Mutex
)

// NewDBHandle stores a database connection and returns a handle.
func NewDBHandle(db *sql.DB) int64 {
	dbMu.Lock()
	defer dbMu.Unlock()
	dbSeq++
	dbs.Store(dbSeq, db)
	return dbSeq
}

// GetDB retrieves a database connection by handle.
func GetDB(handle int64) *sql.DB {
	v, _ := dbs.Load(handle)
	if v == nil {
		return nil
	}
	return v.(*sql.DB)
}

// CloseDB removes a database handle.
func CloseDB(handle int64) {
	if v, ok := dbs.Load(handle); ok {
		v.(*sql.DB).Close()
		dbs.Delete(handle)
	}
}

// QuerySingleRow executes a query returning one row as a map.
func QuerySingleRow(handle int64, query string, args ...interface{}) (map[string]interface{}, error) {
	db := GetDB(handle)
	if db == nil {
		return nil, fmt.Errorf("invalid db handle %d", handle)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("no rows returned")
	}
	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}
	if err := rows.Scan(valPtrs...); err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(cols))
	for i, col := range cols {
		result[col] = vals[i]
	}
	return result, nil
}

// QueryRows executes a query returning multiple rows as a slice of maps.
func QueryRows(handle int64, query string, args ...interface{}) ([]map[string]interface{}, error) {
	db := GetDB(handle)
	if db == nil {
		return nil, fmt.Errorf("invalid db handle %d", handle)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = vals[i]
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// Exec executes a query without returning rows.
func Exec(handle int64, query string, args ...interface{}) error {
	db := GetDB(handle)
	if db == nil {
		return fmt.Errorf("invalid db handle %d", handle)
	}
	_, err := db.Exec(query, args...)
	return err
}
