package converter

import (
	"testing"
)

func TestParseTime(t *testing.T) {
	values := []struct {
		Str      string
		UnixTime int64
	}{
		{"1970-01-01T00:00:00", 0},
		{"2018-12-19T00:31:05", 1545179465},
	}

	t.Logf("use test values: %v", values)

	for _, v := range values {
		res, err := parseTime(v.Str)
		if err != nil {
			t.Errorf("did not expect an error: %s", err)
			continue
		}
		if res.Unix() != v.UnixTime {
			t.Errorf("expected str='%s' to have Unix timestamp %d", v.Str, v.UnixTime)
		}
	}
}

func TestParseInvalidTime(t *testing.T) {
	_, err := parseTime("2018-12-19q00:31:05")
	if err == nil {
		t.Error("expected error when entering an invalid timestamp")
	}
}

func TestParseUpTime(t *testing.T) {
	values := []struct {
		Str     string
		Seconds int
	}{
		{"10T05:40:59", 884459},
		{"3T14:06:42", 310002},
		{"0T00:00:15", 15},
		{"00:00:15", 15},
	}

	t.Logf("use test values: %v", values)

	for _, v := range values {
		res, err := parseUpTime(v.Str)
		if err != nil {
			t.Errorf("did not expect an error: %s", err)
			continue
		}
		if res != v.Seconds {
			t.Errorf("expected str='%s' to return %d seconds", v.Str, v.Seconds)
		}
	}
}

func TestParseInvalidUpTime(t *testing.T) {
	_, err := parseUpTime("10T05:40:-59")
	if err == nil {
		t.Error("expected error when entering an invalid timestamp")
	}
}

func TestParseTimeWithZone(t *testing.T) {
	values := []struct {
		Str      string
		UnixTime int64
	}{
		{"1970-01-01T00:00:00Z", 0},
		{"2018-12-19T00:31:05Z", 1545179465},
	}

	t.Logf("use test values: %v", values)

	for _, v := range values {
		res, err := parseTimeWithZone(v.Str)
		if err != nil {
			t.Errorf("did not expect an error: %s", err)
			continue
		}
		if res.Unix() != v.UnixTime {
			t.Errorf("expected str='%s' to have Unix timestamp %d", v.Str, v.UnixTime)
		}
	}
}

func TestParseInvalidTimeWithZone(t *testing.T) {
	_, err := parseTimeWithZone("1970-01-01T00:00:00")
	if err == nil {
		t.Error("expected error when entering an invalid timestamp")
	}
}
