package handlers

import (
	"FlashQuiz/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StudyHandler struct {
	db *gorm.DB
}

func NewStudyHandler(db *gorm.DB) *StudyHandler {
	return &StudyHandler{db: db}
}

// GetNextCardsRequest -> Struct for getting next cards to study
type GetNextCardsRequest struct {
	DeckID uint `json:"deck_id" binding:"required"`
	Limit  int  `json:"limit"`
}

// GetNextCards -> Get the next flashcards due for review
func (h *StudyHandler) GetNextCards(c *gin.Context) {
	var req GetNextCardsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Default limit to 20 if not specified
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	// Verify the deck exists and user has access to it
	var deck models.Deck
	if err := h.db.First(&deck, req.DeckID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found"})
		return
	}

	if !deck.IsPublic && deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to study this deck"})
		return
	}

	// Get cards and their progress
	// First, get all cards from the deck
	var cards []models.FlashCard
	if err := h.db.Where("deck_id = ?", req.DeckID).Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve flashcards"})
		return
	}

	if len(cards) == 0 {
		c.JSON(http.StatusOK, gin.H{"cards": []string{}, "message": "No cards in this deck"})
		return
	}

	// Extract card IDs
	cardIDs := make([]uint, len(cards))
	for i, card := range cards {
		cardIDs[i] = card.ID
	}

	// Find existing progress records for these cards
	var progresses []models.CardProgress
	if err := h.db.Where("user_id = ? AND card_id IN ?", userID, cardIDs).Find(&progresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve card progress"})
		return
	}

	// Map card IDs to their progress
	progressMap := make(map[uint]*models.CardProgress)
	for i := range progresses {
		progressMap[progresses[i].CardID] = &progresses[i]
	}

	// Group cards by their status: new, due for review, and learning
	var newCards, dueCards, learningCards []models.FlashCard
	now := time.Now()

	for _, card := range cards {
		progress, exists := progressMap[card.ID]
		if !exists {
			// Card has no progress record - it's new
			newCards = append(newCards, card)
			continue
		}

		// Card has progress
		if progress.NextReviewDate.Before(now) {
			// Card is due for review
			dueCards = append(dueCards, card)
		} else {
			// Card is still being learned but not due yet
			learningCards = append(learningCards, card)
		}
	}

	// Prioritize due cards, then new cards
	var cardsToReturn []gin.H
	remainingLimit := limit

	for _, card := range dueCards {
		if remainingLimit <= 0 {
			break
		}

		progress := progressMap[card.ID]
		cardsToReturn = append(cardsToReturn, gin.H{
			"card":     card,
			"progress": progress,
			"status":   "due",
		})
		remainingLimit--
	}

	for _, card := range newCards {
		if remainingLimit <= 0 {
			break
		}

		cardsToReturn = append(cardsToReturn, gin.H{
			"card":     card,
			"progress": nil,
			"status":   "new",
		})
		remainingLimit--
	}

	c.JSON(http.StatusOK, gin.H{
		"cards":          cardsToReturn,
		"due_count":      len(dueCards),
		"new_count":      len(newCards),
		"learning_count": len(learningCards),
	})
}

// UpdateCardProgressRequest -> Struct for updating card progress
type UpdateCardProgressRequest struct {
	CardID      uint `json:"card_id" binding:"required"`
	Performance int  `json:"performance" binding:"required,min=1,max=5"` // 1-5 scale, where 1=fail, 5=perfect
	TimeSpent   int  `json:"time_spent"`                                 // Time spent on review in seconds
}

