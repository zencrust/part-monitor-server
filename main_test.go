package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestOpenDb(t *testing.T) {
	sql, err := Opendb("./goprogTest.db")
	if err != nil {
		t.Error(err)
		return
	}

	defer sql.Close()
}

func TestEmptyOpenDb(t *testing.T) {
	sql, err := Opendb("")
	if err == nil {
		t.Error("Empty path should throw error")
		return
	}

	if sql != nil {
		defer sql.Close()
	}
}

func TestData(t *testing.T) {
	os.Remove("./goprogTest.db")

	sql, err := Opendb("./goprogTest.db")

	if err != nil {
		t.Error(err)
		return
	}

	defer sql.db.Close()

	err = sql.WriteData("test1", time.Now(), 33, "")
	if err != nil {
		t.Error(err)
		return
	}

	err = sql.WriteData("test2", time.Now(), 55, "")
	if err != nil {
		t.Error(err)
		return
	}

	err = sql.WriteData("test3", time.Now(), 90, "")
	if err != nil {
		t.Error(err)
		return
	}

	err = sql.WriteData("test4", time.Now(), 33, "")
	if err != nil {
		t.Error(err)
		return
	}

	results, err := sql.ReadData(10, 0)
	if err != nil {
		t.Error(err)
		return
	}

	err = sql.WriteData("test5", time.Now(), 90, "")
	if err != nil {
		t.Error(err)
		return
	}

	err = sql.WriteData("test6", time.Now(), 33, "")
	if err != nil {
		t.Error(err)
		return
	}

	results, err = sql.ReadData(10, 0)
	if err != nil {
		t.Error(err)
		return
	}

	if len(results) != 6 {
		t.Error("Didnt get exactly 4 records which was appeneded.", "got:", len(results))
	}

	for _, res := range results {
		fmt.Printf("%s, %s, %6.1f, %s\n", res.Name, res.StartTime, res.Duration, res.Comments)
	}

}
