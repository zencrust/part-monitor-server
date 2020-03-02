package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gocarina/gocsv"
)

var defaultMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func getMessageHandler(sql *SQLDB, c mqtt.Client, onData map[string]int64) mqtt.MessageHandler {

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Println(err)
	}
	return func(client mqtt.Client, msg mqtt.Message) {
		// fmt.Printf("%s %s\n", msg.Topic(), msg.Payload())
		//Application Name/Station Name/function/name
		// for example partmon/Station 1/dio/value
		arr := strings.Split(msg.Topic(), "/")

		if len(arr) == 4 && arr[3] == "wifi Signal Strength" {
			// fmt.Println("publish last seen")
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

		if len(arr) != 4 {
			log.Println("unknown topic format", msg.Topic())
			return
		}

		//n := bytes.Index(msg.Payload(), []byte{0})
		s := string(msg.Payload())
		currentPacket, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			log.Println("unknown value format", s)
			return
		}

		device := arr[1]
		//on packet

		fmt.Printf("%s %s\n", msg.Topic(), msg.Payload())
		if previousPacket, ok := onData[device]; ok {
			//value transition from high to low. log total time and delete from available station
			if currentPacket == 0 && previousPacket > 5 {
				log.Println("writing packet to db")
				tm := time.Unix(time.Now().Unix()-previousPacket, 0).In(loc)
				// secs := time.Now().Unix()
				err := sql.WriteData(device, tm, float32(previousPacket), "")
				if err != nil {
					log.Println(err)
				}
			}
		}

		onData[device] = currentPacket
	}
}

func mqttInit(brokerAddress string, applicationName string, sql *SQLDB) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().AddBroker(brokerAddress).SetClientID("dbstoreinstance")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(defaultMessageHandler)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
	onData := make(map[string]int64)
	opts.OnConnect = func(cl mqtt.Client) {
		token := cl.Subscribe(fmt.Sprintf("%s/#", applicationName), 0, getMessageHandler(sql, cl, onData))

		if token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Println("mqtt server error")
		return c, token.Error()
	}
	fmt.Println("mqtt connected")
	// subs := map[string]byte{"partmon/temp/Tank 1": 0}

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

func getDateReportHandler(sql *SQLDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("getDateReportHandler GET params were:", r.URL.Query())
		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		if fromStr == "" || toStr == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "422- Query parameters not supplied")
			return
		}

		table, err := sql.ReadtimeData(fromStr, toStr)
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

		csvContent, err := gocsv.MarshalBytes(&table)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500- something happened while getting the data you have requested")
			return
		}

		w.Write(csvContent)
	}
}

func dbserve(databasePath string) {
	go func() {
		sql, err := Opendb(databasePath)

		if err != nil {
			log.Fatal(err)
		}
		defer sql.Close()

		http.HandleFunc("/api/v1/getreport", getReportHandler(sql))
		http.HandleFunc("/api/v1/getTimereport", getDateReportHandler(sql))

		fmt.Println("starting http server...")
		log.Fatal(http.ListenAndServe(":9503", nil))
	}()
}

func main() {

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	databasePath, found := os.LookupEnv("DATABASE_PATH")
	if !found {
		log.Println("using default database path")
		databasePath = "./partmon.db"
	}

	mqttServer, found := os.LookupEnv("MQTT_SERVER_ADDRESS")
	if !found {
		log.Println("using default mqtt server address")
		// mqttServer = "localhost:1883"
		mqttServer = "broker.hivemq.com:1883"

	}

	applicationName, found := os.LookupEnv("APPNAME")
	if !found {
		log.Println("using default mqtt server address")
		// mqttServer = "localhost:1883"
		applicationName = "partalarm"
	}

	dbserve(databasePath)
	sql, err := Opendb(databasePath)

	if err != nil {
		log.Fatal(err)
	}
	defer sql.Close()

	for {
		c, err := mqttInit(mqttServer, applicationName, sql)
		if err != nil {
			log.Println("mqtt error\n", err)

			if c != nil {
				c.Disconnect(100)
			}
		} else {
			break
		}
	}

	<-ch

}
