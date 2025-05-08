package handlers

import (
	"FlashQuiz/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QuizHandler struct {
	db *gorm.DB
}

func NewQuizHandler(db *gorm.DB) *QuizHandler {
	return &QuizHandler{db: db}
}

// CreateQuizRequest -> Struct for quiz creation request
type CreateQuizRequest struct {
	DeckID      uint   `json:"deck_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	CardCount   int    `json:"card_count"` // Number of cards to include in quiz, 0 means all
}

// CreateQuiz -> Handler to create a new quiz
func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var req CreateQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify the deck exists and user has access to it
	var deck models.Deck
	if err := h.db.First(&deck, req.DeckID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found"})
		return
	}

	if !deck.IsPublic && deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to create a quiz for this deck"})
		return
	}

	// Get cards from the deck
	var cards []models.FlashCard
	query := h.db.Where("deck_id = ?", req.DeckID)

	// If card count is specified, limit the number of cards
	cardCount := req.CardCount
	if cardCount > 0 {
		query = query.Order("RANDOM()").Limit(cardCount)
	}

	if err := query.Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve flashcards"})
		return
	}

	if len(cards) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No cards available in this deck"})
		return
	}

	// Begin transaction to create quiz and questions
	tx := h.db.Begin()

	// Create the quiz
	quiz := models.Quiz{
		UserID:         userID.(uint),
		DeckID:         req.DeckID,
		Title:          req.Title,
		Description:    req.Description,
		TotalQuestions: len(cards),
	}

	if err := tx.Create(&quiz).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quiz"})
		return
	}

	// Create quiz questions for each card
	for _, card := range cards {
		question := models.QuizQuestion{
			QuizID:       quiz.ID,
			CardID:       card.ID,
			QuestionType: "recall", // Default question type
		}

		if err := tx.Create(&question).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quiz questions"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize quiz creation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Quiz created successfully",
		"quiz": gin.H{
			"id":              quiz.ID,
			"title":           quiz.Title,
			"description":     quiz.Description,
			"total_questions": quiz.TotalQuestions,
		},
	})
}

// GetQuiz -> Handler to get a quiz with its questions
func (h *QuizHandler) GetQuiz(c *gin.Context) {
	quizID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var quiz models.Quiz
	if err := h.db.First(&quiz, quizID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}

	// Only the quiz creator can access it
	if quiz.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this quiz"})
		return
	}

	// Get all questions with their associated cards
	var questions []models.QuizQuestion
	if err := h.db.Where("quiz_id = ?", quizID).Preload("FlashCard").Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quiz questions"})
		return
	}

	// Format the response
	formattedQuestions := make([]gin.H, 0, len(questions))
	for _, q := range questions {
		formattedQuestions = append(formattedQuestions, gin.H{
			"id":           q.ID,
			"question":     q.FlashCard.FrontContent,
			"answer":       q.FlashCard.BackContent,
			"content_type": q.FlashCard.ContentType,
			"user_answer":  q.UserAnswer,
			"is_correct":   q.IsCorrect,
			"time_spent":   q.TimeSpent,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"quiz": gin.H{
			"id":              quiz.ID,
			"title":           quiz.Title,
			"description":     quiz.Description,
			"created_at":      quiz.CreatedAt,
			"completed_at":    quiz.CompletedAt,
			"score":           quiz.Score,
			"correct_answers": quiz.CorrectAnswers,
			"total_questions": quiz.TotalQuestions,
			"questions":       formattedQuestions,
		},
	})
}

// GetUserQuizzes -> Handler to get all quizzes for a user
func (h *QuizHandler) GetUserQuizzes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var quizzes []models.Quiz
	if err := h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&quizzes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quizzes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quizzes": quizzes,
	})
}

// SubmitAnswerRequest -> Struct for quiz answer submission
type SubmitAnswerRequest struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
	TimeSpent  int    `json:"time_spent"`
}

// SubmitQuizAnswer -> Handler to submit an answer to a quiz question
func (h *QuizHandler) SubmitQuizAnswer(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Find the question
	var question models.QuizQuestion
	if err := h.db.Preload("Quiz").Preload("FlashCard").First(&question, req.QuestionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	// Check that this question belongs to a quiz owned by the user
	if question.Quiz.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to answer this question"})
		return
	}

	// Simple string comparison to check if the answer is correct
	// In a real app, you might want more sophisticated answer checking
	isCorrect := req.Answer == question.FlashCard.BackContent

	// Update the question with the user's answer
	question.UserAnswer = req.Answer
	question.IsCorrect = isCorrect
	question.TimeSpent = req.TimeSpent

	if err := h.db.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save answer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Answer submitted successfully",
		"is_correct":     isCorrect,
		"correct_answer": question.FlashCard.BackContent,
	})
}

// CompleteQuizRequest -> Struct for completing a quiz
type CompleteQuizRequest struct {
	QuizID uint `json:"quiz_id" binding:"required"`
}

// CompleteQuiz -> Handler to mark a quiz as complete and calculate score
func (h *QuizHandler) CompleteQuiz(c *gin.Context) {
	var req CompleteQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Find the quiz
	var quiz models.Quiz
	if err := h.db.First(&quiz, req.QuizID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}

	// Check that the quiz belongs to the user
	if quiz.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to complete this quiz"})
		return
	}

	// Don't allow completing an already completed quiz
	if quiz.CompletedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This quiz is already completed"})
		return
	}

	// Get all questions for the quiz
	var questions []models.QuizQuestion
	if err := h.db.Where("quiz_id = ?", req.QuizID).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quiz questions"})
		return
	}

	// Calculate score
	correctCount := 0
	for _, q := range questions {
		if q.IsCorrect {
			correctCount++
		}
	}

	// Update quiz with results
	now := time.Now()
	quiz.CompletedAt = &now
	quiz.CorrectAnswers = correctCount
	score := 0.0
	if quiz.TotalQuestions > 0 {
		score = float64(correctCount) / float64(quiz.TotalQuestions) * 100
	}
	quiz.Score = score

	if err := h.db.Save(&quiz).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete quiz"})
		return
	}

	// TODO: Could also update card progress here based on quiz results

	c.JSON(http.StatusOK, gin.H{
		"message":         "Quiz completed successfully",
		"score":           score,
		"correct_answers": correctCount,
		"total_questions": quiz.TotalQuestions,
	})
}
