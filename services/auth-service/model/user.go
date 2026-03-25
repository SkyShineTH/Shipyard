package model

import "time"

// User is the database model for registered accounts.
// gorm.Model is embedded to provide ID, CreatedAt, UpdatedAt, DeletedAt automatically.
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null"     json:"email"`
	Password  string    `gorm:"not null"                 json:"-"` // never serialise password hash to JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
