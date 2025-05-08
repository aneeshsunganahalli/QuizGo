package handlers

import (
	"FlashQuiz/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CardHandler struct {
	db *gorm.DB
}

func NewCardHandler(db *gorm.DB) *CardHandler {
	return &CardHandler{db: db}
}

// CreateCardRequest -> Struct for flashcard creation request
type CreateCardRequest struct {
	DeckID          uint    `json:"deck_id" binding:"required"`
	FrontContent    string  `json:"front_content" binding:"required"`
	BackContent     string  `json:"back_content" binding:"required"`
	ContentType     string  `json:"content_type"`
	DifficultyLevel float64 `json:"difficulty_level"`
}

// CreateCard -> Handler to create a new flashcard
func (h *CardHandler) CreateCard(c *gin.Context) {
	var req CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify the deck exists and belongs to the user
	var deck models.Deck
	if err := h.db.Where("id = ? AND user_id = ?", req.DeckID, userID).First(&deck).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found or you don't have permission to add cards to it"})
		return
	}

	// Set default values if not provided
	contentType := req.ContentType
	if contentType == "" {
		contentType = "text"
	}

	difficultyLevel := req.DifficultyLevel
	if difficultyLevel == 0 {
		difficultyLevel = 0.5 // default difficulty
	}

	// Create a new flashcard
	card := models.FlashCard{
		DeckID:          req.DeckID,
		FrontContent:    req.FrontContent,
		BackContent:     req.BackContent,
		ContentType:     contentType,
		DifficultyLevel: difficultyLevel,
	}

	// Save card to database
	if err := h.db.Create(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create flashcard"})
		return
	}

	// Update card count in the deck
	if err := h.db.Model(&deck).Update("card_count", deck.CardCount+1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deck card count"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Flashcard created successfully",
		"card":    card,
	})
}

// GetCardByID -> Handler to get a specific flashcard
func (h *CardHandler) GetCardByID(c *gin.Context) {
	cardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var card models.FlashCard
	// Get the card and its associated deck
	if err := h.db.Preload("Deck").First(&card, cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Flashcard not found"})
		return
	}

	// Check if user has permission to view this card
	// (either the user owns the deck or the deck is public)
	if !card.Deck.IsPublic && card.Deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this flashcard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"card": card,
	})
}

// GetCardsByDeck -> Handler to get all flashcards in a deck
func (h *CardHandler) GetCardsByDeck(c *gin.Context) {
	deckID, err := strconv.ParseUint(c.Param("deck_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if the user has access to this deck
	var deck models.Deck
	if err := h.db.First(&deck, deckID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found"})
		return
	}

	if !deck.IsPublic && deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view cards in this deck"})
		return
	}

	// Get all cards in the deck
	var cards []models.FlashCard
	if err := h.db.Where("deck_id = ?", deckID).Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve flashcards"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cards": cards,
	})
}

// UpdateCardRequest -> Struct for flashcard update request
type UpdateCardRequest struct {
	FrontContent    string  `json:"front_content"`
	BackContent     string  `json:"back_content"`
	ContentType     string  `json:"content_type"`
	DifficultyLevel float64 `json:"difficulty_level"`
}

// UpdateCard -> Handler to update a flashcard
func (h *CardHandler) UpdateCard(c *gin.Context) {
	cardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var card models.FlashCard
	// Get the card with its deck
	if err := h.db.Preload("Deck").First(&card, cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Flashcard not found"})
		return
	}

	// Check if user owns the deck that contains this card
	if card.Deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this flashcard"})
		return
	}

	var req UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.FrontContent != "" {
		card.FrontContent = req.FrontContent
	}
	if req.BackContent != "" {
		card.BackContent = req.BackContent
	}
	if req.ContentType != "" {
		card.ContentType = req.ContentType
	}
	if req.DifficultyLevel != 0 {
		card.DifficultyLevel = req.DifficultyLevel
	}

	// Save updated card
	if err := h.db.Save(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update flashcard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Flashcard updated successfully",
		"card":    card,
	})
}

// DeleteCard -> Handler to delete a flashcard
func (h *CardHandler) DeleteCard(c *gin.Context) {
	cardID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var card models.FlashCard
	// Get the card with its deck
	if err := h.db.Preload("Deck").First(&card, cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Flashcard not found"})
		return
	}

	// Check if user owns the deck that contains this card
	if card.Deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this flashcard"})
		return
	}

	// Begin a transaction to delete the card and update the deck's card count
	tx := h.db.Begin()
	
	// Delete the card
	if err := tx.Delete(&card).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete flashcard"})
		return
	}
	
	// Update deck's card count
	if err := tx.Model(&card.Deck).Update("card_count", card.Deck.CardCount-1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deck card count"})
		return
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process changes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Flashcard deleted successfully",
	})
}

// BulkImportRequest -> Struct for bulk importing cards
type BulkImportRequest struct {
	DeckID uint                   `json:"deck_id" binding:"required"`
	Cards  []BulkImportCardEntry `json:"cards" binding:"required"`
}

type BulkImportCardEntry struct {
	FrontContent string `json:"front_content" binding:"required"`
	BackContent  string `json:"back_content" binding:"required"`
	ContentType  string `json:"content_type"`
}

// BulkImportCards -> Handler to import multiple cards at once
func (h *CardHandler) BulkImportCards(c *gin.Context) {
	var req BulkImportRequest
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

	// Verify the deck exists and belongs to the user
	var deck models.Deck
	if err := h.db.Where("id = ? AND user_id = ?", req.DeckID, userID).First(&deck).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found or you don't have permission to add cards to it"})
		return
	}

	// Begin a transaction for bulk import
	tx := h.db.Begin()

	importedCards := make([]models.FlashCard, 0, len(req.Cards))
	for _, cardEntry := range req.Cards {
		contentType := cardEntry.ContentType
		if contentType == "" {
			contentType = "text"
		}

		card := models.FlashCard{
			DeckID:          req.DeckID,
			FrontContent:    cardEntry.FrontContent,
			BackContent:     cardEntry.BackContent,
			ContentType:     contentType,
			DifficultyLevel: 0.5, // default difficulty
		}

		if err := tx.Create(&card).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import cards"})
			return
		}

		importedCards = append(importedCards, card)
	}

	// Update card count in the deck
	newCardCount := deck.CardCount + len(req.Cards)
	if err := tx.Model(&deck).Update("card_count", newCardCount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deck card count"})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process import"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Cards imported successfully",
		"imported":     len(importedCards),
		"cards":        importedCards,
		"new_count":    newCardCount,
	})
}