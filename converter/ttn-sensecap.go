package converter

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

// Parse payload of SenseCAP S2120 8-in-1 LoRaWAN Weather Sensor
// Code is compatible with TTN decoder function shown here:
// https://github.com/Seeed-Solution/TTN-Payload-Decoder/blob/master/SenseCAP_S2120_Weather_Station_Decoder.js

type ttnSensecapMessage struct {
	ReceivedAt    time.Time `json:"received_at"`
	UplinkMessage struct {
		DecodedPayload struct {
			Err      int `json:"err"`
			Messages []struct {
				MeasurementId    string  `json:"measurementId"`
				MeasurementValue float64 `json:"measurementValue"`
				Type             string  `json:"type"`
			} `json:"messages"`
			Valid bool `json:"valid"`
		} `json:"decoded_payload"`
	} `json:"uplink_message"`
}

func init() {
	registerHandler("ttn-sensecap-s2120", ttnSensecapS2120Handler)
}

// example messages:
//         "messages": [
//          {
//            "measurementId": "4104",
//            "measurementValue": 228,
//            "type": "Wind Direction Sensor"
//          },
//          {
//            "measurementId": "4105",
//            "measurementValue": 0,
//            "type": "Wind Speed"
//          },

//          {
//            "measurementId": "4097",
//            "measurementValue": 8.9,
//            "type": "Air Temperature"
//          },
//          {
//            "measurementId": "4098",
//            "measurementValue": 57,
//            "type": "Air Humidity"
//          },
//          {
//            "measurementId": "4099",
//            "measurementValue": 1651,
//            "type": "Light Intensity"
//          },
//          {
//            "measurementId": "4190",
//            "measurementValue": 0,
//            "type": "UV Index"
//          },
//          {
//            "measurementId": "4113",
//            "measurementValue": 0,
//            "type": "Rain Gauge"
//          },
//          {
//            "measurementId": "4101",
//            "measurementValue": 96950,
//            "type": "Barometric Pressure"
//          }

