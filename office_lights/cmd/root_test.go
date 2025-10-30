package cmd

import (
	"testing"
	"time"
)

func TestParseAPITime(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Amsterdam")
	date := time.Date(2025, 7, 12, 0, 0, 0, 0, location)

	tests := []struct {
		name       string
		timeStr    string
		date       time.Time
		location   *time.Location
		wantHour   int
		wantMinute int
		wantSecond int
		wantError  bool
	}{
		{
			name:       "Morning time",
			timeStr:    "5:35:12 AM",
			date:       date,
			location:   location,
			wantHour:   5,
			wantMinute: 35,
			wantSecond: 12,
			wantError:  false,
		},
		{
			name:       "Evening time",
			timeStr:    "9:49:57 PM",
			date:       date,
			location:   location,
			wantHour:   21,
			wantMinute: 49,
			wantSecond: 57,
			wantError:  false,
		},
		{
			name:       "Noon",
			timeStr:    "12:00:00 PM",
			date:       date,
			location:   location,
			wantHour:   12,
			wantMinute: 0,
			wantSecond: 0,
			wantError:  false,
		},
		{
			name:       "Midnight",
			timeStr:    "12:00:00 AM",
			date:       date,
			location:   location,
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
			wantError:  false,
		},
		{
			name:      "Invalid format",
			timeStr:   "25:00:00 PM",
			date:      date,
			location:  location,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAPITime(tt.timeStr, tt.date, tt.location)
			if (err != nil) != tt.wantError {
				t.Errorf("parseAPITime() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if got.Hour() != tt.wantHour {
					t.Errorf("parseAPITime() hour = %v, want %v", got.Hour(), tt.wantHour)
				}
				if got.Minute() != tt.wantMinute {
					t.Errorf("parseAPITime() minute = %v, want %v", got.Minute(), tt.wantMinute)
				}
				if got.Second() != tt.wantSecond {
					t.Errorf("parseAPITime() second = %v, want %v", got.Second(), tt.wantSecond)
				}
				// Verify the date is correct
				if got.Year() != tt.date.Year() || got.Month() != tt.date.Month() || got.Day() != tt.date.Day() {
					t.Errorf("parseAPITime() date = %v, want %v", got.Format("2006-01-02"), tt.date.Format("2006-01-02"))
				}
			}
		})
	}
}

func TestGetTime(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Amsterdam")
	date := time.Date(2025, 7, 12, 0, 0, 0, 0, location)

	tests := []struct {
		name       string
		timeStr    string
		date       time.Time
		location   *time.Location
		wantHour   int
		wantMinute int
		wantSecond int
	}{
		{
			name:       "Custom morning time",
			timeStr:    "8:30:45 AM",
			date:       date,
			location:   location,
			wantHour:   8,
			wantMinute: 30,
			wantSecond: 45,
		},
		{
			name:       "Custom evening time",
			timeStr:    "6:15:30 PM",
			date:       date,
			location:   location,
			wantHour:   18,
			wantMinute: 15,
			wantSecond: 30,
		},
		{
			name:       "Midnight",
			timeStr:    "12:00:00 AM",
			date:       date,
			location:   location,
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
		},
		{
			name:       "Noon",
			timeStr:    "12:00:00 PM",
			date:       date,
			location:   location,
			wantHour:   12,
			wantMinute: 0,
			wantSecond: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTime(tt.timeStr, tt.date, tt.location)
			if got.Hour() != tt.wantHour {
				t.Errorf("getTime() hour = %v, want %v", got.Hour(), tt.wantHour)
			}
			if got.Minute() != tt.wantMinute {
				t.Errorf("getTime() minute = %v, want %v", got.Minute(), tt.wantMinute)
			}
			if got.Second() != tt.wantSecond {
				t.Errorf("getTime() second = %v, want %v", got.Second(), tt.wantSecond)
			}
			// Verify date is preserved
			if got.Year() != tt.date.Year() || got.Month() != tt.date.Month() || got.Day() != tt.date.Day() {
				t.Errorf("getTime() date = %v, want %v", got.Format("2006-01-02"), tt.date.Format("2006-01-02"))
			}
		})
	}
}

func TestTimezoneDST(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Amsterdam")

	// Test summer time (CEST - UTC+2)
	summerDate := time.Date(2025, 7, 12, 0, 0, 0, 0, location)
	summerTime, _ := parseAPITime("9:49:57 PM", summerDate, location)

	// Test winter time (CET - UTC+1)
	winterDate := time.Date(2025, 12, 15, 0, 0, 0, 0, location)
	winterTime, _ := parseAPITime("4:30:00 PM", winterDate, location)

	// Get timezone offset
	_, summerOffset := summerTime.Zone()
	_, winterOffset := winterTime.Zone()

	// Summer should be UTC+2 (7200 seconds)
	if summerOffset != 7200 {
		t.Errorf("Summer time offset = %v, want 7200 (UTC+2)", summerOffset)
	}

	// Winter should be UTC+1 (3600 seconds)
	if winterOffset != 3600 {
		t.Errorf("Winter time offset = %v, want 3600 (UTC+1)", winterOffset)
	}
}

