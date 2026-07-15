//go:build cffi

package sqlbridge

/*
#include <stdlib.h>
*/
import "C"
import (
	"database/sql"
	"encoding/json"
	"unsafe"

	_ "modernc.org/sqlite"
)

//export SQLOpen
func SQLOpen(path *C.char) C.int {
	goPath := C.GoString(path)
	db, err := OpenDB(goPath)
	if err != nil {
		return -1
	}
	id := NewDBHandle(db)
	return C.int(id)
}

//export SQLClose
func SQLClose(handle C.int) {
	CloseDB(int64(handle))
}

//export SQLSeed
func SQLSeed(handle C.int) C.int {
	db := GetDB(int64(handle))
	if db == nil {
		return -1
	}
	if err := SeedIfEmpty(db); err != nil {
		return -1
	}
	return 0
}

//export SQLQueryRow
func SQLQueryRow(handle C.int, query *C.char) *C.char {
	goQuery := C.GoString(query)
	result, err := QuerySingleRow(int64(handle), goQuery)
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}
	b, _ := json.Marshal(result)
	return C.CString(string(b))
}

//export SQLExec
func SQLExec(handle C.int, query *C.char) C.int {
	goQuery := C.GoString(query)
	if err := Exec(int64(handle), goQuery); err != nil {
		return -1
	}
	return 0
}

func init() {
	_ = unsafe.Pointer(nil)
	_ = (*sql.DB)(nil)
}
