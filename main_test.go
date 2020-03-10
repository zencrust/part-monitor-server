package main

import (
	"fmt"
	"os"
	"testing"
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
	d1 := MqttTable{
		AlertId:       "123456",
		Alert:         "Breakdown",
		AcknowledgeBy: "User1",
		AlertType:     "Tyep1",
		InitiateTime:  "2006-01-02 15:04:05",
		Location:      "Location1",
		SlaLevel:      1,
	}
	d2 := MqttTable{
		AlertId:       "123457",
		Alert:         "Breakdown",
		AcknowledgeBy: "User2",
		AlertType:     "Tyep2",
		InitiateTime:  "2006-01-02 15:04:05",
		Location:      "Location2",
		SlaLevel:      1,
	}

	d3 := MqttTable{
		AlertId:       "123458",
		Alert:         "Breakdown",
		AcknowledgeBy: "User3",
		AlertType:     "Tyep3",
		InitiateTime:  "2006-01-02 15:04:05",
		Location:      "Location3",
		SlaLevel:      1,
	}

	d4 := MqttTable{
		AlertId:       "123459",
		Alert:         "Breakdown",
		AcknowledgeBy: "User4",
		AlertType:     "Tyep4",
		InitiateTime:  "2006-01-02 15:04:05",
		Location:      "Location4",
		SlaLevel:      1,
	}

	err = sql.WriteData(d1)
	if err != nil {
		t.Error(err)
		return
	}
	err = sql.WriteData(d2)
	if err != nil {
		t.Error(err)
		return
	}
	err = sql.WriteData(d3)
	if err != nil {
		t.Error(err)
		return
	}
	err = sql.WriteData(d4)
	if err != nil {
		t.Error(err)
		return
	}

	results, err := sql.ReadData(10, 0)
	if err != nil {
		t.Error(err)
		return
	}

	if len(results) != 4 {
		t.Error("Didnt get exactly 4 records which was appeneded.", "got:", len(results))
	}

	for _, res := range results {
		fmt.Printf("%s, %s, %s, %s\n", res.Location, res.InitiateTime, res.Alert, res.AlertType)
	}

}
