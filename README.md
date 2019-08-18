# part-monitor-server
[![Go Report Card](https://goreportcard.com/badge/github.com/zencrust/part-monitor-server)](https://goreportcard.com/report/github.com/zencrust/part-monitor-server)

hostip=$(ip addr show wlan0 | awk '$1 == "inet" {gsub(/\/.*$/, "", $2); print $2}')


docker run -it -d --name "part-monitor-server" --restart unless-stopped -e MQTT_SERVER_ADDRESS=tcp://$hostip:1883 -p9503:9503 -e DATABASE_PATH=/mnt/partmon.db -v ~/part-monitor-server:/mnt/ --network=bridge karthikr19/part-monitor-server:latest