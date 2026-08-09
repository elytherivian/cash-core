package common

import "time"

type Lifecycle struct {
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	IsActive  bool       `gorm:"column:is_active;not null;default:true" json:"is_active"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

type LifecycleResponse struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	IsActive  bool       `json:"is_active"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (l Lifecycle) Response(location *time.Location) LifecycleResponse {
	if location == nil {
		location = time.UTC
	}
	response := LifecycleResponse{
		CreatedAt: l.CreatedAt.In(location), UpdatedAt: l.UpdatedAt.In(location), IsActive: l.IsActive,
	}
	if l.DeletedAt != nil {
		deletedAt := l.DeletedAt.In(location)
		response.DeletedAt = &deletedAt
	}
	return response
}

type Page struct {
	Limit  int
	Offset int
}

type PageData struct {
	Items  any   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
