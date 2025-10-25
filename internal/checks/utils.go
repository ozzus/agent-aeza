package checks

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func stringParam(params map[string]interface{}, key, fallback string) string {
	if params == nil {
		return fallback
	}

	if value, ok := params[key]; ok {
		switch v := value.(type) {
		case string:
			if v == "" {
				return fallback
			}
			return v
		case fmt.Stringer:
			return v.String()
		default:
			str := fmt.Sprintf("%v", value)
			if str == "" {
				return fallback
			}
			return str
		}
	}

	return fallback
}

func lowerStringParam(params map[string]interface{}, key, fallback string) string {
	return strings.ToLower(stringParam(params, key, fallback))
}

func intParam(params map[string]interface{}, key string, fallback int) int {
	if params == nil {
		return fallback
	}

	if value, ok := params[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float32:
			return int(v)
		case float64:
			return int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				return parsed
			}
		}
	}

	return fallback
}

func durationParam(params map[string]interface{}, key string, fallback time.Duration) time.Duration {
	if params == nil {
		return fallback
	}

	if value, ok := params[key]; ok {
		switch v := value.(type) {
		case time.Duration:
			return v
		case int:
			return time.Duration(v) * time.Millisecond
		case int64:
			return time.Duration(v) * time.Millisecond
		case float64:
			return time.Duration(v) * time.Millisecond
		case string:
			if parsed, err := time.ParseDuration(v); err == nil {
				return parsed
			}
		}
	}

	return fallback
}

func formatSeconds(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	return fmt.Sprintf("%.3f s", d.Seconds())
}

func formatMilliseconds(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	return fmt.Sprintf("%.1f ms", float64(d.Microseconds())/1000.0)
}

func formatTTL(d time.Duration) string {
	if d <= 0 {
		return "N/A"
	}

	totalSeconds := int(d.Round(time.Second).Seconds())
	if totalSeconds < 1 {
		return "<1 sec"
	}

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	parts := make([]string, 0, 3)
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hr", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d min", minutes))
	}
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%d sec", seconds))
	}

	return strings.Join(parts, " ")
}
