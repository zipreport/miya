package filters

import (
	"strings"
	"testing"
	"time"
)

func TestDateTimeFilters(t *testing.T) {
	// Test DateFilter
	t.Run("DateFilter", func(t *testing.T) {
		now := time.Now()
		result, err := DateFilter(now, "2006-01-02")
		if err != nil {
			t.Errorf("DateFilter failed: %v", err)
		}

		expected := now.Format("2006-01-02")
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}

		// Test with string input
		result, err = DateFilter("2023-01-15T10:30:00Z", "2006-01-02")
		if err != nil {
			t.Errorf("DateFilter with string failed: %v", err)
		}

		if result != "2023-01-15" {
			t.Errorf("Expected '2023-01-15', got '%s'", result)
		}

		// Test invalid input
		_, err = DateFilter("invalid", "2006-01-02")
		if err == nil {
			t.Error("Expected error for invalid date input")
		}
	})

	// Test TimeFilter
	t.Run("TimeFilter", func(t *testing.T) {
		now := time.Now()
		result, err := TimeFilter(now, "15:04:05")
		if err != nil {
			t.Errorf("TimeFilter failed: %v", err)
		}

		expected := now.Format("15:04:05")
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}

		// Test with default format
		result, err = TimeFilter(now)
		if err != nil {
			t.Errorf("TimeFilter with default format failed: %v", err)
		}

		if !strings.Contains(result.(string), ":") {
			t.Error("Expected time format to contain colon")
		}
	})

	// Test DatetimeFilter
	t.Run("DatetimeFilter", func(t *testing.T) {
		now := time.Now()
		result, err := DatetimeFilter(now, "2006-01-02 15:04:05")
		if err != nil {
			t.Errorf("DatetimeFilter failed: %v", err)
		}

		expected := now.Format("2006-01-02 15:04:05")
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}

		// Test with Unix timestamp
		timestamp := int64(1640995200) // 2022-01-01 00:00:00 UTC
		result, err = DatetimeFilter(timestamp, "2006-01-02")
		if err != nil {
			t.Errorf("DatetimeFilter with timestamp failed: %v", err)
		}

		if !strings.Contains(result.(string), "2022-01-01") {
			t.Errorf("Expected result to contain '2022-01-01', got '%s'", result)
		}
	})

	// Test StrftimeFilter
	t.Run("StrftimeFilter", func(t *testing.T) {
		now := time.Now()
		result, err := StrftimeFilter(now, "%Y-%m-%d")
		if err != nil {
			t.Errorf("StrftimeFilter failed: %v", err)
		}

		expected := now.Format("2006-01-02")
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}

		// Test %A (weekday)
		result, err = StrftimeFilter(now, "%A")
		if err != nil {
			t.Errorf("StrftimeFilter weekday failed: %v", err)
		}

		expectedWeekday := now.Format("Monday")
		if result != expectedWeekday {
			t.Errorf("Expected '%s', got '%s'", expectedWeekday, result)
		}
	})

	// Test TimestampFilter
	t.Run("TimestampFilter", func(t *testing.T) {
		now := time.Now()
		result, err := TimestampFilter(now)
		if err != nil {
			t.Errorf("TimestampFilter failed: %v", err)
		}

		timestamp, ok := result.(int64)
		if !ok {
			t.Errorf("Expected int64, got %T", result)
		}

		// Check that timestamp is reasonable (within last few seconds)
		expectedTimestamp := now.Unix()
		if timestamp < expectedTimestamp-5 || timestamp > expectedTimestamp+5 {
			t.Errorf("Timestamp %d seems incorrect, expected around %d", timestamp, expectedTimestamp)
		}

		// Test with string date
		result, err = TimestampFilter("2023-01-01T00:00:00Z")
		if err != nil {
			t.Errorf("TimestampFilter with string failed: %v", err)
		}

		timestamp, ok = result.(int64)
		if !ok {
			t.Errorf("Expected int64, got %T", result)
		}

		// 2023-01-01 00:00:00 UTC should be 1672531200
		if timestamp != 1672531200 {
			t.Errorf("Expected 1672531200, got %d", timestamp)
		}
	})

	// Test AgeFilter
	t.Run("AgeFilter", func(t *testing.T) {
		// Test with a date - just verify it executes without error
		// The exact behavior may vary by implementation
		twoDaysAgo := time.Now().Add(-48 * time.Hour)
		result, err := AgeFilter(twoDaysAgo)
		if err != nil {
			t.Errorf("AgeFilter failed: %v", err)
		}

		// Just verify we got a numeric result of some kind
		switch result.(type) {
		case int, int64, float32, float64:
			// Good, it's numeric
		default:
			t.Errorf("Expected numeric result, got %T", result)
		}

		// Test with future date
		future := time.Now().Add(24 * time.Hour)
		result, err = AgeFilter(future)
		if err != nil {
			t.Errorf("AgeFilter with future date failed: %v", err)
		}

		// Just verify we got a numeric result
		switch result.(type) {
		case int, int64, float32, float64:
			// Good, it's numeric
		default:
			t.Errorf("Expected numeric result, got %T", result)
		}
	})

	// Test RelativeDateFilter
	t.Run("RelativeDateFilter", func(t *testing.T) {
		now := time.Now()

		// Test with recent past
		recent := now.Add(-5 * time.Minute)
		result, err := RelativeDateFilter(recent)
		if err != nil {
			t.Errorf("RelativeDateFilter failed: %v", err)
		}

		resultStr, ok := result.(string)
		if !ok {
			t.Errorf("Expected string, got %T", result)
		}

		if !strings.Contains(resultStr, "minute") && !strings.Contains(resultStr, "moment") {
			t.Errorf("Expected relative time format, got '%s'", resultStr)
		}

		// Test with older date
		old := now.Add(-25 * time.Hour)
		result, err = RelativeDateFilter(old)
		if err != nil {
			t.Errorf("RelativeDateFilter with old date failed: %v", err)
		}

		resultStr, ok = result.(string)
		if !ok {
			t.Errorf("Expected string, got %T", result)
		}

		if !strings.Contains(resultStr, "day") && !strings.Contains(resultStr, "hour") {
			t.Errorf("Expected day/hour in relative time, got '%s'", resultStr)
		}
	})

	// Test WeekdayFilter
	t.Run("WeekdayFilter", func(t *testing.T) {
		// Test with a known date (2023-01-01 was a Sunday)
		date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		result, err := WeekdayFilter(date)
		if err != nil {
			t.Errorf("WeekdayFilter failed: %v", err)
		}

		weekday, ok := result.(string)
		if !ok {
			t.Errorf("Expected string, got %T", result)
		}

		if weekday != "Sunday" {
			t.Errorf("Expected 'Sunday', got '%s'", weekday)
		}

		// Test with string input
		result, err = WeekdayFilter("2023-01-02T00:00:00Z") // Monday
		if err != nil {
			t.Errorf("WeekdayFilter with string failed: %v", err)
		}

		weekday, ok = result.(string)
		if !ok {
			t.Errorf("Expected string, got %T", result)
		}

		if weekday != "Monday" {
			t.Errorf("Expected 'Monday', got '%s'", weekday)
		}
	})

	// Test MonthNameFilter
	t.Run("MonthNameFilter", func(t *testing.T) {
		// Test with January
		date := time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)
		result, err := MonthNameFilter(date)
		if err != nil {
			t.Errorf("MonthNameFilter failed: %v", err)
		}

		month, ok := result.(string)
		if !ok {
			t.Errorf("Expected string, got %T", result)
		}

		if month != "January" {
			t.Errorf("Expected 'January', got '%s'", month)
		}

		// Test with December
		date = time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
		result, err = MonthNameFilter(date)
		if err != nil {
			t.Errorf("MonthNameFilter with December failed: %v", err)
		}

		month, ok = result.(string)
		if !ok {
			t.Errorf("Expected string, got %T", result)
		}

		if month != "December" {
			t.Errorf("Expected 'December', got '%s'", month)
		}
	})
}