// UpdateCardProgress -> Update a card's progress after the user reviews it
func (h *StudyHandler) UpdateCardProgress(c *gin.Context) {
	var req UpdateCardProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify the card exists
	var card models.FlashCard
	if err := h.db.Preload("Deck").First(&card, req.CardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
		return
	}

	// Verify the user has access to the card's deck
	if !card.Deck.IsPublic && card.Deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this card's progress"})
		return
	}

	// Get or create progress record
	var progress models.CardProgress
	err := h.db.Where("user_id = ? AND card_id = ?", userID, req.CardID).First(&progress).Error

	isNew := err != nil

	if isNew {
		// Create a new progress record
		progress = models.CardProgress{
			UserID:       userID.(uint),
			CardID:       req.CardID,
			EaseFactor:   2.5, // Default value
			Interval:     0,
			ReviewCount:  0,
			CorrectCount: 0,
			Status:       "new",
		}
	}

	// Update the progress using the SuperMemo SM-2 algorithm
	// This is a simplified version of the algorithm

	// Update last reviewed time
	progress.LastReviewedAt = time.Now()
	progress.ReviewCount++

	// Update ease factor and interval based on performance
	// SM-2 algorithm uses a 0-5 scale, where:
	// 0 = complete blackout, 1 = incorrect but remembered, 2 = incorrect but close
	// 3 = correct but difficult, 4 = correct, 5 = correct and easy

	isCorrect := req.Performance >= 3
	if isCorrect {
		progress.CorrectCount++
	}

	// Calculate new ease factor (EF)
	easeFactor := progress.EaseFactor + (0.1 - (5-float64(req.Performance))*(0.08+(5-float64(req.Performance))*0.02))
	if easeFactor < 1.3 {
		easeFactor = 1.3 // Minimum ease factor
	}
	progress.EaseFactor = easeFactor

	// Calculate new interval
	var newInterval int
	if req.Performance < 3 {
		// If response was incorrect, start over
		newInterval = 1
		progress.Status = "learning"
	} else {
		// If response was correct, increase interval
		if progress.Interval == 0 {
			newInterval = 1
		} else if progress.Interval == 1 {
			newInterval = 6
		} else {
			newInterval = int(float64(progress.Interval) * progress.EaseFactor)
		}

		if progress.Status == "new" {
			progress.Status = "learning"
		} else if newInterval > 21 {
			// After 3 weeks interval, consider it "review" status
			progress.Status = "review"
		}
	}

	progress.Interval = newInterval
	progress.NextReviewDate = progress.LastReviewedAt.AddDate(0, 0, newInterval)

	// Save the progress
	if isNew {
		if err := h.db.Create(&progress).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create card progress"})
			return
		}
	} else {
		if err := h.db.Save(&progress).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card progress"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Card progress updated successfully",
		"progress":         progress,
		"next_review_date": progress.NextReviewDate.Format(time.RFC3339),
		"interval_days":    progress.Interval,
	})
}

// GetStudyStatsRequest -> Struct for getting study stats
type GetStudyStatsRequest struct {
	DeckID uint `form:"deck_id"`
}

// GetStudyStats -> Get statistics about a user's study progress
func (h *StudyHandler) GetStudyStats(c *gin.Context) {
	var req GetStudyStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Base query for card progress
	query := h.db.Model(&models.CardProgress{}).Where("user_id = ?", userID)

	// If deck ID is specified, filter by deck
	if req.DeckID > 0 {
		// Verify the deck exists and user has access to it
		var deck models.Deck
		if err := h.db.First(&deck, req.DeckID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found"})
			return
		}

		if !deck.IsPublic && deck.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this deck's stats"})
			return
		}

		// Join with FlashCard to filter by deck_id
		query = query.Joins("JOIN flash_cards ON card_progresses.card_id = flash_cards.id").
			Where("flash_cards.deck_id = ?", req.DeckID)
	}

	// Get counts by status
	var newCount, learningCount, reviewCount int64

	if err := query.Where("status = ?", "new").Count(&newCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	if err := query.Where("status = ?", "learning").Count(&learningCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	if err := query.Where("status = ?", "review").Count(&reviewCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	// Get total cards reviewed and correct percentage
	var totalReviewed, totalCorrect int64

	// Fix: Using Select() instead of Sum() for aggregation
	if err := query.Select("COALESCE(SUM(review_count), 0)").Scan(&totalReviewed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	// Fix: Using Select() instead of Sum() for aggregation
	if err := query.Select("COALESCE(SUM(correct_count), 0)").Scan(&totalCorrect).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	// Calculate accuracy percentage
	var accuracyPercentage float64 = 0
	if totalReviewed > 0 {
		accuracyPercentage = float64(totalCorrect) / float64(totalReviewed) * 100
	}

	// Get cards due today
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

	var dueToday int64
	if err := query.Where("next_review_date <= ?", endOfDay).Count(&dueToday).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	// Get recent activity (reviews per day for the last week)
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day()-7, 0, 0, 0, 0, now.Location())

	type DailyActivity struct {
		Date    string `json:"date"`
		Reviews int64  `json:"reviews"`
	}

	var dailyActivity []DailyActivity

	// For simplicity, we're using SQLite's date() function
	// In a production app, you might need to adjust this for your specific database
	rows, err := h.db.Raw(`
		SELECT 
			date(last_reviewed_at) as review_date, 
			COUNT(*) as reviews
		FROM 
			card_progresses
		WHERE 
			user_id = ? 
			AND last_reviewed_at >= ?
		GROUP BY 
			date(last_reviewed_at)
		ORDER BY 
			review_date ASC
	`, userID, startOfWeek).Rows()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve activity stats"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var date string
		var reviews int64
		if err := rows.Scan(&date, &reviews); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse activity stats"})
			return
		}
		dailyActivity = append(dailyActivity, DailyActivity{Date: date, Reviews: reviews})
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"new_count":      newCount,
			"learning_count": learningCount,
			"review_count":   reviewCount,
			"due_today":      dueToday,
			"total_reviewed": totalReviewed,
			"accuracy":       accuracyPercentage,
			"daily_activity": dailyActivity,
		},
	})
}
