OOS=linux go build -ldflags="-s -w"
upx -qqq go-mqtt-to-influxdb