// TestLightsShouldBeOn tests the core logic for determining if lights should be on
func TestLightsShouldBeOn(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Amsterdam")
	date := time.Date(2025, 7, 12, 0, 0, 0, 0, location)

	// Simulate sunrise at 5:35:12 AM and sunset at 9:49:57 PM
	sunrise, _ := parseAPITime("5:35:12 AM", date, location)
	sunset, _ := parseAPITime("9:49:57 PM", date, location)

	tests := []struct {
		name        string
		checkTime   string
		expectedOn  bool
		description string
	}{
		{
			name:        "Before sunrise",
			checkTime:   "5:00:00 AM",
			expectedOn:  true,
			description: "Lights should be ON before sunrise",
		},
		{
			name:        "Exact sunrise time",
			checkTime:   "5:35:12 AM",
			expectedOn:  true,
			description: "At exact sunrise, lights should still be ON (not after sunrise)",
		},
		{
			name:        "One second after sunrise",
			checkTime:   "5:35:13 AM",
			expectedOn:  false,
			description: "One second after sunrise, lights should be OFF",
		},
		{
			name:        "Midday",
			checkTime:   "12:00:00 PM",
			expectedOn:  false,
			description: "During the day, lights should be OFF",
		},
		{
			name:        "Afternoon",
			checkTime:   "3:30:00 PM",
			expectedOn:  false,
			description: "In the afternoon, lights should be OFF",
		},
		{
			name:        "One second before sunset",
			checkTime:   "9:49:56 PM",
			expectedOn:  false,
			description: "One second before sunset, lights should still be OFF",
		},
		{
			name:        "Exact sunset time",
			checkTime:   "9:49:57 PM",
			expectedOn:  true,
			description: "At exact sunset, lights should be ON (not before sunset)",
		},
		{
			name:        "After sunset",
			checkTime:   "10:00:00 PM",
			expectedOn:  true,
			description: "After sunset, lights should be ON",
		},
		{
			name:        "Late night",
			checkTime:   "11:59:59 PM",
			expectedOn:  true,
			description: "Late at night, lights should be ON",
		},
		{
			name:        "Midnight",
			checkTime:   "12:00:00 AM",
			expectedOn:  true,
			description: "At midnight, lights should be ON",
		},
		{
			name:        "Early morning",
			checkTime:   "3:00:00 AM",
			expectedOn:  true,
			description: "In early morning before sunrise, lights should be ON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTime, _ := parseAPITime(tt.checkTime, date, location)
			gotOn := lightsShouldBeOn(checkTime, sunrise, sunset, false)

			if gotOn != tt.expectedOn {
				t.Errorf("%s: got lights ON=%v, want ON=%v. Time: %s, Sunrise: %s, Sunset: %s",
					tt.description, gotOn, tt.expectedOn,
					checkTime.Format("3:04:05 PM"),
					sunrise.Format("3:04:05 PM"),
					sunset.Format("3:04:05 PM"))
			}
		})
	}
}

// TestWinterVsSummerTimes tests the lights logic across different seasons
func TestWinterVsSummerTimes(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Amsterdam")

	tests := []struct {
		name       string
		date       string
		sunrise    string
		sunset     string
		checkTime  string
		expectedOn bool
	}{
		{
			name:       "Summer - midday should be OFF",
			date:       "12-07-2025",
			sunrise:    "5:35:12 AM",
			sunset:     "9:49:57 PM",
			checkTime:  "12:00:00 PM",
			expectedOn: false,
		},
		{
			name:       "Summer - early morning should be ON",
			date:       "12-07-2025",
			sunrise:    "5:35:12 AM",
			sunset:     "9:49:57 PM",
			checkTime:  "4:00:00 AM",
			expectedOn: true,
		},
		{
			name:       "Winter - midday should be OFF (short day)",
			date:       "15-12-2025",
			sunrise:    "8:42:00 AM",
			sunset:     "4:28:00 PM",
			checkTime:  "12:00:00 PM",
			expectedOn: false,
		},
		{
			name:       "Winter - 5 PM should be ON (already dark)",
			date:       "15-12-2025",
			sunrise:    "8:42:00 AM",
			sunset:     "4:28:00 PM",
			checkTime:  "5:00:00 PM",
			expectedOn: true,
		},
		{
			name:       "Winter - 7 AM should be ON (still dark)",
			date:       "15-12-2025",
			sunrise:    "8:42:00 AM",
			sunset:     "4:28:00 PM",
			checkTime:  "7:00:00 AM",
			expectedOn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := getDate(tt.date)
			sunrise, _ := parseAPITime(tt.sunrise, date, location)
			sunset, _ := parseAPITime(tt.sunset, date, location)
			checkTime, _ := parseAPITime(tt.checkTime, date, location)

			gotOn := lightsShouldBeOn(checkTime, sunrise, sunset, false)

			if gotOn != tt.expectedOn {
				t.Errorf("%s: got lights ON=%v, want ON=%v", tt.name, gotOn, tt.expectedOn)
			}
		})
	}
}

// TestDSTTransition tests behavior around Daylight Saving Time transitions
func TestDSTTransition(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Amsterdam")

	// DST starts: Last Sunday of March (March 30, 2025 - clocks forward)
	// DST ends: Last Sunday of October (October 26, 2025 - clocks backward)

	tests := []struct {
		name       string
		date       string
		wantOffset int // in seconds
	}{
		{
			name:       "Before DST starts (March)",
			date:       "15-03-2025",
			wantOffset: 3600, // CET is UTC+1
		},
		{
			name:       "After DST starts (April)",
			date:       "15-04-2025",
			wantOffset: 7200, // CEST is UTC+2
		},
		{
			name:       "During summer (July)",
			date:       "12-07-2025",
			wantOffset: 7200, // CEST is UTC+2
		},
		{
			name:       "After DST ends (November)",
			date:       "15-11-2025",
			wantOffset: 3600, // CET is UTC+1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := getDate(tt.date)
			testTime := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, location)

			_, offset := testTime.Zone()

			if offset != tt.wantOffset {
				t.Errorf("%s: offset = %d seconds, want %d seconds", tt.name, offset, tt.wantOffset)
			}
		})
	}
}
