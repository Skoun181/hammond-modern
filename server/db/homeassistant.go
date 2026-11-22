package db

import "time"

// HomeAssistantKey grants read-only access to this user's maintenance data.
// Only a SHA-256 hash is persisted; the bearer token is shown once.
type HomeAssistantKey struct {
	OwnerUserID string `gorm:"primaryKey;size:64"`
	TokenHash string `gorm:"index;not null;size:64"`
	CreatedAt time.Time
}
