package entity

import "time"

type Artist struct {
	ID             string
	Name           string
	Email          string
	Phone          string
	BusinessName   string
	PrimaryService string // "makeup" | "hair" | "attire"
	City           string
	Instagram      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	CreatedBy      string
	UpdatedBy      string
	DeletedBy      string
}
