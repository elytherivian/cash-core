package common

import (
	"testing"
	"time"
)

func TestLifecycleResponseConvertsTimesToConfiguredLocation(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	deletedAt := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	response := Lifecycle{
		CreatedAt: time.Date(2026, time.August, 9, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.August, 9, 11, 0, 0, 0, time.UTC),
		DeletedAt: &deletedAt,
	}.Response(location)

	if response.CreatedAt.Format(time.RFC3339) != "2026-08-09T18:00:00+08:00" ||
		response.UpdatedAt.Format(time.RFC3339) != "2026-08-09T19:00:00+08:00" ||
		response.DeletedAt == nil || response.DeletedAt.Format(time.RFC3339) != "2026-08-09T20:00:00+08:00" {
		t.Fatalf("response = %+v", response)
	}
}
