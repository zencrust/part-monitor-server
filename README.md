# part-monitor-server
_________________

[![Go Report Card](https://goreportcard.com/badge/github.com/zencrust/part-monitor-server)](https://goreportcard.com/report/github.com/zencrust/part-monitor-server)
[![Build Status](https://dev.azure.com/automationkarthik/partalarm/_apis/build/status/zencrust.part-monitor-server?branchName=master)](https://dev.azure.com/automationkarthik/partalarm/_build/latest?definitionId=11&branchName=master)
_________________
docker run -it -d --name "part-monitor-server" --restart unless-stopped -e MQTT_SERVER_ADDRESS=tcp://localhost:1883 -p9503:9504 -e DATABASE_PATH=/mnt/partmon2.db -v ~/part-monitor-server:/mnt/ --network=host karthikr19/part-monitor-server:latest
