package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"
)

type mqttPayLoad struct {
	Value     bool    `json:"value"`
	Duration  float32 `json:"duration"`
	startDate time.Time
}

var defaultMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func getMessageHandler(sql *SQLDB) mqtt.MessageHandler {
	var onData map[string]*mqttPayLoad
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("%s %s\n", msg.Topic(), msg.Payload())
		currentPacket := mqttPayLoad{}
		err := json.Unmarshal(msg.Payload(), &currentPacket)
		if err != nil {
			log.Println(err)
			return
		}
		arr := strings.Split(msg.Topic(), "/")

		//Application Name/Station Name/function/name
		// for example partmon/Station 1/dio/value
		if len(arr) != 4 {
			log.Println("unknown topic format", msg.Topic())
			return
		}
		if arr[3] != "value" {
			// topic not required here
			return
		}

		device := arr[1]
		if val, ok := onData[device]; ok {
			//value transistion from high to low. log total time and delete from available station
			if !currentPacket.Value {
				delete(onData, device)
				err := sql.WriteData(device, val.startDate, val.Duration, "")
				if err != nil {
					log.Println(err)
				}
			} else {
				// value is high and its just an update. update current duration alone
				val.Duration = currentPacket.Duration
			}

		} else if currentPacket.Value {
			// value is not availble in onData i.e value before was false. This is starting trigger from low to high
			currentPacket.startDate = time.Now().In(loc)
			onData[device] = &currentPacket
		}
	}
}

func mqttInit(brokerAddress string, sql *SQLDB) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().AddBroker(brokerAddress).SetClientID("dbstoreinstance")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(defaultMessageHandler)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return c, token.Error()
	}

	// subs := map[string]byte{"partmon/temp/Tank 1": 0}
	token := c.Subscribe("partmon/#", 0, getMessageHandler(sql))

	if token.Wait() && token.Error() != nil {
		return c, token.Error()
	}

	return c, nil
}

func getReportHandler(sql *SQLDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET params were:", r.URL.Query())
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")
		if limitStr == "" || offsetStr == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "422- Query paramaters not supplied")
			return
		}
		limit, err1 := strconv.ParseUint(limitStr, 10, 16)
		offset, err2 := strconv.ParseUint(offsetStr, 10, 16)

		if err1 != nil || err2 != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "422- Query paramaters not valid")
			return
		}

		if limit > 100 {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			fmt.Fprintf(w, "413- given length is too high")
			return
		}

		table, err := sql.ReadData(uint16(limit), uint16(offset))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500- something happened while getting the data you have requested")
			return
		}
		if len(table) == 0 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "{}")
			return
		}

		v, err := json.Marshal(table)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500- something happened while getting the data you have requested")
			return
		}

		w.Write(v)
	}
}

func main() {
	databasePath, found := os.LookupEnv("DATABASE_PATH")
	if !found {
		log.Println("using default database path")
		databasePath = "./goprog.db"
	}

	mqttServer, found := os.LookupEnv("MQTT_SERVER_ADDRESS")
	if !found {
		log.Println("using default mqtt server address")
		mqttServer = "tcp://zencrust.cf:1883"
	}

	sql, err := Opendb(databasePath)

	if err != nil {
		log.Fatal(err)
	}
	defer sql.Close()

	c, err := mqttInit(mqttServer, sql)
	if err != nil {
		log.Println(err)
	}

	http.HandleFunc("/api/v1/getreport", getReportHandler(sql))
	fmt.Println("starting http server...")
	log.Fatal(http.ListenAndServe(":9503", nil))
	c.Disconnect(250)
}
