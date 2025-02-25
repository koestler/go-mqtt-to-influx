# go-mqtt-to-influx
[![Docker Image CI](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/docker-image.yml/badge.svg)](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/docker-image.yml)
[![Run tests](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/test.yml/badge.svg)](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/koestler/go-mqtt-to-influx/v2.svg)](https://pkg.go.dev/github.com/koestler/go-mqtt-to-influx/v2)

This tool connects to one or multiple [MQTT](http://mqtt.org/) servers to receive data from IOT-sensors.
The messages are then parsed using easy to implement device / message specific converters to generate
data points which are then written to an [Influx Database](https://github.com/influxdata/influxdb).

Currently, the following devices are supported:
* All devices connected via [go-iotdevice](https://github.com/koestler/go-iotdevice) which includes
  [Victron Energy](https://www.victronenergy.com/) devices like the [BMV 702](https://www.victronenergy.com/battery-monitors/bmv-702),
  [Shelly EM3 Energy Meter](https://www.shelly.cloud/en-ch/products/product-overview/shelly-3-em),
  [Finder7M38](https://www.findernet.com/en/uk/series/7m-series-smart-energy-meters/type/type-7m-38-three-phase-multi-function-bi-directional-energy-meters-with-backlit-matrix-lcd-display/).
* Devices running the [Sonoff-Tasmota](https://github.com/arendst/Sonoff-Tasmota) firmware.
* Various devices connected via a Lora WAN Network like [The Things Network](https://www.thethingsnetwork.org/).
  * [Dragino LoraWAN sensors](https://www.dragino.com/)
  * [Dragino LSN50-v2](https://www.dragino.com/products/lora-lorawan-end-node/item/155-lsn50-v2.html)
  * [SenseCAP S2120](https://www.seeedstudio.com/sensecap-s2120-lorawan-8-in-1-weather-sensor-p-5436.html)
  * [Fencyboy](https://fencyboy.com/)

You are more than welcome to help support new devices. Send pull requests of converters including some tests
or open an issue including examples of topics and messages.

This tool consists of the following components:
* **mqttClient**: Connects to a MQTT servers
                  using [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) for MQTT v3.3 
                  and [paho.golang/paho](https://github.com/eclipse/paho.golang) for MQTT v5
                  to **receive** raw data.
* **converter**: Parses the message topics and bodies and **converts** them into InfluxDB data points.
* **influxClient**: Connects to an InfluxDB v2 server and **writes** data to it.
* **statistics**: Optional module to measure the flow rate of messages per device / topic / converter / database.
* **httpServer**: Optional module to output statistics and other debug information.
* **localDb**: Optional module to record a backlog of data to a local [Sqlite3](https://www.sqlite.org/) database
               while the InfluxDB is unavailable. The module aggregates small batches into bigger batches to 
               allow for a relatively quick writing of all data once the InfluxDB is back online.

## Deployment
The cpu & memory requirements for this tool are quite minimal but depend on the number of messages to be handled.
All my instances run without issues on a Raspberry Pi Zero 2 W.

<details>
<summary>
Deployment without docker
</summary>
I use docker to deploy this tool.
Alternatively, you can use `go install` to build binary locally.

```bash
go install github.com/koestler/go-mqtt-to-influx/v2@latest
curl https://raw.githubusercontent.com/koestler/go-mqtt-to-influx/main/documentation/config.yaml -o config.yaml
# adapt config.yaml and configure mqtt / influx connection and converters.

# start the tool
go-mqtt-to-influx
```
</details>


### Docker

There are [GitHub actions](https://github.com/koestler/go-mqtt-to-influx/actions/workflows/docker-image.yml)
to automatically cross-compile amd64, arm64 and arm/v7
publicly available [docker images](https://github.com/koestler/go-mqtt-to-influx/pkgs/container/go-mqtt-to-influx).
The docker-container is built on top of alpine, the binary is `/go-mqtt-to-influx` and the config is
expected to be at `/app/config.yaml` and the local-db to be at `/app/db`. The container runs as non-root user `app`.

The GitHub tags use semantic versioning and whenever a tag like v2.3.4 is built, it is pushed to docker tags
v2, v2.3, and v2.3.4.

For auto-restart on system reboots, configuration, and networking I use `docker compose`. Here is an example file:
```yaml
# documentation/docker-compose.yml

version: "3"
services:
  go-mqtt-to-influx:
    restart: always
    image: ghcr.io/koestler/go-mqtt-to-influx:v2
    volumes:
      - ${PWD}/db:/app/db
      - ${PWD}/config.yaml:/app/config.yaml:ro

```

Note the mount of /app/db. It makes the database persist recreation of the docker container.
It assumes hat you have the following configuration in config.yaml:
LocalDb:
```yaml
  Path: /app/db/local.db
```

### Quick setup
[Install Docker](https://docs.docker.com/engine/install/) first.

```bash
# create a directory for the docker-composer project and config file
mkdir -p /srv/dc/mqtt-to-influx # or wherever you want to put docker-compose files
cd /srv/dc/mqtt-to-influx
curl https://raw.githubusercontent.com/koestler/go-mqtt-to-influx/main/documentation/docker-compose.yml -o docker-compose.yml
curl https://raw.githubusercontent.com/koestler/go-mqtt-to-influx/main/documentation/config.yaml -o config.yaml
# adapt config.yaml and configure mqtt / influx connection and converters.

# create the volume for the local db and change permissions
mkdir db && sudo chown 100:100 db

# start the container
docker compose up -d

# optional: check the log output to see how it's going
docker compose logs -f

# when config.yaml is changed, the container needs to be restarter
docker compose restart

# upgrade to the newest tag
docker compose pull
docker compose up
```

## Config
The configuration is stored in a single yaml file. By default, it is read from `./config.yaml`.
This can be changed using the `--config=another-config.yaml` command line option.

There are mandatory fields and there are optional fields which have reasonable default values. 

### Complete, explained example
The following configuration file contains all possible configuration options.

```yaml
# documentation/full-config.yaml

# Configuration file format version. Always set to 0 since not other format is supported yet (reserved for future use).
Version: 0

# HttpServer: When this section is present, a http server is started.
HttpServer:                                                # optional, default Disabled
  Bind: "[::1]"                                            # optional, default [::1]; set to 127.0.0.1 for ipv4 loop-back, or [::] / 0.0.0.0 to listen on all ports
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
    LogHandleOnce: False                                   # optional, default False, when enabled, the first time this converter is executed, a log message is generated

  ttn:                                                     # mandatory, an arbitrary name used in log outputs
    Implementation: ttn
    MqttTopics:                                            # mandatory, list must not be empty, selects what mqtt subscriptions shall be created for that converter
      - Topic: "ttn/devices/%Device%/up"
        Device: "+"
    MqttClients:                                           # defines which mqtt clients this converter shall subscribe to, if omitted or empty, all clients are used
      - ttn                                                # e.g. only subscribe on ttn for the dagino sensor updates
    InfluxClients:                                         # defines which influxDb clients this converter shall write data to, if omitted or empty, data is sent to all clients
      - local                                              # e.g. only sends dragino data to the local db since the internet server has another instance of this tool running
    LogHandleOnce: False                                   # optional, default False, when enabled, the first time this converter is executed, a log message is generated

  tasmota-state:                                           # mandatory, an arbitrary name used in log outputs
    Implementation: tasmota-state
    MqttTopics:                                            # mandatory, list must not be empty, selects what mqtt subscriptions shall be created for that converter
      - Topic: "%Prefix%tele/%Device%/STATE"               # e.g. subscribes 'my-project/tele/+/+/STATE' on local-mosquitto
                                                           # and subscribes 'v3/project@ttn/tele/+/+/SATE' on ttn
        Device: "+/+"                                      # e.g. when topic is 'my-project/tele/mezzo/light0/STATE', deviceName=mezzo/light0
    LogHandleOnce: False                                   # optional, default False, when enabled, the first time this converter is executed, a log message is generated

  tasmota-sensor:                                          # mandatory, an arbitrary name used in log outputs
    Implementation: tasmota-sensor
    MqttTopics:                                            # mandatory, list must not be empty, selects what mqtt subscriptions shall be created for that converter
      - Topic: "%Prefix%tele/%Device%/SENSOR"
        Device: "+/+"
    LogHandleOnce: False                                    # optional, default False, when enabled, the first time this converter is executed, a log message is generated


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

### Minimalistic Example
This a minimalistic configuration example serving as a good starting point.

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

  ttn:
    Implementation: ttn
    MqttTopics:
      - Topic: "ttn/devices/%Device%/up"

  go-iotdevices:
    Implementation: go-iotdevice
    MqttTopics:
      - Topic: "%Prefix%tele/go-iotdevice/%Device%/state"

```  

## Converters

### lwt

LWT (Last Will Topic) Messages are used to broadcast the availability (online/offline) of a device.
This follows the format used by [Tasmota](https://github.com/arendst/Sonoff-Tasmota/wiki/MQTT-Overview).

Example:
* Topic: `piegn/tele/software/srv1-go-iotdevice/LWT`
* Payload: `Online`,
* Output: `boolValue,device=software/srv1-go-iotdevice,field=Available value=true`


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


### go-iotdevice

[go-iotdevice](https://github.com/koestler/go-iotdevice) can read out various sensor values like voltages, currents,
from battery monitors and solar chargers. Whe configured, it can send all those values in a single mqtt message
at regular intervals.

Configuration template:
```yaml
    Implementation: go-iotdevice
    MqttTopics:
      - Topic: "%Prefix%tele/go-iotdevice/%Device%/state"
```

Example:
* Mqtt Topic: `project/tele/go-iotdevice/device-name/state`
* Mqtt Payload:
```json
{
  "Time": "2022-08-19T14:52:19Z",
  "NextTelemetry": "2022-08-19T14:52:24Z",
  "Model": "BMV-702",
  "SecondsSinceLastUpdate": 0.576219653,
  "NumericValues": {
    "AmountOfChargedEnergy": {
      "Value": 1883.52,
      "Unit": "kWh"
    },
    "CurrentHighRes": {
      "Value": -0.58,
      "Unit": "A"
    },
    "NumberOfCycles": {
      "Value": 241,
      "Unit": ""
    },
    "ProductId": {
      "Value": 4261544960,
      "Unit": ""
    },
    "SOC": {
      "Value": 58.16,
      "Unit": "%"
    },
    "TTG": {
      "Value": 5742,
      "Unit": "min"
    },
    "Uptime": {
      "Value": 17182790,
      "Unit": "s"
    }
  },
  "TextValues": {
    "ModelName": {
      "Value": "BMV-702"
    },
    "SerialNumber": {
      "Value": "HQ15149CFQI,HQ1515RP6L7,"
    },
    "SynchronizationState": {
      "Value": "true"
    }
  }
}
```
* InfluxDB line protocol:
  * `clock,device=24v-bmv timeValue="2022-08-19T14:52:19Z"`
  * `telemetry,device=24v-bmv,field=AmountOfChargedEnergy,sensor=BMV-702,unit=kWh floatValue=1883.52`
  * `telemetry,device=24v-bmv,field=CurrentHighRes,sensor=BMV-702,unit=A floatValue=-0.58`
  * `telemetry,device=24v-bmv,field=ModelName,sensor=BMV-702 stringValue="BMV-702"`
  * `telemetry,device=24v-bmv,field=NumberOfCycles,sensor=BMV-702,unit= floatValue=241`
  * `telemetry,device=24v-bmv,field=ProductId,sensor=BMV-702,unit= floatValue=4.26154496e+09`
  * `telemetry,device=24v-bmv,field=SOC,sensor=BMV-702,unit=% floatValue=58.16`
  * `telemetry,device=24v-bmv,field=SerialNumber,sensor=BMV-702 stringValue="HQ15149CFQI,HQ1515RP6L7,"`
  * `telemetry,device=24v-bmv,field=SynchronizationState,sensor=BMV-702 stringValue="true"`
  * `telemetry,device=24v-bmv,field=TTG,sensor=BMV-702,unit=min floatValue=5742`
  * `telemetry,device=24v-bmv,field=Uptime,sensor=BMV-702,unit=s floatValue=1.718279e+07`

### ttn

TTN is a generic converter for messages from [The Things Network](https://www.thethingsnetwork.org/).
It currently support the following devices:
* [Dragino LoraWAN sensors](https://www.dragino.com/)
* [Fencyboy](https://fencyboy.com/)
* [SenseCAP S2120](https://www.seeedstudio.com/sensecap-s2120-lorawan-8-in-1-weather-sensor-p-5436.html)

For all devices, a line for the lora measurement is produced:
* "lora,devEui=2CF7F1C0443003DD,device=s2120-0,gatewayEui=E45F01FFFEDECBE3,gatewayId=piegn-srv3 channelRssi=-80i,consumedAirtimeUs=71936i,gatewayIdx=0i,rssi=-80i,snr=6.5",

Depending on the VersionIDs received, the correct sub-converter is executed. The VersionIDs are only available, when the devices
is taken from the TTN device registry. A fallback to the device name is implemented. 
If you manually create the device, make sure to include "dragino", "fencyboy" or "sensecap" in the device name.

At the moment, the following sub-converters are available:

#### dragino
The exact lines depend on the fields the sensor produces. The current implementation is tested with the LSN50, LHT52 and D20S sensors.

* InfluxDB line protocol:
  * "telemetry,device=lsn50-temp-1,field=AlarmStatus,sensor=lsn50v2-d20-d22-d23 boolValue=false",
  * "telemetry,device=lsn50-temp-1,field=BatV,sensor=lsn50v2-d20-d22-d23,unit=V floatValue=3.655",
  * "telemetry,device=lsn50-temp-1,field=TempBlack,sensor=lsn50v2-d20-d22-d23,unit=°C floatValue=37.5",
  * "telemetry,device=lsn50-temp-1,field=TempRed,sensor=lsn50v2-d20-d22-d23,unit=°C floatValue=28.3",
  * "telemetry,device=lsn50-temp-1,field=TempWhite,sensor=lsn50v2-d20-d22-d23,unit=°C floatValue=20.4",
  * "telemetry,device=lsn50-temp-1,field=WorkMode,sensor=lsn50v2-d20-d22-d23 stringValue=\"DS18B20\"",

#### fencyboy

* InfluxDB line protocol:
  * "telemetry,device=fencyboy-0,field=ActiveMode,sensor=fencyboy boolValue=true",
  * "telemetry,device=fencyboy-0,field=BatteryVoltage,sensor=fencyboy,unit=V floatValue=3.361",
  * "telemetry,device=fencyboy-0,field=FenceVoltage,sensor=fencyboy,unit=V floatValue=10246",
  * "telemetry,device=fencyboy-0,field=FenceVoltageMax,sensor=fencyboy,unit=V floatValue=10368",
  * "telemetry,device=fencyboy-0,field=FenceVoltageMin,sensor=fencyboy,unit=V floatValue=10175",
  * "telemetry,device=fencyboy-0,field=FenceVoltageStd,sensor=fencyboy floatValue=32.2",
  * "telemetry,device=fencyboy-0,field=Impulses,sensor=fencyboy intValue=466i",
  * "telemetry,device=fencyboy-0,field=RemainingCapacity,sensor=fencyboy,unit=mAh floatValue=600.8870239257812",
  * "telemetry,device=fencyboy-0,field=Temperature,sensor=fencyboy,unit=°C floatValue=5.32",

#### senscap

* InfluxDB line protocol:
  * "telemetry,device=s2120-0,field=Air\\ Humidity,sensor=sensecaps2120-8-in-1,unit=%\\ RH floatValue=66",
  * "telemetry,device=s2120-0,field=Air\\ Temperature,sensor=sensecaps2120-8-in-1,unit=°C floatValue=6.5",
  * "telemetry,device=s2120-0,field=Barometric\\ Pressure,sensor=sensecaps2120-8-in-1,unit=Pa floatValue=96920",
  * "telemetry,device=s2120-0,field=Light\\ Intensity,sensor=sensecaps2120-8-in-1,unit=Lux floatValue=0",
  * "telemetry,device=s2120-0,field=Rainfall,sensor=sensecaps2120-8-in-1,unit=mm/h floatValue=0",
  * "telemetry,device=s2120-0,field=UV\\ Index,sensor=sensecaps2120-8-in-1,unit= floatValue=0",
  * "telemetry,device=s2120-0,field=Wind\\ Direction,sensor=sensecaps2120-8-in-1,unit=° floatValue=79",
  * "telemetry,device=s2120-0,field=Wind\\ Speed,sensor=sensecaps2120-8-in-1,unit=m/s floatValue=0",

## Development
Development is done on Ubuntu and Mac.
Install [GitHub CLI](https://cli.github.com/) and [golang](https://go.dev/doc/install).

### Local development
```bash
gh clone koestler/go-mqtt-to-influx
cd go-mqtt-to-influx
go build
./go-mqtt-to-influx
```

### Compile and run inside docker
```bash
gh clone koestler/go-mqtt-to-influx
cd go-mqtt-to-influx
docker build -f docker/Dockerfile -t go-mqtt-to-influx .
docker run --rm --name go-mqtt-to-influx -p 127.0.0.1:8000:8000 \
  -v "$(pwd)"/documentation/config.yaml:/app/config.yaml:ro \
  go-mqtt-to-influx
```

### Run tests
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