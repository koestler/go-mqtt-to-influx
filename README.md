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
# documentation/minimal-config.yaml

Version: 0                                                 # mandatory, version is always 0 (reserved for later use)

HttpServer:                                                # optional, default Disabled, start the http server
  Port: 8042                                               # optional, default 8042
Statistics:                                                # optional, default Disabled, start the statistics module (needs some additional memory/cpu)
  Enabled: True                                            # mandatory, set to True to enable

MqttClients:                                               # mandatory, a list of MQTT servers to connect to
  example:                                                 # mandatory, an arbitrary name used in log outputs and for reference in the converters section
    Broker: "tcp://mqtt.exampel.com:1883"                  # mandatory, the address / port of the server
    User: Bob                                              # optional, if given used for login
    Password: Jeir2Jie4zee                                 # optional, if given used for login

InfluxClients:                                             # mandatory, a list of Influx DB Client configuration
  example:                                                 # mandatory, an arbitrary name used in log outputs and for reference in the converters section
    Address: "http://influx.example.com:8086"              # mandatory, the address / port of the server
    User: Alice                                            # optional, if given used for login
    Password: An2iu2egheijeG                               # optional, if given used for login

Converters:                                                # mandatory, a list of Converters to run
  lwt:                                                     # mandatory, an arbitrary name used in log outputs
    Implementation: lwt                                    # mandatory, which converter to use
    MqttTopics:                                            # mandatory, a list of topics to subscribe to
      - tele/+/+/LWT

  tasmota-state:
    Implementation: tasmota-state
    MqttTopics:
      - tele/+/+/STATE

  tasmota-sensor:
    Implementation: tasmota-sensor
    MqttTopics:
      - tele/+/+/SENSOR
```  

### Complete, explained example

```yaml
# documentation/config.yaml

Version: 0                                                 # mandatory, version is always 0 (reserved for later use)
LogConfig: True                                            # optional, default False, outputs the configuration including defaults on startup
LogWorkerStart: True                                       # optional, default False, write log for starting / stoping of workers
LogMqttDebug: False                                        # optional, default False, enable debug output of the mqtt module
HttpServer:                                                # optional, default Disabled, start the http server
  Bind: 0.0.0.0                                            # optional, default ::1 (ipv6 loopback)
  Port: 80                                                 # optional, default 8042
  LogRequests: True                                        # optional, default False, log all requests to stdout
Statistics:                                                # optional, default Disabled, start the statistics module (needs some additional memory/cpu)
  Enabled: True                                            # mandatory, set to True to enable
  HistoryResolution: 1s                                    # optional, default 1s, time resolution for aggregation, decrease with caution
  HistoryMaxAge: 10m                                       # optional, default 10min, how many time steps to keep, increase with caution

MqttClients:                                               # mandatory, a list of MQTT servers to connect to
  0-piegn-mosquitto:                                       # mandatory, an arbitrary name used in log outputs and for reference in the converters section
    Broker: "tcp://mqtt.exampel.com:1883"                  # mandatory, the address / port of the server
    User: Bob                                              # optional, if given used for login
    Password: Jeir2Jie4zee                                 # optional, if given used for login
    ClientId: "config-tester"                              # optional, default go-mqtt-to-influx, client-id sent to the server
    Qos: 2                                                 # optional, default 0, QOS-level used for subscriptions
    AvailabilityTopic: test/%Prefix%tele/%ClientId%/LWT    # optional, if given, a message with Online/Offline will be published on connect/disconnect
                                                           # supported placeholders:
                                                           # - %Prefix$   : as specified in this config section
                                                           # - %ClientId% : as specified in this config section
    TopicPrefix: piegn/                                    # optional, default empty
    LogMessages: False                                     # optional, default False, logs all received messages

  1-local-mosquitto:                                       # optional, a second MQTT erver
    Broker: "tcp://172.17.0.5:1883"                        # optional, the second MQTT servers broker...

InfluxClients:                                             # mandatory, a list of Influx DB Client configuration
  0-piegn:                                                 # mandatory, an arbitrary name used in log outputs and for reference in the converters section
    Address: "http://influx.example.com:8086"              # mandatory, the address / port of the server
    User: Alice                                            # optional, if given used for login
    Password: An2iu2egheijeG                               # optional, if given used for login
    Database: test-database                                # optional, default go-mqtt-to-influx, the db to be used
    WriteInterval: 400ms                                   # optional, default 200ms, how often to write a batch of points, if 0, each point is sent immediately in a separate request
    TimePrecision: 1ms                                     # optional, default 1s, influx db time precision
    LogLineProtocol: True                                  # optional, default False, outputs the Influx Line Protocol of each point

  1-local:                                                 # optional, a second Influx Server
    Address: "http://[::1]:8086"                           # optional, the address of the second Influx server ...

Converters:                                                # mandatory, a list of Converters to run
  0-lwt:                                                   # mandatory, an arbitrary name used in log outputs
    Implementation: lwt                                    # mandatory, which converter to use
    TargetMeasurement: boolValue                           # optional, default depending on implementation
    MqttTopics:                                            # mandatory, a list of topics to subscribe to
      - %Prefix%tele/+/+/LWT
    MqttClients:                                           # optional, default all configured, a list of MQTT Clients to receive data from
      - 0-piegn-mosquitto
      - 1-local-mosquitto
    InfluxClients:                                         # optional, default all configured, a list of Influx Clients to send data to
      - 0-piegn
      - 1-local
    LogHandleOnce: True                                    # optional, default False, if True each topic is logged once by each converter

  1-ve:                                                    # optional, a second Converter
    Implementation: go-iotdevice
    MqttTopics:
      - %Prefix%tele/ve/#
    LogHandleOnce: True

  2-tasmota-state:
    Implementation: tasmota-state
    MqttTopics:
      - %Prefix%tele/+/+/STATE
    LogHandleOnce: True

  3-tasmota-sensor:
    Implementation: tasmota-sensor
    MqttTopics:
      - %Prefix%tele/+/+/SENSOR
    LogHandleOnce: True
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

## Development

### Compile and run inside docker
```bash
docker build -f docker/Dockerfile -t go-mqtt-to-influx .
docker run --rm --name go-mqtt-to-influx -p 127.0.0.1:8042:8042 \
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

# License
MIT License