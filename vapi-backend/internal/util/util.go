package util

import (
	"encoding/json"
	"fmt"
	"time"
)

// parseArgsMap handles both a JSON object and a JSON-encoded string (Vapi/OpenAI format).
func parseArgsMap(args json.RawMessage) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal(args, &m); err == nil {
		return m
	}
	// Vapi sends arguments as a JSON string — unwrap and try again
	var s string
	if err := json.Unmarshal(args, &s); err == nil {
		json.Unmarshal([]byte(s), &m)
	}
	return m
}

func ParseArg(args json.RawMessage, key string) string {
	m := parseArgsMap(args)
	if v, ok := m[key].(string); ok {
		return v
	}
	// Also try snake_case <-> camelCase conversion
	altKey := camelToSnake(key)
	if altKey != key {
		if v, ok := m[altKey].(string); ok {
			return v
		}
	}
	altKey2 := snakeToCamel(key)
	if altKey2 != key {
		if v, ok := m[altKey2].(string); ok {
			return v
		}
	}
	return ""
}

func ParseArgOpt(args json.RawMessage, key string) *string {
	m := parseArgsMap(args)
	if v, ok := m[key].(string); ok && v != "" {
		return &v
	}
	// Also try snake_case <-> camelCase conversion
	altKey := camelToSnake(key)
	if altKey != key {
		if v, ok := m[altKey].(string); ok && v != "" {
			return &v
		}
	}
	altKey2 := snakeToCamel(key)
	if altKey2 != key {
		if v, ok := m[altKey2].(string); ok && v != "" {
			return &v
		}
	}
	return nil
}

// camelToSnake converts camelCase to snake_case.
func camelToSnake(s string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, c+'a'-'A')
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}

// snakeToCamel converts snake_case to camelCase.
func snakeToCamel(s string) string {
	var result []byte
	upperNext := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			upperNext = true
		} else if upperNext {
			if c >= 'a' && c <= 'z' {
				result = append(result, c-'a'+'A')
			} else {
				result = append(result, c)
			}
			upperNext = false
		} else {
			result = append(result, c)
		}
	}
	return string(result)
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
	// Handle both HH:MM (slot format) and HH:MM:SS (PostgreSQL TIME::text)
	if len(s) > 5 {
		s = s[:5]
	}
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
