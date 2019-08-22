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

var defaultMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func getMessageHandler(sql *SQLDB, c mqtt.Client) mqtt.MessageHandler {
	onData := make(map[string]int64)
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Println(err)
	}
	return func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("%s %s\n", msg.Topic(), msg.Payload())
		//Application Name/Station Name/function/name
		// for example partmon/Station 1/dio/value
		arr := strings.Split(msg.Topic(), "/")
		if len(arr) != 4 {
			log.Println("unknown topic format", msg.Topic())
			return
		}

		// if arr[2] == "telemetry" || arr[2] == "dio" {

		if arr[3] == "wifi Signal Strength" {
			fmt.Println("publish last seen")
			topicLastseen := arr[0] + "/" + arr[1] + "/" + arr[2] + "/" + "last update time"
			tot := c.Publish(topicLastseen, 0, true, strconv.FormatInt(time.Now().Unix(), 10))
			tot.Wait()
			if tot.Error() != nil {
				log.Println(tot.Error())
			}
		}

		if arr[2] == "telemetry" {
			return
		}

		if arr[3] == "Swicth Pressed" {
			// topic not required here
			return
		}

		//n := bytes.Index(msg.Payload(), []byte{0})
		s := string(msg.Payload())
		currentPacketDevice, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			log.Println("unknown value format", s)
			return
		}

		var currentPacket int64
		topicModified := arr[0] + "/" + arr[1] + "/" + arr[2] + "/" + "Swicth Pressed"
		if currentPacketDevice > 0 {
			currentPacket = time.Now().Unix()
		}
		log.Println("mqtt pub", topicModified, currentPacketDevice)
		c.Publish(topicModified, 0, true, strconv.FormatInt(currentPacket, 10)).Wait()

		device := arr[1]
		//on packet

		if previousPacket, ok := onData[device]; ok {
			//value transition from high to low. log total time and delete from available station
			if currentPacket == 0 && previousPacket > 1566129872 {
				log.Println("writing packet to db")
				tm := time.Unix(previousPacket, 0).In(loc)
				secs := time.Now().Unix()
				err := sql.WriteData(device, tm, float32(secs-previousPacket), "")
				if err != nil {
					log.Println(err)
				}
			}
		}

		onData[device] = currentPacket
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
		log.Println("mqtt server error")
		return c, token.Error()
	}
	fmt.Println("mqtt connected")
	// subs := map[string]byte{"partmon/temp/Tank 1": 0}
	token := c.Subscribe("partalarm/#", 0, getMessageHandler(sql, c))

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
			fmt.Fprintf(w, "422- Query parameters not supplied")
			return
		}
		limit, err1 := strconv.ParseUint(limitStr, 10, 16)
		offset, err2 := strconv.ParseUint(offsetStr, 10, 16)

		if err1 != nil || err2 != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "422- Query parameters not valid")
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
	//knt := 0
	databasePath, found := os.LookupEnv("DATABASE_PATH")
	if !found {
		log.Println("using default database path")
		databasePath = "./pythonsqlite.db"
	}

	mqttServer, found := os.LookupEnv("MQTT_SERVER_ADDRESS")
	if !found {
		log.Println("using default mqtt server address")
		mqttServer = "tcp://localhost:1883"
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
	//for 1 == 1{
	//	time.Sleep(1 * time.Second)
	//}
	//cha := make(chan os.Signal, 1)
	//<- cha
	fmt.Println("starting http server...")
	log.Fatal(http.ListenAndServe(":9503", nil))
	c.Disconnect(250)
}
