package user

import (
	"testing"
	"time"

	"cash-core/internal/common"

	"github.com/google/uuid"
)

func TestUserResponseUsesConfiguredLocation(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	response := User{
		ID: uuid.New(), Username: "cash-user",
		Lifecycle: common.Lifecycle{
			CreatedAt: time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, time.August, 9, 13, 0, 0, 0, time.UTC),
			IsActive:  true,
		},
	}.Response(location)

	if response.CreatedAt.Format(time.RFC3339) != "2026-08-09T20:00:00+08:00" ||
		response.UpdatedAt.Format(time.RFC3339) != "2026-08-09T21:00:00+08:00" {
		t.Fatalf("response = %+v", response)
	}
}
