package handlers

import (
	"FlashQuiz/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeckHandler struct {
	db *gorm.DB
}

func NewDeckHandler(db *gorm.DB) *DeckHandler {
	return &DeckHandler{db: db}
}

// CreateDeckRequest -> Struct for deck creation request
type CreateDeckRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IsPublic    bool   `json:"is_public"`
}

// CreateDeck -> Handler to create a new deck
func (h *DeckHandler) CreateDeck(c *gin.Context) {
	var req CreateDeckRequest
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

	// Create a new deck
	deck := models.Deck{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		IsPublic:    req.IsPublic,
		CardCount:   0,
		UserID:      userID.(uint),
	}

	// Save deck to database
	if err := h.db.Create(&deck).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deck"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Deck created successfully",
		"deck":    deck,
	})
}

// GetDecks -> Handler to get all decks for a user
func (h *DeckHandler) GetDecks(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get query parameters
	includePublic := c.Query("include_public") == "true"
	categoryFilter := c.Query("category")

	var decks []models.Deck
	query := h.db

	// Apply filters
	if includePublic {
		query = query.Where("user_id = ? OR is_public = ?", userID, true)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	if categoryFilter != "" {
		query = query.Where("category = ?", categoryFilter)
	}

	// Execute query
	if err := query.Find(&decks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve decks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"decks": decks,
	})
}

// GetDeckByID -> Handler to get a specific deck
func (h *DeckHandler) GetDeckByID(c *gin.Context) {
	deckID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var deck models.Deck
	// Get the deck with its flashcards
	if err := h.db.Preload("FlashCards").First(&deck, deckID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found"})
		return
	}

	// Check if user has permission to view this deck
	if !deck.IsPublic && deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this deck"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deck": deck,
	})
}

// UpdateDeckRequest -> Struct for deck update request
type UpdateDeckRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IsPublic    *bool  `json:"is_public"` // Pointer to differentiate between false and not provided
}

// UpdateDeck -> Handler to update a deck
func (h *DeckHandler) UpdateDeck(c *gin.Context) {
	deckID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var deck models.Deck
	if err := h.db.First(&deck, deckID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found"})
		return
	}

	// Check if user owns this deck
	if deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this deck"})
		return
	}

	var req UpdateDeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Title != "" {
		deck.Title = req.Title
	}
	if req.Description != "" {
		deck.Description = req.Description
	}
	if req.Category != "" {
		deck.Category = req.Category
	}
	if req.IsPublic != nil {
		deck.IsPublic = *req.IsPublic
	}

	// Save updated deck
	if err := h.db.Save(&deck).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deck"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deck updated successfully",
		"deck":    deck,
	})
}

// DeleteDeck -> Handler to delete a deck
func (h *DeckHandler) DeleteDeck(c *gin.Context) {
	deckID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deck ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var deck models.Deck
	if err := h.db.First(&deck, deckID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found"})
		return
	}

	// Check if user owns this deck
	if deck.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this deck"})
		return
	}

	// Delete the deck (GORM will handle the cascade deletion of related cards if set up properly)
	if err := h.db.Delete(&deck).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete deck"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deck deleted successfully",
	})
}
