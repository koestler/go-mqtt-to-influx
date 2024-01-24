package converter

import (
	"encoding/json"
	"log"
	"strconv"
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
			} `json:"messages"`
			Valid bool `json:"valid"`
		} `json:"decoded_payload"`
	} `json:"uplink_message"`
}

func ttnSensecapHandler(c Config, device, model string, input Input, outputFunc OutputFunc) {
	// parse payload
	var message ttnSensecapMessage
	payload := input.Payload()
	if err := json.Unmarshal(payload, &message); err != nil {
		log.Printf("ttn-sensecap[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	// send points
	count := 0
	for _, m := range message.UplinkMessage.DecodedPayload.Messages {
		mId, err := strconv.Atoi(m.MeasurementId)
		if err != nil {
			log.Printf("ttn-sensecap[%s]: invalid MeasurementId='%s'", c.Name(), m.MeasurementId)
			continue
		}
		name, _, unit := senscapMeasurementDecoder(mId)

		count += 1
		outputFunc(telemetryOutputMessage{
			timeStamp:  message.ReceivedAt,
			device:     device,
			field:      name,
			unit:       &unit,
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

// List from https://sensecap-docs.seeed.cc/measurement_list.html
func senscapMeasurementDecoder(measurementId int) (name, valueRange, unit string) {
	switch measurementId {
	case 4097:
		return "Air Temperature", "-40~90", "°C"
	case 4098:
		return "Air Humidity", "0~100", "% RH"
	case 4099:
		return "Light Intensity", "0~188000", "Lux"
	case 4100:
		return "CO2", "0~10000", "ppm"
	case 4101:
		return "Barometric Pressure", "300~1100000", "Pa"
	case 4102:
		return "Soil Temperature", "-30~70", "°C"
	case 4103:
		return "Soil Moisture", "0~100", "%"
	case 4104:
		return "Wind Direction", "0~360", "°"
	case 4105:
		return "Wind Speed", "0~60", "m/s"
	case 4106:
		return "pH", "0~14", "PH"
	case 4107:
		return "Light Quantum", "0~5000", "umol/㎡s"
	case 4108:
		return "Electrical Conductivity", "0~23", "mS/cm"
	case 4109:
		return "Dissolved Oxygen", "0~20", "mg/L"
	case 4110:
		return "Soil Volumetric Water Content", "0~100", "%"
	case 4111:
		return "Soil Electrical Conductivity", "0~23", "mS/cm"
	case 4112:
		return "Soil Temperature(Soil Temperature, VWC & EC Sensor)", "-40~60", "°C"
	case 4113:
		return "Rainfall Hourly", "0~240", "mm/hour"
	case 4115:
		return "Distance", "28~250", "cm"
	case 4116:
		return "Water Leak", "true / false", ""
	case 4117:
		return "Liguid Level", "0~500", "cm"
	case 4118:
		return "NH3", "0~100", "ppm"
	case 4119:
		return "H2S", "0~100", "ppm"
	case 4120:
		return "Flow Rate", "0~65535", "m3/h"
	case 4121:
		return "Total Flow", "0~6553599", "m3"
	case 4122:
		return "Oxygen Concentration", "0~25", "%vol"
	case 4123:
		return "Water Eletrical Conductivity", "0~20000", "us/cm"
	case 4124:
		return "Water Temperature", "-40~80", "°C"
	case 4125:
		return "Soil Heat Flux", "-500~500", "W/㎡"
	case 4126:
		return "Sunshine Duration", "0~10000", "h"
	case 4127:
		return "Total Solar Radiation", "0~5000", "W/㎡"
	case 4128:
		return "Water Surface Evaporation", "0~10000", "mm"
	case 4129:
		return "Photosynthetically Active Radiation(PAR)", "0～5000", "umol/㎡s"
	case 4130:
		return "Accelerometer", "0,0,0~x.xx,y.yy,z.zz", "m/s"
	case 4131:
		return "Sound Intensity", "0~100", ""
	case 4133:
		return "Soil Tension", "-100~0", "KPA"
	case 4134:
		return "Salinity", "0~20000", "mg/L"
	case 4135:
		return "TDS", "0~20000", "mg/L"
	case 4136:
		return "Leaf Temperature", "0~100", "°C"
	case 4137:
		return "Leaf Wetness", "-40~85", "%"
	case 4138:
		return "Soil Moisture-10cm", "0~100", "%"
	case 4139:
		return "Soil Moisture-20cm", "0~100", "%"
	case 4140:
		return "Soil Moisture-30cm", "0~100", "%"
	case 4141:
		return "Soil Moisture-40cm", "0~100", "%"
	case 4142:
		return "Soil Temperature-10cm", "-30~70", "°C"
	case 4143:
		return "Soil Temperature-20cm", "-30~70", "°C"
	case 4144:
		return "Soil Temperature-30cm", "-30~70", "°C"
	case 4145:
		return "Soil Temperature-40cm", "-30~70", "°C"
	case 4146:
		return "PM2.5", "0~1000", "μg/m3"
	case 4147:
		return "PM10", "0~2000", "μg/m3"
	case 4148:
		return "Noise", "30~130", "dB"
	case 4150:
		return "AccelerometerX", "-49.99~49.99", "m/s²"
	case 4151:
		return "AccelerometerY", "-49.99~49.99", "m/s²"
	case 4152:
		return "AccelerometerZ", "-49.99~49.99", "m/s²"
	case 4154:
		return "Salinity", "0~70", "PSU"
	case 4155:
		return "ORP", "-1500~-1500", "mV"
	case 4156:
		return "Turbidity", "0~1000", "NTU"
	case 4157:
		return "Ammonia ion", "0~100", "mg/L"
	case 4158:
		return "Eletrical Conductivity", "0~23", "mS/cm"
	case 4159:
		return "Eletrical Conductivity", "0~23", "mS/cm"
	case 4160:
		return "Eletrical Conductivity", "0~23", "mS/cm"
	case 4161:
		return "Eletrical Conductivity", "0~23", "mS/cm"
	case 4162:
		return "N Content", "0~1999", "mg/kg"
	case 4163:
		return "P Content", "0~1999", "mg/kg"
	case 4164:
		return "K Content", "0~1999", "mg/kg"
	case 4165:
		return "Measurement1", "~", ""
	case 4166:
		return "Measurement2", "~", ""
	case 4167:
		return "Measurement3", "~", ""
	case 4168:
		return "Measurement4", "~", ""
	case 4169:
		return "Measurement5", "~", ""
	case 4170:
		return "Measurement6", "~", ""
	case 4171:
		return "Measurement7", "~", ""
	case 4172:
		return "Measurement8", "~", ""
	case 4173:
		return "Measurement9", "~", ""
	case 4174:
		return "Measurement10", "~", ""
	case 4175:
		return "AI Detection No.01", "1.00~39.99", ""
	case 4176:
		return "AI Detection No.02", "1.00~39.99", ""
	case 4177:
		return "AI Detection No.03", "1.00~39.99", ""
	case 4178:
		return "AI Detection No.04", "1.00~39.99", ""
	case 4179:
		return "AI Detection No.05", "1.00~39.99", ""
	case 4180:
		return "AI Detection No.06", "1.00~39.99", ""
	case 4181:
		return "AI Detection No.07", "1.00~39.99", ""
	case 4182:
		return "AI Detection No.08", "1.00~39.99", ""
	case 4183:
		return "AI Detection No.09", "1.00~39.99", ""
	case 4184:
		return "AI Detection No.10", "1.00~39.99", ""
	case 4190:
		return "UV Index", "0~16.0", ""
	case 4191:
		return "Peak Wind Gust", "0~100", "m/s"
	case 4192:
		return "Sound Intensity", "0~1023", "dB"
	case 4193:
		return "Light Intensity", "0~1023", ""
	case 4195:
		return "TVOC", "0~60000", "ppb"
	case 4196:
		return "Soil moisture intensity", "0~1023", ""
	case 4197:
		return "longitude", "-180~180", "°"
	case 4198:
		return "latitude", "-90~90", "°"
	case 4199:
		return "Light", "0~100", "%"
	case 4200:
		return "SOS Event", "0~1", ""
	case 4201:
		return "Ultraviolet Radiation", "0~200", "W/㎡"
	case 4202:
		return "Dew point temperature", "0~50", "°C"
	case 4203:
		return "Temperature", "-40~150", "°C"
	case 4204:
		return "Soil Pore Water Eletrical Conductivity", "0~32", "mS/cm"
	case 4205:
		return "Epsilon", "0~100", ""
	case 4206:
		return "VOC_INDEX", "~", ""
	case 4207:
		return "Noise", "~", ""
	case 4208:
		return "Custom event", "~", ""
	case 4209:
		return "Motion Id", "0~255", ""
	case 5001:
		return "Wi-Fi MAC Address", "~", ""
	case 5002:
		return "Bluetooth Beacon MAC Address", "~", ""
	case 5003:
		return "Event list", "~", ""
	case 5100:
		return "Switch", "100~200", ""
	}
	return "", "", ""
}
