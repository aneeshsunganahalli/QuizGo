package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `json:"username" gorm:"uniqueIndex;not null"`
	Email        string `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string `json:"-" gorm:"not null"` // "-" means don't show in JSON responses

	// Relationships
	Decks          []Deck         `json:"decks,omitempty" gorm:"foreignKey:UserID"`
	CardProgresses []CardProgress `json:"card_progresses,omitempty" gorm:"foreignKey:UserID"`
	Quizzes        []Quiz         `json:"quizzes,omitempty" gorm:"foreignKey:UserID"`
}
