package util

import (
	"encoding/json"
	"fmt"
	"time"
)

func ParseArg(args json.RawMessage, key string) string {
	var m map[string]interface{}
	json.Unmarshal(args, &m)
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func ParseArgOpt(args json.RawMessage, key string) *string {
	var m map[string]interface{}
	json.Unmarshal(args, &m)
	if v, ok := m[key].(string); ok && v != "" {
		return &v
	}
	return nil
}

func TimeString(t time.Time) string {
	h := t.Hour()
	m := t.Minute()
	period := "AM"
	if h >= 12 {
		period = "PM"
	}
	if h > 12 {
		h -= 12
	}
	if h == 0 {
		h = 12
	}
	return fmt.Sprintf("%d:%02d %s", h, m, period)
}

func MustParseTime(s string) time.Time {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return time.Now()
	}
	return t
}

func JoinStr(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
