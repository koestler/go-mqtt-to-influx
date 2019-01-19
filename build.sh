VERSION=`git symbolic-ref -q --short HEAD || git describe --tags --exact-match`
OOS=linux go build -ldflags="-s -w -X main.buildVersion=$VERSION -X main.buildTime=`date -Is`"
upx -qqq go-mqtt-to-influx