func ttnSensecapHandler(c Config, device, model string, input Input, outputFunc OutputFunc) {
	// parse payload
	var message ttnSensecapMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("ttn-sensecap[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	// send points
	count := 0
	for _, m := range message.UplinkMessage.DecodedPayload.Messages {
		count += 1
		outputFunc(telemetryOutputMessage{
			timeStamp: message.ReceivedAt,
			device:    device,
			field:     m.Type,
			unit:,
			sensor:     model,
			floatValue: &m.MeasurementValue,
		})
	}

	// any points sent?
	if count < 1 && c.LogDebug() {
		log.Printf("ttn-sensecap[%s]: no measurement saved; payload='%s'",
			c.Name(), input.Payload(),
		)
		return
	}
}

func ttnSensecapField(t string) string {
	return strings.ReplaceAll(t, " ", "")
}

func ttnSenscapeUnit(t string) string {
	switch t {
	case "Air Temperature":
		return "°C"
	case "Air Humidity":
		return "%"
	case "Light Intensity":
		return "lux"
	case "Barometric Pressure":
		return "hPa"
	case "Wind Speed":
		return "m/s"
	case "WindDirection":
		return "°"
	case "Rainfall":
		return "mm"
	case "Battery":
		return "V"
	default:
		return ""
	}
}()

func senscapMeasurementIdDecoder(id uint) (string, string) {
	switch id {
	case 4097:
		return "Air Temperature", "°C"
	case 4098:
		return "Air Humidity", "%"
	case 4099:
		return "Light Intensity", "lux"
	case 4100:
		return "CO2", "ppm"
	case 4101:
		return "Barometric Pressure", "hPa"
	case 4102:
		return "Soil Temperature", "°C"
	case 4103:
		return "Soil Moisture", "%"
	case 4104:
		return "Wind Direction", "°"
	case 4105:
		return "Wind Speed", "m/s"
	case 4106:
		return "pH", "pH"
	case 4107:
		return "Light Quantum", "umol/㎡s"
	case 4108:
		return "Electrical Conductivity", "mS/cm"
	case 4109:
		return "Dissolved Oxygen", "mg/L"
	case 4110:
		return "Soil Volumetric Water Content", "%"
	case 4111:
		return "Soil Electrical Conductivity", "mS/cm"
	case 4112:
		return "Soil Temperature(Soil Temperature, VWC & EC Sensor)", "°C"
	case 4113:
		return "Rainfall Hourly", "mm/hour"
	case 4115:
		return "Distance", "cm"
	case 4116:
		return "Water Leak", ""
	case 4117:
		return "Liguid Level", "cm"
	case 4118:
		return "NH3", "ppm"
	case 4119:
		return "H2S", "ppm"
	case 4120:
		return "Flow Rate", "m3/h"
	case 4121:
		return "Total Flow", "m3"
	case 4122:
		return "Oxygen Concentration", "%vol"
	case 4123:
		return "Water Eletrical Conductivity", "us/cm"
	case 4124:
		return "Water Temperature", "°C"
	case 4125:
		return "Soil Heat Flux", "W/㎡"
	case 4126:
		return "Sunshine Duration", "h"
	case 4127:
		return "Total Solar Radiation", "W/㎡
	case 4128:
		return "Water Surface Evaporation", "mm"
	case 4129:
		return "Photosynthetically Active Radiation(PAR)", "umol/㎡s"
	case 4130:
		return "Accelerometer", "m/s"
	case 4131:
		return "Sound Intensity", ""
	case 4133:
		return "Soil Tension", "KPA"
	case 4134:
		return "Salinity", "mg/L"
	case 4135:
		return "TDS", "mg/L"
	case 4136:
		return "Leaf Temperature", "°C"
	case 4137:
		return "Leaf Wetness", "%"
	case 4138:
		return "Soil Moisture-10cm", "%"
	case 4139:
		return "Soil Moisture-20cm", "%"
	case 4140:
		return "Soil Moisture-30cm", "%"
	case 4141:
		return "Soil Moisture-40cm", "%"
	case 4142:
		return "Soil Temperature-10cm", "°C"
	case 4143:
		return "Soil Temperature-20cm", "°C"
	case 4144:
		return "Soil Temperature-30cm", "°C"
	case 4145:
		return "Soil Temperature-40cm", "°C"
	case 4146:
		return "PM2.5", "μg/m3"
	case 4147:
		return "PM10", "μg/m3"
	case 4148:
		return "Noise", "dB"
	case 4150:
		return "AccelerometerX", "m/s²"
	case 4151:
		return "AccelerometerY", "m/s²"
	case 4152:
		return "AccelerometerZ", "m/s²"
	case 4154:
		return "Salinity", "PSU"
	case 4155:
		return "ORP", "mV"
	case 4156:
		return "Turbidity", "NTU"
	case 4157:
		return "Ammonia ion", "mg/L"
	case 4158:
		return "Eletrical Conductivity", "mS/cm"
	case 4159:
		return "Eletrical Conductivity", "mS/cm"
	case 4160:
		return "Eletrical Conductivity", "mS/cm"
	case 4161:
		return "Eletrical Conductivity", "mS/cm"
	case 4162:
		return "N Content", "mg/kg"
	case 4163:
		return "P Content", "mg/kg"
	case 4164:
		return "K Content", "mg/kg"
	case 4165:
		return "Measurement1", ""
	case 4166:
		return "Measurement2", ""
	case 4167:
		return "Measurement3", ""
	case 4168:
		return "Measurement4", ""
	case 4169:
		return "Measurement5", ""
	case 4170:
		return "Measurement6", ""
	case 4171:
		return "Measurement7", ""
	case 4172:
		return "Measurement8", ""
	case 4173:
		return "Measurement9", ""
	case 4174:
		return "Measurement10", ""
	case 4175:
		return "AI Detection No.01", ""
	case 4176:
		return "AI Detection No.02", ""
	case 4177:
		return "AI Detection No.03", ""
	case 4178:
		return "AI Detection No.04", ""
	case 4179:
		return "AI Detection No.05", ""
	case 4180:
		return "AI Detection No.06", ""
	case 4181:
		return "AI Detection No.07", ""
	case 4182:
		return "AI Detection No.08", ""
	case 4183:
		return "AI Detection No.09", ""
	case 4184:
		return "AI Detection No.10", ""
	case 4190:
		return "UV Index", ""
	case 4191:
		return "Peak Wind Gust", "m/s"
	case 4192:
		return "Sound Intensity", "dB"
	case 4193:
		return "Light Intensity", ""
	case 4195:
		return "TVOC", "ppb"
	case 4196:
		return "Soil moisture intensity", ""
	case 4197:
		return "longitude", "°"
	case 4198:
		return "latitude", "°"
	case 4199:
		return "Light", "%"
	case 4200:
		return "SOS Event", ""
	case 4201:
		return "Ultraviolet Radiation", "W/㎡"
	case 4202:
		return "Dew point temperature", "°C"
	case 4203:
		return "Temperature", "°C"
	case 4204:
		return "Soil Pore Water Eletrical Conductivity", "mS/cm"
	case 4205:
		return "Epsilon", ""
	case 4206:
		return "VOC_INDEX", ""
	case 4207:
		return "Noise", ""
	case 4208:
		return "Custom event", ""
	case 4209:
		return "Motion Id", ""
	case 5001:
		return "Wi-Fi MAC Address", ""
	case 5002:
		return "Bluetooth Beacon MAC Address", ""
	case 5003:
		return "Event list", ""
	case 5100:
		return "Switch", ""
	default:
		return "", ""
	}
}
