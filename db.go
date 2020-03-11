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
	ID              int    `csv:"Id"`
	AlertId         string `csv:"AlertId"`
	Alert           string `csv:"Alert"`
	AlertType       string `csv:"AlertType"`
	Location        string `csv:"Location"`
	InitiatedBy     string `csv:"InitiatedBy"`
	AcknowledgeBy   string `csv:"AcknowledgeBy"`
	ResolvedBy      string `csv:"ResolvedBy"`
	InitiateTime    string `csv:"InitiateTime"`
	IsActive        bool   `csv:"IsActive"`
	AcknowledgeTime string `csv:"AcknowledgeTime"`
	ResolvedTime    string `csv:"ResolvedTime"`
	SlaLevel        int    `csv:"SlaLevel"`
}

type MqttTable struct {
	AlertId         string
	Alert           string
	AlertType       string
	Location        string
	InitiatedBy     string
	AcknowledgeBy   string
	ResolvedBy      string
	InitiateTime    string
	IsActive        bool
	AcknowledgeTime string
	ResolvedTime    string
	SlaLevel        int
}

// CsvTable report data model
type CsvTable struct {
	AlertId         string    `csv:"AlertId"`
	Alert           string    `csv:"Alert"`
	AlertType       string    `csv:"AlertType"`
	Location        string    `csv:"Location"`
	InitiatedBy     string    `csv:"InitiatedBy"`
	AcknowledgeBy   string    `csv:"AcknowledgeBy"`
	ResolvedBy      string    `csv:"ResolvedBy"`
	InitiateTime    time.Time `csv:"InitiateTime"`
	AcknowledgeTime time.Time `csv:"AcknowledgeTime"`
	ResolvedTime    time.Time `csv:"ResolvedTime"`
	SlaLevel        int       `csv:"SlaLevel"`
}

const createTableDef = `CREATE TABLE IF NOT EXISTS partmon_report (
    id        INTEGER  PRIMARY KEY,
	AlertId      text  NOT NULL,
	Alert      text  NOT NULL,
    AlertType      text  NOT NULL,
    Location      text  NOT NULL,
    InitiatedBy      text  NOT NULL,
    AcknowledgeBy      text ,
    ResolvedBy      text ,
    InitiateTime      DATETIME  NOT NULL,
    AcknowledgeTime      DATETIME,
    ResolvedTime      DATETIME,
    SlaLevel  INTEGER
);`

const writeDef = `INSERT INTO partmon_report(AlertId,Alert,AlertType, 
	Location, InitiatedBy, AcknowledgeBy, ResolvedBy, InitiateTime, 
	AcknowledgeTime, ResolvedTime, SlaLevel)
            VALUES(?,?,?,?,?,?,?,?,?,?,?);`

const readDef = `SELECT
 * FROM partmon_report
 ORDER BY InitiateTime DESC
 LIMIT ? OFFSET ?;`

const readDefDate = `SELECT
 * FROM partmon_report
 where InitiateTime >= ? AND InitiateTime <= ?
 ORDER BY InitiateTime DESC;`

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
		err = rows.Scan(&table.ID,
			&table.AlertId,
			&table.Alert,
			&table.AlertType,
			&table.Location,
			&table.InitiatedBy,
			&table.AcknowledgeBy,
			&table.ResolvedBy,
			&table.InitiateTime,
			&table.AcknowledgeTime,
			&table.ResolvedTime,
			&table.SlaLevel)

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
func (sql *SQLDB) ReadtimeData(startTime string, endTime string) ([]CsvTable, error) {
	rows, err := sql.db.Query(readDefDate, startTime, endTime)
	if err != nil {
		return nil, err
	}

	tables := make([]CsvTable, 0, 100)
	rows.Scan()
	for rows.Next() {
		table := CsvTable{}
		// var StartTime time.Time
		var Id int
		err = rows.Scan(&Id,
			&table.AlertId,
			&table.Alert,
			&table.AlertType,
			&table.Location,
			&table.InitiatedBy,
			&table.AcknowledgeBy,
			&table.ResolvedBy,
			&table.InitiateTime,
			&table.AcknowledgeTime,
			&table.ResolvedTime,
			&table.SlaLevel)

		if err != nil {
			return tables, err
		}

		// table.StartDate = StartTime.Format("2006-01-02")
		// table.StartTime = StartTime.Format("15:04:05")
		tables = append(tables, table)
	}

	err = rows.Close()
	if err != nil {
		return tables, err
	}

	return tables, nil
}

// WriteData writes data to Sqlite3 report table
func (sql *SQLDB) WriteData(data MqttTable) error {
	layout := "01/Mar/2006 03:04:05 PM"
	InitiateTime, _ := time.Parse(layout, data.InitiateTime)
	AcknowledgeTime, _ := time.Parse(layout, data.AcknowledgeTime)
	ResolvedTime, _ := time.Parse(layout, data.ResolvedTime)

	_, err := sql.db.Exec(writeDef,
		data.AlertId,
		data.Alert,
		data.AlertType,
		data.Location,
		data.InitiatedBy,
		data.AcknowledgeBy,
		data.ResolvedBy,
		InitiateTime,
		AcknowledgeTime,
		ResolvedTime,
		data.SlaLevel)
	if err != nil {
		return err
	}

	return nil
}

// Close closes database connection
func (sql *SQLDB) Close() {
	sql.db.Close()
}
