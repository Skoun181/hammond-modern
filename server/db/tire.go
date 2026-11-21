package db

import "time"

// TireDetails contains per-tire metadata. Optional measurements use pointers
// so an unknown value is distinguishable from a measured zero.
type TireDetails struct {
	VehicleID    string   `json:"vehicleId" gorm:"index"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	Season       string   `json:"season"`
	Size         string   `json:"size"`
	DOT          string   `json:"dot"`
	SetName      string   `json:"setName"`
	LoadIndex    string   `json:"loadIndex"`
	SpeedIndex   string   `json:"speedIndex"`
	RunFlat      bool     `json:"runFlat"`
	Status       string   `json:"status"`
	Position     string   `json:"position"`
	TreadDepth   *float64 `json:"treadDepth"`
	Pressure     *float64 `json:"pressure"`
	PurchaseDate string   `json:"purchaseDate"`
	PurchasePrice *float64 `json:"purchasePrice"`
	Currency     string   `json:"currency"`
	Mileage      *int64   `json:"mileage"`
	Storage      string   `json:"storage"`
	Notes        string   `json:"notes"`
}

type Tire struct {
	ID string `json:"id" gorm:"primaryKey;size:32"`
	OwnerUserID string `json:"-" gorm:"index;not null"`
	TireDetails `gorm:"embedded"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
