package eapx

/*
#cgo pkg-config: libmdbsql glib-2.0
#include <stdlib.h>
#include <glib.h>
#include <mdbsql.h>

// The mdbtools public API exposes several things as C macros or as direct
// struct-field access. cgo cannot see macros and cannot dereference GLib
// container macros, so we wrap every access we need in a real function.

static int mdbx_has_error(MdbSQL *sql) {
	return mdb_sql_has_error(sql);
}

static const char *mdbx_last_error(MdbSQL *sql) {
	return mdb_sql_last_error(sql);
}

static unsigned int mdbx_num_columns(MdbSQL *sql) {
	return sql->num_columns;
}

static const char *mdbx_column_name(MdbSQL *sql, unsigned int i) {
	MdbSQLColumn *col = g_ptr_array_index(sql->columns, i);
	return col->name;
}

// Returns the string value bound for column i of the current row.
// mdbtools always binds values as NUL-terminated C strings.
static const char *mdbx_value(MdbSQL *sql, unsigned int i) {
	return (const char *)g_ptr_array_index(sql->bound_values, i);
}

static long mdbx_row_count(MdbSQL *sql) {
	return sql->row_count;
}

static int mdbx_fetch_row(MdbSQL *sql) {
	return mdb_sql_fetch_row(sql, sql->cur_table);
}

static int mdbx_has_result(MdbSQL *sql) {
	return sql->cur_table != NULL;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// mdbSQL is a thin ownership wrapper around a C MdbSQL engine handle bound to
// one database file. It is not safe for concurrent use; callers serialize
// access.
type mdbSQL struct {
	handle *C.MdbSQL
}

// mdbOpen initializes an MdbSQL engine and connects it to the file at path.
func mdbOpen(path string) (*mdbSQL, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	// mdb_sql_init allocates with g_malloc0, which aborts rather than returning
	// nil, so there is no nil case to guard here.
	handle := C.mdb_sql_init()

	if C.mdb_sql_open(handle, cPath) == nil {
		msg := C.GoString(C.mdbx_last_error(handle))
		C.mdb_sql_exit(handle)
		if msg == "" {
			msg = "unable to open database"
		}
		return nil, fmt.Errorf("eapx: %s", msg)
	}

	return &mdbSQL{handle: handle}, nil
}

// close releases the engine and the underlying database handle. Connector.Close
// calls this exactly once, on a live handle, then drops its reference to the
// mdbSQL value.
func (m *mdbSQL) close() {
	C.mdb_sql_exit(m.handle)
}

// query runs a single SQL statement and returns its full result set. The engine
// is reset afterwards so the handle can be reused for the next statement.
// Connector.Query guarantees the handle is live before calling this.
func (m *mdbSQL) query(sql string) (*ResultSet, error) {
	cSQL := C.CString(sql)
	defer C.free(unsafe.Pointer(cSQL))

	C.mdb_sql_run_query(m.handle, cSQL)

	if C.mdbx_has_error(m.handle) != 0 {
		msg := C.GoString(C.mdbx_last_error(m.handle))
		C.mdb_sql_reset(m.handle)
		return nil, fmt.Errorf("eapx: %s", msg)
	}
	if C.mdbx_has_result(m.handle) == 0 {
		C.mdb_sql_reset(m.handle)
		return nil, fmt.Errorf("eapx: statement produced no result set")
	}

	numCols := int(C.mdbx_num_columns(m.handle))
	rs := &ResultSet{Columns: make([]string, numCols)}
	for i := 0; i < numCols; i++ {
		rs.Columns[i] = C.GoString(C.mdbx_column_name(m.handle, C.uint(i)))
	}

	for C.mdbx_fetch_row(m.handle) != 0 {
		row := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			row[i] = C.GoString(C.mdbx_value(m.handle, C.uint(i)))
		}
		rs.Rows = append(rs.Rows, row)
	}

	rs.RowCount = int(C.mdbx_row_count(m.handle))
	C.mdb_sql_reset(m.handle)
	return rs, nil
}
