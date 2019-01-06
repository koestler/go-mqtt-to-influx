package converter

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"regexp"
	"strconv"
	"time"
)

const (
	timeFormat   string = "2006-01-02T15:04:05"
	uptimeFormat string = "^(([0-9]+)T)?([0-9]{1,2}):([0-9]{1,2}):([0-9]{1,2})$"
)

func parseTime(timeStr string) (res time.Time, err error) {
	if len(timeStr) < 1 {
		return res, errors.New("empty timeStr")
	}

	res, err = time.Parse(timeFormat, timeStr)
	if err != nil {
		log.Printf("time: cannot parse timeString='%s': %s : expect format %s", timeStr, err, timeFormat)
	}
	return
}

// example: 3T14:06:42
var uptimeParser = regexp.MustCompile(uptimeFormat)

// returns uptime in seconds
func parseUpTime(timeStr string) (res int, err error) {
	parts := uptimeParser.FindStringSubmatch(timeStr)

	var days, hours, minutes, seconds int

	if len(parts) == 6 {
		// Atoi won't fail due to regexp above
		days, _ = strconv.Atoi(parts[2])
		hours, _ = strconv.Atoi(parts[3])
		minutes, _ = strconv.Atoi(parts[4])
		seconds, _ = strconv.Atoi(parts[5])
	} else if len(parts) == 4 {
		days = 0
		// Atoi won't fail due to regexp above
		hours, _ = strconv.Atoi(parts[2])
		minutes, _ = strconv.Atoi(parts[3])
		seconds, _ = strconv.Atoi(parts[4])
	} else {
		return 0, fmt.Errorf("converter: cannot parse uptime str='%s' : expect format %s",
			timeStr, uptimeFormat,
		)
	}

	return 24*60*60*days + 60*60*hours + 60*minutes + seconds, nil
}
