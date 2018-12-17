package influxDbClient

import "time"

type RawPoint struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Time        time.Time
}

type Point struct {
	Tags   map[string]string
	Fields map[string]interface{}
}

type Points []Point

func (points Points) ToRaw(measurement string, time time.Time) (rawPoints []RawPoint) {
	rawPoints = make([]RawPoint, len(points))
	for i, point := range points {
		rawPoints[i] = RawPoint{
			Measurement: measurement,
			Tags:        point.Tags,
			Fields:      point.Fields,
			Time:        time,
		}
	}
	return
}
