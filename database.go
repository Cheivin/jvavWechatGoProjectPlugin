package main

import (
	"database/sql"
	"wechat-hub-plugin/hub"
)

type DB struct {
	db *sql.DB
}

func (d DB) Query(sql string, args ...any) (map[string]any, error) {
	stmt, err := d.db.Prepare(sql)
	defer func() {
		_ = stmt.Close()
	}()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(args...)
	defer func() {
		_ = rows.Close()
	}()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)
	entry := make(map[string]interface{})
	values := make([]interface{}, count)
	valPointers := make([]interface{}, count)
	if rows.Next() {
		for i := 0; i < count; i++ {
			valPointers[i] = &values[i]
		}
		_ = rows.Scan(valPointers...)
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
	}
	return entry, nil

}

func (d DB) QueryAll(sql string, args ...any) ([]map[string]any, error) {
	stmt, err := d.db.Prepare(sql)
	defer func() {
		_ = stmt.Close()
	}()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(args...)
	defer func() {
		_ = rows.Close()
	}()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)
	mData := make([]map[string]any, 0)
	values := make([]interface{}, count)
	valPointers := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valPointers[i] = &values[i]
		}
		_ = rows.Scan(valPointers...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		mData = append(mData, entry)
	}
	return mData, nil

}

func NewDB(db *sql.DB) hub.DBInterface {
	return &DB{db: db}
}
