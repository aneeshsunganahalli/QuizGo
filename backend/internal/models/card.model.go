package models

import (
	"time"

	"gorm.io/gorm"
)

// Deck -> Group of flashcards
type Deck struct {
	gorm.Model
	Title       string      `json:"title" gorm:"not null"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	CardCount   int         `json:"card_count"`
	IsPublic    bool        `json:"is_public"`
	UserID      uint        `json:"user_id" gorm:"index"`
	User        User        `json:"-" gorm:"foreignKey:UserID"`
	FlashCards  []FlashCard `json:"flash_cards,omitempty" gorm:"foreignKey:DeckID"`
	Quizzes     []Quiz      `json:"quizzes,omitempty" gorm:"foreignKey:DeckID"`
}

type FlashCard struct {
	gorm.Model
	DeckID          uint           `json:"deck_id" gorm:"index"`
	Deck            Deck           `json:"-" gorm:"foreignKey:DeckID"`
	FrontContent    string         `json:"front_content" gorm:"not null"`
	BackContent     string         `json:"back_content" gorm:"not null"`
	ContentType     string         `json:"content_type" gorm:"default:'text'"`
	DifficultyLevel float64        `json:"difficulty_level" gorm:"default:0.5"`
	CardProgresses  []CardProgress `json:"card_progresses,omitempty" gorm:"foreignKey:CardID"`
	QuizQuestions   []QuizQuestion `json:"quiz_questions,omitempty" gorm:"foreignKey:CardID"`
}

// CardProgress -> User's progress on a specific flashcard
type CardProgress struct {
	gorm.Model
	UserID         uint      `json:"user_id" gorm:"index;not null"`
	User           User      `json:"-" gorm:"foreignKey:UserID"`
	CardID         uint      `json:"card_id" gorm:"index;not null"`
	FlashCard      FlashCard `json:"-" gorm:"foreignKey:CardID"`
	EaseFactor     float64   `json:"ease_factor" gorm:"default:2.5"`
	Interval       int       `json:"interval" gorm:"default:0"` // days
	NextReviewDate time.Time `json:"next_review_date"`
	ReviewCount    int       `json:"review_count" gorm:"default:0"`
	CorrectCount   int       `json:"correct_count" gorm:"default:0"`
	LastReviewedAt time.Time `json:"last_reviewed_at"`
	Status         string    `json:"status" gorm:"default:'new'"` // e.g., "new", "learning", "review"
}

type Quiz struct {
	gorm.Model
	UserID         uint           `json:"user_id" gorm:"index;not null"`
	User           User           `json:"-" gorm:"foreignKey:UserID"`
	DeckID         uint           `json:"deck_id" gorm:"index;not null"`
	Deck           Deck           `json:"-" gorm:"foreignKey:DeckID"`
	Title          string         `json:"title" gorm:"not null"`
	Description    string         `json:"description"`
	CompletedAt    *time.Time     `json:"completed_at"` // Using pointer for nullable time
	Score          float64        `json:"score" gorm:"default:0"`
	TotalQuestions int            `json:"total_questions" gorm:"default:0"`
	CorrectAnswers int            `json:"correct_answers" gorm:"default:0"`
	Questions      []QuizQuestion `json:"questions,omitempty" gorm:"foreignKey:QuizID"`
}

// QuizQuestion -> Represents a question in a quiz
type QuizQuestion struct {
	gorm.Model
	QuizID       uint      `json:"quiz_id" gorm:"index;not null"`
	Quiz         Quiz      `json:"-" gorm:"foreignKey:QuizID"`
	CardID       uint      `json:"card_id" gorm:"index;not null"`
	FlashCard    FlashCard `json:"-" gorm:"foreignKey:CardID"`
	QuestionType string    `json:"question_type" gorm:"default:'recall'"` // e.g., "multiple_choice", "true_false", "recall"
	UserAnswer   string    `json:"user_answer"`
	IsCorrect    bool      `json:"is_correct" gorm:"default:false"`
	TimeSpent    int       `json:"time_spent"` // in seconds
}
