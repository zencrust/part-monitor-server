package main

import (
	"database/sql"
	"fmt"
	"time"

	// this is required for sqlite3 database
	_ "github.com/mattn/go-sqlite3"
)

// ScanTable report data model
type ScanTable struct {
	ID        int       `csv:"Id"`
	Name      string    `csv:"Station Name"`
	StartTime time.Time `csv:"Start Time"`
	Duration  float32   `csv:"Duration"`
	Comments  string    `csv:"Comments"`
}

const createTableDef = `CREATE TABLE IF NOT EXISTS partmon_report (
    id        INTEGER  PRIMARY KEY,
    name      text  NOT NULL,
    start_time DATETIME NOT NULL,
    duration  REAL     NOT NULL,
    comments  text
);`

const writeDef = `INSERT INTO partmon_report(name,start_time,duration, comments)
            VALUES(?,?,?,?);`

const readDef = `SELECT
 * FROM partmon_report
 ORDER BY start_time DESC
 LIMIT ? OFFSET ?;`

const readDefDate = `SELECT
 * FROM partmon_report
 where start_time >= ? AND start_time <= ?
 ORDER BY start_time DESC;`

// SQLDB main datastructure from database
type SQLDB struct {
	db *sql.DB
}

// Opendb opens or creates sqlite3 database, creates tables and make it ready for
// CRUD operations
func Opendb(filePath string) (*SQLDB, error) {

	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createTableDef)

	ret := &SQLDB{
		db: db,
	}

	if err != nil {
		return ret, err
	}

	return ret, nil
}

// ReadData writes data to Sqlite3 report table
func (sql *SQLDB) ReadData(limit uint16, offset uint16) ([]ScanTable, error) {
	rows, err := sql.db.Query(readDef, limit, offset)
	if err != nil {
		return nil, err
	}

	tables := make([]ScanTable, 0, limit)
	rows.Scan()
	for rows.Next() {
		table := ScanTable{}
		err = rows.Scan(&table.ID, &table.Name, &table.StartTime, &table.Duration, &table.Comments)

		if err != nil {
			return tables, err
		}

		tables = append(tables, table)
	}

	err = rows.Close()
	if err != nil {
		return tables, err
	}

	return tables, nil
}

// ReadtimeData read data
func (sql *SQLDB) ReadtimeData(startTime string, endTime string) ([]ScanTable, error) {
	rows, err := sql.db.Query(readDefDate, startTime, endTime)
	if err != nil {
		return nil, err
	}

	tables := make([]ScanTable, 0, 100)
	rows.Scan()
	for rows.Next() {
		table := ScanTable{}
		err = rows.Scan(&table.ID, &table.Name, &table.StartTime, &table.Duration, &table.Comments)

		if err != nil {
			return tables, err
		}

		tables = append(tables, table)
	}

	err = rows.Close()
	if err != nil {
		return tables, err
	}

	return tables, nil
}

// WriteData writes data to Sqlite3 report table
func (sql *SQLDB) WriteData(name string, startTime time.Time, duration float32, comments string) error {
	_, err := sql.db.Exec(writeDef, name, startTime, duration, comments)
	if err != nil {
		return err
	}

	return nil
}

// Close closes database connection
func (sql *SQLDB) Close() {
	sql.db.Close()
}
