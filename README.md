# part-monitor-server
[![Go Report Card](https://goreportcard.com/badge/github.com/zencrust/part-monitor-server)](https://goreportcard.com/report/github.com/zencrust/part-monitor-server)

docker run -it -d --name "part-monitor-server" --restart unless-stopped -e MQTT_SERVER_ADDRESS=tcp://localhost:1883 -e DATABASE_PATH=/mnt/partmon.db -v ~/part-monitor-server:/mnt/ --network=host karthikr19/part-monitor-server:latest