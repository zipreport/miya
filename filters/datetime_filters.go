package filters

import (
	"fmt"
	"strings"
	"time"
)

// DateFilter formats a date/time value
func DateFilter(value interface{}, args ...interface{}) (interface{}, error) {
	format := "2006-01-02" // Default format
	if len(args) > 0 {
		format = ToString(args[0])
	}

	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("date filter requires a date/time value: %v", err)
	}

	return t.Format(convertJinjaTimeFormat(format)), nil
}

// TimeFilter formats a time value
func TimeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	format := "15:04:05" // Default format
	if len(args) > 0 {
		format = ToString(args[0])
	}

	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("time filter requires a date/time value: %v", err)
	}

	return t.Format(convertJinjaTimeFormat(format)), nil
}

// DatetimeFilter formats a datetime value
func DatetimeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	format := "2006-01-02 15:04:05" // Default format
	if len(args) > 0 {
		format = ToString(args[0])
	}

	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("datetime filter requires a date/time value: %v", err)
	}

	return t.Format(convertJinjaTimeFormat(format)), nil
}

// StrftimeFilter formats time using strftime-like format
func StrftimeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("strftime filter requires format argument")
	}

	format := ToString(args[0])
	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("strftime filter requires a date/time value: %v", err)
	}

	return t.Format(convertStrftimeFormat(format)), nil
}

// TimestampFilter returns Unix timestamp
func TimestampFilter(value interface{}, args ...interface{}) (interface{}, error) {
	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("timestamp filter requires a date/time value: %v", err)
	}

	return t.Unix(), nil
}

// AgeFilter calculates age in years from a date
func AgeFilter(value interface{}, args ...interface{}) (interface{}, error) {
	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("age filter requires a date/time value: %v", err)
	}

	now := time.Now()
	age := now.Year() - t.Year()

	// Adjust if birthday hasn't occurred this year
	if now.YearDay() < t.YearDay() {
		age--
	}

	return age, nil
}

// RelativeDateFilter returns human-readable relative date
func RelativeDateFilter(value interface{}, args ...interface{}) (interface{}, error) {
	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("relative_date filter requires a date/time value: %v", err)
	}

	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		diff = -diff
		return formatRelativeTime(diff, true), nil
	}

	return formatRelativeTime(diff, false), nil
}

// WeekdayFilter returns the weekday name
func WeekdayFilter(value interface{}, args ...interface{}) (interface{}, error) {
	abbreviated := false
	if len(args) > 0 {
		abbreviated = ToBool(args[0])
	}

	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("weekday filter requires a date/time value: %v", err)
	}

	if abbreviated {
		return t.Format("Mon"), nil
	}
	return t.Format("Monday"), nil
}

// MonthNameFilter returns the month name
func MonthNameFilter(value interface{}, args ...interface{}) (interface{}, error) {
	abbreviated := false
	if len(args) > 0 {
		abbreviated = ToBool(args[0])
	}

	t, err := parseTimeValue(value)
	if err != nil {
		return nil, fmt.Errorf("month_name filter requires a date/time value: %v", err)
	}

	if abbreviated {
		return t.Format("Jan"), nil
	}
	return t.Format("January"), nil
}

// Helper functions

func parseTimeValue(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v != nil {
			return *v, nil
		}
		return time.Time{}, fmt.Errorf("nil time pointer")
	case int64:
		return time.Unix(v, 0), nil
	case int:
		return time.Unix(int64(v), 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	case string:
		// Try common formats
		formats := []string{
			time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
			time.RFC822,           // "02 Jan 06 15:04 MST"
			"2006-01-02",          // Date only
			"15:04:05",            // Time only
			"2006-01-02 15:04:05", // DateTime
			"01/02/2006",          // US format
			"02/01/2006",          // European format
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse time string: %s", v)
	default:
		return time.Time{}, fmt.Errorf("unsupported type for time value: %T", value)
	}
}

func convertJinjaTimeFormat(jinjaFormat string) string {
	// Check if this looks like a strftime format (contains % symbols)
	// If so, convert it using the strftime converter
	if strings.Contains(jinjaFormat, "%") {
		return convertStrftimeFormat(jinjaFormat)
	}
	// Otherwise, assume it's already a Go time format
	return jinjaFormat
}

func convertStrftimeFormat(strftimeFormat string) string {
	// Convert common strftime formats to Go time format
	replacements := map[string]string{
		"%Y": "2006",    // 4-digit year
		"%y": "06",      // 2-digit year
		"%m": "01",      // month (01-12)
		"%d": "02",      // day (01-31)
		"%H": "15",      // hour (00-23)
		"%I": "03",      // hour (01-12)
		"%M": "04",      // minute (00-59)
		"%S": "05",      // second (00-59)
		"%p": "PM",      // AM/PM
		"%B": "January", // full month name
		"%b": "Jan",     // abbreviated month name
		"%A": "Monday",  // full weekday name
		"%a": "Mon",     // abbreviated weekday name
		"%Z": "MST",     // timezone abbreviation
	}

	result := strftimeFormat
	for strftime, golang := range replacements {
		// Simple string replacement - a more robust implementation would use regex
		for i := 0; i < len(result); i++ {
			if i+len(strftime) <= len(result) && result[i:i+len(strftime)] == strftime {
				result = result[:i] + golang + result[i+len(strftime):]
				i += len(golang) - 1
			}
		}
	}

	return result
}

func formatRelativeTime(duration time.Duration, future bool) string {
	seconds := int(duration.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	var result string
	var unit string

	switch {
	case years > 0:
		result = fmt.Sprintf("%d", years)
		unit = "year"
		if years > 1 {
			unit += "s"
		}
	case months > 0:
		result = fmt.Sprintf("%d", months)
		unit = "month"
		if months > 1 {
			unit += "s"
		}
	case weeks > 0:
		result = fmt.Sprintf("%d", weeks)
		unit = "week"
		if weeks > 1 {
			unit += "s"
		}
	case days > 0:
		result = fmt.Sprintf("%d", days)
		unit = "day"
		if days > 1 {
			unit += "s"
		}
	case hours > 0:
		result = fmt.Sprintf("%d", hours)
		unit = "hour"
		if hours > 1 {
			unit += "s"
		}
	case minutes > 0:
		result = fmt.Sprintf("%d", minutes)
		unit = "minute"
		if minutes > 1 {
			unit += "s"
		}
	default:
		result = fmt.Sprintf("%d", seconds)
		unit = "second"
		if seconds != 1 {
			unit += "s"
		}
	}

	if future {
		return fmt.Sprintf("in %s %s", result, unit)
	}
	return fmt.Sprintf("%s %s ago", result, unit)
}
