package db

import "time"

// MaintenanceTask belongs to a user and one of their accessible vehicles.
// No notifications are sent; due status is computed on read.
type MaintenanceTask struct {
	Base
	OwnerUserID string     `json:"-" gorm:"index;not null"`
	VehicleID   string     `json:"vehicleId" gorm:"index;not null"`
	Title       string     `json:"title" gorm:"not null"`
	Notes       string     `json:"notes"`
	DueDate     *time.Time `json:"dueDate"`
	DueOdo      *int       `json:"dueOdo"`
	CompletedAt *time.Time `json:"completedAt"`
}
