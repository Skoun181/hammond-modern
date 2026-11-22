package db

import "time"

// MaintenanceTask belongs to a user and one of their accessible vehicles.
// Due status is computed on read; repeat rules are separate.
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
