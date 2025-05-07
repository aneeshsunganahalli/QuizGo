package models

import (
	"golang.org/x/crypto/bcrypt"
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

// Hash Password -> Hashes the password using bcrypt and stores it in PasswordHash, straight from documentation
func (u *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}
	
	u.PasswordHash = string(bytes)
	return nil
}

// CheckPassword -> Compares the provided password with the stored hashed password, also from documentation
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}
