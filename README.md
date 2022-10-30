# go-mqtt-to-influx
[![Docker Image CI](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/docker-image.yml/badge.svg)](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/docker-image.yml)
[![Run tests](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/test.yml/badge.svg)](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/test.yml)

**This readme is outdated (major version v1) and should be updated to reflect the changes for v2.**


This daemon connects to [MQTT servers](http://mqtt.org/) and stores the received messages 
in an [Influx Database](https://github.com/influxdata/influxdb).

The tool can connect to one or multiple MQTT servers, handles multiple different topics/data formats and
saves the data in one or multiple databases.

The tool was originally written for one specific project where the data is measured by Sonoff devices and
Victron Energy battery monitors / solar chargers. The devices send the data to an
[Eclipse Mosquitto](https://github.com/eclipse/mosquitto) MQTT server, `go-mqtt-to-influx` writes the data
to a local Influx Db Server for making it available by [Grafana](https://grafana.com/) and a second Influx Db
Server for long term data storage. It, therefore, supports the following inputs:
* Telemetry and sensor data produced by [go-iotdevice](https://github.com/koestler/go-iotdevice)
* Telemetry and sensor data produced by devices running [Sonoff-Tasmota](https://github.com/arendst/Sonoff-Tasmota)

However, you are more than welcome to help support new devices. Send push requests of converters including some tests
or open an issue including examples of topics and messages.

The tool consists of the following components:
* **mqttClient**: connects to a MQTT Server and *receives* messages
* **converter**: parses the message topics and bodies and converts them into influx data points
* **influxClient**: connects to an Influx Database Server and *writes* data to it
* **httpServer**: optional module to output statistics
* **statistics**: optional module to compute statistics about handles messages and rates

## Basic Usage
```
Usage:
  go-mqtt-to-influx [-c <path to yaml config file>]

Application Options:
      --version     Print the build version and timestamp
  -c, --config=     Config File in yaml format (default: ./config.yaml)
      --cpuprofile= write cpu profile to <file>
      --memprofile= write memory profile to <file>

Help Options:
  -h, --help        Show this help message
```

Setup like this:
```bash
# create configuration files
mkdir -p /srv/dc/mqtt-to-influx
cd /srv/dc/mqtt-to-influx
curl https://raw.githubusercontent.com/koestler/go-mqtt-to-influx/main/documentation/docker-compose.yml -o docker-compose.yml
curl https://raw.githubusercontent.com/koestler/go-mqtt-to-influx/main/documentation/config.yaml -o config.yaml
# edit config.yaml

# start it
docker-compose up -d
```

## Config
The Configuration is stored in one yaml file. There are mandatory fields and there are optional fields which
have a default value. 

### Minimalistic Example
```yaml
# documentation/config.yaml

Version: 0

HttpServer:
  Bind: 0.0.0.0
  LogRequests: True

LocalDb:
  Path: "/app/db/local.db"

Statistics:

LogConfig: True
LogWorkerStart: True

MqttClients:
  local-mosquitto:                                         # connect to a local mosquitto server allowing for anonymous connections
    Broker: "tcp://mqtt.exampel.com:1883"

  ttn:                                                     # connect to the things network mqtt server
    Broker: "ssl://eu1.cloud.thethings.network:8883"
    ProtocolVersion: 3
    User: project@ttn
    Password: "NNSXS.VM5PRxxx"
    TopicPrefix: "v3/project@ttn/"
    AvailabilityTopic: ""                                  # disable availability topic on ttn, nobody will listen for it
    Qos: 0                                                 # ttn only supports QOS = 0

InfluxClients:
  example-influx:
    Url: "http://influx.example.com:8086"
    Token: "pfYLu9SjvgblMFL5jzNepJ7PHpKsTjAeVmAMCYHll3BH2cNW5bIz7AdrIbfnsH0tXKcQU9JGr8K-LB1Vdpupmg=="
    Org: "iot"
    Bucket: "iot"

Converters:
  go-iotdevices:
    Implementation: go-iotdevice
    MqttTopics:
      - Topic: "%Prefix%tele/go-iotdevice/%Device%/state"

  tasmota-state:
    Implementation: tasmota-state
    MqttTopics:
      - Topic: "%Prefix%tele/%Device%/STATE"
        Device: "+/+"

  tasmota-sensor:
    Implementation: tasmota-sensor
    MqttTopics:
      - Topic: "%Prefix%tele/%Device%/SENSOR"
        Device: "+/+"

  ttn-dragino:
    Implementation: ttn-dragino
    MqttTopics:
      - Topic: "ttn/devices/%Device%/up"

```  

### Complete, explained example

```yaml
# documentation/full-config.yaml

# Configuration file format version. Always set to 0 since not other format is supported yet (reserved for future use).

# HttpServer: When this section is present, a http server is started.
HttpServer:                                                # optional, default Disabled
  Bind: [::1]                                              # optional, default [::1]; set to 127.0.0.1 for ipv4 loop-back, or [::] / 0.0.0.0 to listen on all ports
  Port: 8000                                               # optional, default 8000; what tcp port the server is listing on; running as root is required when a low port like 80 is used
  LogRequests: False                                       # optional, default False; log all requests to stdout

# LocalDb: When this section is present, a local sqlite-database is created to store a backlog of data waiting to be written to influx
LocalDb:                                                   # optional, default Disabled
  Path: "/app/db/local.db"                                 # optional, default ./go-mqtt-to-influx.db, where to put the file. Use /app/db/XXX when using the docker container.

# Statistics: When this section is enabled, event counters for received, converted and saved events are stored in memory.
# This module might up significant amounts of memory.
# It stores a counter for each mqttTopic, for each mqttClient/influxClient/converter and for each time step.
# The number of time steps is HistoryMaxAge / HistoryResolution.
Statistics:                                                # optional, default Disabled, start the statistics module (needs some additional memory/cpu)
  HistoryResolution: 10s                                   # optional, default 10s, time resolution for aggregation, decrease with caution
  HistoryMaxAge: 10m                                       # optional, default 10min, how many time steps to keep, increase with caution

LogConfig: True                                            # optional, default False, outputs the used configuration including defaults on startup
LogWorkerStart: True                                       # optional, default False, write log for starting / stopping of worker threads

# A map of MQTT servers to connect to
MqttClients:                                               # mandatory, the list must not be empty
  local-mosquitto:                                         # mandatory, an arbitrary name used in log outputs and for reference in the converters section
    Broker: "tcp://mqtt.exampel.com:1883"                  # mandatory, the address / port of the server
    ProtocolVersion: 5                                     # optional, default  5, 3 for mqtt v3.3.x and 5 for mqtt v5.
    User: Bob                                              # optional, if given used for login
    Password: Jeir2Jie4zee                                 # optional, if given used for login
    #ClientId: "go-mqtt-to-influx"                         # optional, default go-mqtt-to-influx-UUID (UUID is generated), client-id sent to the server
    Qos: 1                                                 # optional, default 1, QOS-level used for subscriptions / availability messages must be 0, 1, or 2
    KeepAlive: 60s                                         # optional, default 60s, how often ping messages are sent to the server
    ConnectRetryDelay: 10s                                 # optional, default 10s, how often to try to reestablish a connection to the server
    ConnectTimeout: 5s                                     # optional, default 5s, how long to wait for the connect response, increase on very slow notworks
    AvailabilityTopic: test/%Prefix%tele/%ClientId%/LWT    # optional, if given, a message with "online" / "offline" as payload will be published on connect / disconnect
                                                           # supported placeholders:
                                                           # - %Prefix$   : as specified in TopicPrefix in this config section
                                                           # - %ClientId% : as specified in ClientId this config section
    TopicPrefix: my-project/                               # optional, default empty, used to generate Mqtt Message topics
    LogDebug: False                                        # optional, default False, when enabled debug log of the mqtt client is enabled.
    LogMessages: False                                     # optional, default False, when enabled, all received messages are logged

  ttn:                                                     # optional, a second MQTT server, use The Things Network as an example
    Broker: "ssl://eu1.cloud.thethings.network:8883"
    ProtocolVersion: 3                                     # the things network's MQTT server does not support v5 yet, use v3 instead
    User: project@ttn                                      # see Integrations -> MQTT on thethings.network
    Password: "NNSXS.VM5PRxxx"                             # see Integrations -> MQTT on thethings.network
    TopicPrefix: "v3/project@ttn/"                         # see Integrations -> MQTT on thethings.network
    AvailabilityTopic: ""                                  # disable availability topic on ttn, nobody will listen for it
    Qos: 0                                                 # ttn only supports QOS = 0
    # only for ttn implementation: when KeepAlive interval is set to low, regular reconnects occur. 60s works fine.

# A map of InfluxDB server to send data to
InfluxClients:                                             # mandatory, the list must not be empty
  example-influx:                                          # mandatory, an arbitrary name used in log outputs and for reference in the converters section
    Url: "http://influx.example.com:8086"                  # mandatory, the url to the server
    Token: "pfYLu9SjvgblMFL5jzNepJ7PHpKsTjAeVmAMCYHll3BH2cNW5bIz7AdrIbfnsH0tXKcQU9JGr8K-LB1Vdpupmg=="
                                                           # mandatory, the influxDb API Token
    Org: "iot"                                             # mandatory, the influxDb organisation name
    Bucket: "iot"                                          # mandatory, the influxDb bucket to which the data shall be written
    WriteInterval: 10s                                     # optional, default 10s, defines how often data is sent to the influxDb, in between it is stores in memory.
    RetryInterval: 10s                                     # optional, default 10s, retry after this time when connection fails or on non 2xx-response
    AggregateInterval: 60s                                 # optional, default 60s, how often the local db aggregates multiple data batches into one.
    TimePrecision: 1ms                                     # optional, default 1s, influxDb time precision
    ConnectTimeout: 5s                                     # optional, default 5s, how long to wait for the connect response, increase on very slow notworks
    BatchSize: 5000                                        # optional, default 5000, points are grouped into batches of this size; a batch is sent when it is full or when WriteInterval elapses
    RetryQueueLimit: 20                                    # optional, default 20, discard the oldest batches in the retry queue when this limit is reached (limits memory usage)
    LogDebug: True                                         # optional, default False, outputs the influxDb Line Protocol of each point

  local:                                                   # optional, a second Influx Server
    Url: "http://[::1]:8086"                               # optional, the address of the second Influx server ...
    Token: "xxx"
    Org: "dev"
    Bucket: "dev"

# A map of converters that receive data from mqtt servers and forward points to influxDb servers
# This file contains an example configuration for all available implementations.
Converters:                                                # mandatory, the list must not be empty
  go-iotdevices:                                           # mandatory, an arbitrary name used in log outputs
    Implementation: go-iotdevice                           # mandatory, selects the converter implementation
    MqttTopics:                                            # mandatory, list must not be empty, selects what mqtt subscriptions shall be created for that converter
      - Topic: "%Prefix%tele/go-iotdevice/%Device%/state"  # mandatory, the topic(s) to subscribe to
                                                           #            wildcards + and # might be used
                                                           #            %Prefix% depends on the TopicPrefix defined in the MqttClient config section
                                                           #            %Device% is a placeholder for the deviceName sent to the influxDb
        Device: "+"                                        # optional,  default '+', %Device% in the topic is replaced with this
                                                           #            can be a static value like 'sensor-1'
                                                           #            can be a wildcard + for a single word
                                                           #            can be a wildcard +/+ to match something like a/b and foo/sensor-1
                                                           #            can be # to match an unlimited number of levels
    MqttClients:                                           # defines which mqtt clients this converter shall subscribe to, if omitted or empty, all clients are used
      - local-mosquitto                                    # the arbitrary name defined in the MqttClients configuration section
      - ttn
    InfluxClients:                                         # defines which influxDb clients this converter shall write data to, if omitted or empty, data is sent to all clients
      - example-influx                                     # the arbitrary name defined in the InfluxClients configuration section
      - local
    LogHandleOnce: True                                    # optional, default True, when enabled, the first time this converter is executed, a log message is generated

  ttn-dragino:                                             # mandatory, an arbitrary name used in log outputs
    Implementation: ttn-dragino
    MqttTopics:                                            # mandatory, list must not be empty, selects what mqtt subscriptions shall be created for that converter
      - Topic: "ttn/devices/%Device%/up"
        Device: "+"
    MqttClients:                                           # defines which mqtt clients this converter shall subscribe to, if omitted or empty, all clients are used
      - ttn                                                # e.g. only subscribe on ttn for the dagino sensor updates
    InfluxClients:                                         # defines which influxDb clients this converter shall write data to, if omitted or empty, data is sent to all clients
      - local                                              # e.g. only sends dragino data to the local db since the internet server has another instance of this tool running
    LogHandleOnce: True                                    # optional, default True, when enabled, the first time this converter is executed, a log message is generated

  tasmota-state:                                           # mandatory, an arbitrary name used in log outputs
    Implementation: tasmota-state
    MqttTopics:                                            # mandatory, list must not be empty, selects what mqtt subscriptions shall be created for that converter
      - Topic: "%Prefix%tele/%Device%/STATE"               # e.g. subscribes 'my-project/tele/+/+/STATE' on local-mosquitto
                                                           # and subscribes 'v3/project@ttn/tele/+/+/SATE' on ttn
        Device: "+/+"                                      # e.g. when topic is 'my-project/tele/mezzo/light0/STATE', deviceName=mezzo/light0
    LogHandleOnce: True                                    # optional, default True, when enabled, the first time this converter is executed, a log message is generated

  tasmota-sensor:                                          # mandatory, an arbitrary name used in log outputs
    Implementation: tasmota-sensor
    MqttTopics:                                            # mandatory, list must not be empty, selects what mqtt subscriptions shall be created for that converter
      - Topic: "%Prefix%tele/%Device%/SENSOR"
        Device: "+/+"
    LogHandleOnce: True                                    # optional, default True, when enabled, the first time this converter is executed, a log message is generated


# A list of influxDb tags that should be added depending on the deviceName.
# This is useful to e.g. group sensors by building, by type or so and use this in influxDb queries.
# You can use either use the Matches or the Equals property. A device can match multiple tags. The tags are then merged (last one in the list overwrites)
InfluxAuxiliaryTags:                                       # optional, the list can be omitted or left empty
  - Equals: mezzo/bridge0                                  # Match the device with deviceName=mezzo/bridge0.
    TagValues:                                             # A map of tagName: value to add as influxDb tag
      displayName: Sonoff Bridge                           # adds displayName="Sonoff Bridge"

  - Matches: mezzo/.*                                      # optional, match the deviceName by regular expression
    TagValues:                                             # A map of tagName: value to add as influxDb tag
      area: Mezzo

```  

## Converters
Currently, the following converter implementations exist:

### lwt

LWT (Last Will Topic) Messages are used to broadcast the availability (online/offline) of a device.
This follows the format used by [Tasmota](https://github.com/arendst/Sonoff-Tasmota/wiki/MQTT-Overview).

Example:
* Topic: `piegn/tele/software/srv1-go-iotdevice/LWT`
* Payload: `Online`,
* Output: `boolValue,device=software/srv1-go-iotdevice,field=Available value=true`

### go-iotdevice

**go-iotdevice** can read out various sensor values like voltages, currents, and power from BMV-702 battery monitors and
solar chargers made by [Victron Energy](https://www.victronenergy.com/) and send them to an MQTT server. This
converter can read and parse those.

Example:
* Topic: `piegn/tele/ve/24v-bmv`
* Payload:
```json
{
  "Time":"2019-01-06T23:40:03",
  "NextTele":"2019-01-06T23:40:13",
  "TimeZone":"UTC",
  "Model":"bmv700",
  "Values":{
    "AmountOfChargedEnergy":{"Value":756.6,"Unit":"kWh"},
    "AmountOfDischargedEnergy":{"Value":363.1,"Unit":"kWh"},
    "Consumed":{"Value":-7.2,"Unit":"Ah"},
    "Current":{"Value":-0.7,"Unit":"A"},
    "StateOfCharge":{"Value":99,"Unit":"%"},
    "Power":{"Value":-18,"Unit":"W"},
    "TimeToGo":{"Value":14400,"Unit":"min"}
  }
}
```
* Output lines:
  * `floatValue,device=24v-bmv,field=AmountOfChargedEnergy,sensor=bmv700,unit=kWh value=756.6"`
  * `floatValue,device=24v-bmv,field=AmountOfDischargedEnergy,sensor=bmv700,unit=kWh value=363.1"`
  * `floatValue,device=24v-bmv,field=Consumed,sensor=bmv700,unit=Ah value=-7.2"`
  * `floatValue,device=24v-bmv,field=Current,sensor=bmv700,unit=A value=-0.7"`
  * `floatValue,device=24v-bmv,field=StateOfCharge,sensor=bmv700,unit=% value=99"`
  * `floatValue,device=24v-bmv,field=Power,sensor=bmv700,unit=W value=-18"`
  * `floatValue,device=24v-bmv,field=TimeToGo,sensor=bmv700,unit=min value=14400"`

### tasmota-state
[Tasmota](https://github.com/arendst/Sonoff-Tasmota/wiki/MQTT-Overview) sends state messages whenever a switch
is turned on or off. This messages also include the current uptime of the device, the supply voltage and
details about the current wifi connection. All this data is stored.

Example:
* Topic: `piegn/tele/elektronik/control0/STATE`
* Payload:
```json
{
  "Time":"2019-01-10T22:45:22",
  "Uptime":"9T09:29:01",
  "Vcc":3.108,
  "POWER1":"OFF",
  "POWER2":"ON",
  "POWER3":"OFF",
  "POWER4":"OFF",
  "Wifi":{"AP":1,"SSId":"piegn-iot","BSSId":"04:F0:21:2F:B7:CC","Channel":1,"RSSI":100}
}
```

* Output lines:
  * `timeValue,device=elektronik/control0 value="2019-01-10 22:45:22 +0000 UTC"`
  * `floatValue,device=elektronik/control0,field=UpTime,unit=s value=811741`
  * `floatValue,device=elektronik/control0,field=Vcc,unit=V value=3.108`
  * `boolValue,device=elektronik/control0,field=Power1 value=false`
  * `boolValue,device=elektronik/control0,field=Power2 value=true`
  * `boolValue,device=elektronik/control0,field=Power3 value=false`
  * `boolValue,device=elektronik/control0,field=Power4 value=false`
  * `wifi,BSSId=04:F0:21:2F:B7:CC,SSId=piegn-iot,device=elektronik/control0 AP=1i,Channel=1i,RSSI=100i`

### tasmota-sensor
[Tasmota](https://github.com/arendst/Sonoff-Tasmota/wiki/MQTT-Overview) sends periodic sensor measurement messages.

Example:
* Topic: `piegn/tele/elektronik/control0/SENSOR`
* Payload: `{"Time":"2019-01-10T22:15:52","SI7021":{"Temperature":5.4,"Humidity":27.7},"TempUnit":"C"}`
* Output lines:
  * `floatValue,device=elektronik/control0,field=Temperature,sensor=SI7021,unit=C value=5.4`
  * `floatValue,device=elektronik/control0,field=Humidity,sensor=SI7021,unit=% value=27.7`

## Deployment
The cpu / memory requirements for this tool are quite minimal but depend on the number of messages to be handled.
I run it on [Raspberry Pi Zero 2 W](https://www.raspberrypi.com/products/raspberry-pi-zero-2-w/) and on
[PC Engines APU 2](https://www.pcengines.ch/apu2.htm) in the field as well as X86 virtual servers in the cloud.

There are [github actions](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/docker-image.yml)
to automatically cross-compile amd64, arm64 and arm/v7
[docker images](https://hub.docker.com/r/koestler/go-mqtt-to-influx/tags).

The github tags use semantic versioning and whenever a tag like v2.3.4 is build, it is pushed to docker tags
v2, v2.3 and v2.3.4.

For auto-restart on system reboots, configuration and networking I use `docker compose`. Here is an example file:
```yaml
# documentation/docker-compose.yml

version: "3"
services:
  go-mqtt-to-influx:
    restart: always
    image: koestler/go-mqtt-to-influx:v2
    volumes:
      - /srv/volumes/mqtt-to-influx/db:/app/db
      - ${PWD}/config.yaml:/app/config.yaml:ro
```

Note the mount of /app/db. It makes the database persist recreation of the docker container.
It assumes hat you have the following configuration in config.yaml:
LocalDb:
```yaml
  Path: /app/db/local.db
```

Quick setup:
```bash
mkdir -p /srv/dc/mqtt-to-influx
cd /srv/dc/mqtt-to-influx
curl https://raw.githubusercontent.com/koestler/go-mqtt-to-influx/main/documentation/docker-compose.yml -o docker-compose.yml
curl https://github.com/koestler/go-mqtt-to-influx/blob/main/documentation/config.yaml -o config.yaml
# adapt config.yaml and configure mqtt / influx connection and converters.
docker compose up -d
docker compose logs -f
```

## Development
Development is done on Ubuntu and Mac.
Install [github cli](https://cli.github.com/) and [golang](https://go.dev/doc/install).

### Local development
```bash
gh clone koestler/go-mqtt-to-influx
cd go-mqtt-to-influx
go build
./go-mqtt-to-influx
```

### Compile and run inside docker
```bash
git clone koestler/go-mqtt-to-influx
cd go-mqtt-to-influx
docker build -f docker/Dockerfile -t go-mqtt-to-influx .
docker run --rm --name go-mqtt-to-influx -p 127.0.0.1:8000:8000 \
  -v "$(pwd)"/documentation/config.yaml:/app/config.yaml:ro \
  go-mqtt-to-influx
```

### run tests
```bash
go install github.com/golang/mock/mockgen@v1.6.0
go generate ./...
go test ./...
```

### Update README.md
```bash
npx embedme README.md
```

### Upgrade dependencies
```bash
go get -t -u ./...
go generate ./...
go test ./...
```

# License
MIT License