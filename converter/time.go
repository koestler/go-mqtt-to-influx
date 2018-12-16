package converter

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"
)

const timeFormat string = "2006-01-02T15:04:05"

func parseTime(timeStr string) (res time.Time, err error) {
	res, err = time.Parse(timeFormat, timeStr)
	if err != nil {
		log.Printf("time: cannot parse time string: '%s'", timeStr)
	}
	return
}

// example: 8T04:17:27
var uptimeParser = regexp.MustCompile("^([0-9]+)T([0-9]{1,2}):([0-9]{1,2}):([0-9]{1,2})$")

// returns uptime in seconds
func parseUpTime(timeStr string) (res int, err error) {
	parts := uptimeParser.FindStringSubmatch(timeStr)

	if len(parts) != 5 {
		return 0, fmt.Errorf("converter: cannot parse uptime str='%s'", timeStr)
	}

	// Atoi won't fail due to regexp above
	days, _ := strconv.Atoi(parts[1])
	hours, _ := strconv.Atoi(parts[2])
	minutes, _ := strconv.Atoi(parts[3])
	seconds, _ := strconv.Atoi(parts[4])

	return 24*60*60*days + 60*60*hours + 60*minutes + seconds, nil;
}
