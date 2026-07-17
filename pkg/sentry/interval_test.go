package sentry

import (
	"slices"
	"testing"
	"time"
)

func TestSnapSpansStatsInterval(t *testing.T) {
	tests := []struct {
		name      string
		interval  time.Duration
		timeRange time.Duration
		want      string
	}{
		{"below minimum falls back to range-derived interval", 14 * time.Second, time.Hour, "15s"},
		{"larger than range falls back to range-derived interval", 2 * time.Hour, time.Hour, "15s"},
		{"zero interval derives from range", 0, 4 * time.Hour, "30s"},

		{"minimum allowed is kept", 15 * time.Second, time.Hour, "15s"},
		{"snaps to nearest allowed value", 32 * time.Second, time.Hour, "30s"},
		{"snaps up to nearest allowed value", 28 * time.Second, time.Hour, "30s"},
		{"equidistant value snaps down", 45 * time.Second, time.Hour, "30s"},
		{"above maximum snaps to one day", 25 * time.Hour, 30 * 24 * time.Hour, "24h"},

		{"rounding up past the range snaps down to largest fitting", 50 * time.Second, 55 * time.Second, "30s"},
		{"range below minimum still sends the minimum", 10 * time.Second, 10 * time.Second, "15s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := snapSpansStatsInterval(tt.interval, tt.timeRange)
			if got != tt.want {
				t.Errorf("snapSpansStatsInterval(%v, %v) = %q, want %q", tt.interval, tt.timeRange, got, tt.want)
			}
		})
	}
}

func TestSnapSpansStatsIntervalUnsortedList(t *testing.T) {
	slices.Reverse(allowedSpansStatsIntervals)
	snapSpansStatsInterval(32*time.Second, time.Hour)
	if !slices.IsSorted(allowedSpansStatsIntervals) {
		t.Error("allowedSpansStatsIntervals should be sorted after the call")
	}
}