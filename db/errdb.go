package db

import (
	"database/sql"
)

func NewEErrDb(db *sql.DB) *eliminateErrDb {
	tx, err := db.Begin()

	return &eliminateErrDb{db, tx, err, false}
}

type eliminateErrDb struct {
	db    *sql.DB
	tx    *sql.Tx
	err   error
	dirty bool
}

func (tx *eliminateErrDb) Exec(query string, args ...interface{}) sql.Result {
	tx.dirty = true
	if tx.err != nil {
		return nil
	}

	result, err := tx.tx.Exec(query, args...)
	if err != nil {
		tx.err = err
		return nil
	}
	return result
}

func (tx *eliminateErrDb) Commit() error {
	if tx.err != nil {
		_ = tx.tx.Rollback()

		return tx.err
	}

	return tx.tx.Commit()
}

func (tx *eliminateErrDb) Query(query string, args ...interface{}) *ErrorRow {
	if tx.err != nil {
		return nil
	}

	rows, err := tx.db.Query(query, args...)
	if err != nil {
		tx.err = err
		return nil
	}

	return &ErrorRow{tx, rows}
}

type ErrorRow struct {
	tx   *eliminateErrDb
	rows *sql.Rows
}

func (rs *ErrorRow) Next() bool {
	if rs == nil {
		return false
	}

	return rs.rows.Next()
}

func (rs *ErrorRow) Scan(dest ...interface{}) {
	if rs.tx.err != nil {
		return
	}

	err := rs.rows.Scan(dest...)
	if err != nil {
		rs.tx.err = err
	}
}
