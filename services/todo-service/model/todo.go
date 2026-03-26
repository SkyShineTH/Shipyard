package model

import "time"

// Todo is owned by one user (UserID from JWT). Other users never see it.
type Todo struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"index;not null"`
	Title     string    `json:"title" gorm:"not null"`
	Completed bool      `json:"completed" gorm:"not null;default:false"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
