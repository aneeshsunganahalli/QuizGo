package routes

import (
	"FlashQuiz/internal/api/handlers"
	"FlashQuiz/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Middleware
	router.Use(middleware.LoggerMiddleware())

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(db)
	deckHandler := handlers.NewDeckHandler(db)
	cardHandler := handlers.NewCardHandler(db)
	quizHandler := handlers.NewQuizHandler(db)
	studyHandler := handlers.NewStudyHandler(db)

	// Public routes for authentication
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.RegisterUser)
		authRoutes.POST("/login", authHandler.Login)
	}

	// Protected routes that require authentication
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Deck routes
		decks := api.Group("/decks")
		{
			decks.POST("", deckHandler.CreateDeck)
			decks.GET("", deckHandler.GetDecks)
			decks.GET("/:id", deckHandler.GetDeckByID)
			decks.PUT("/:id", deckHandler.UpdateDeck)
			decks.DELETE("/:id", deckHandler.DeleteDeck)
		}

		// Flashcard routes
		cards := api.Group("/cards")
		{
			cards.POST("", cardHandler.CreateCard)
			cards.GET("/:id", cardHandler.GetCardByID)
			cards.GET("/deck/:deck_id", cardHandler.GetCardsByDeck)
			cards.PUT("/:id", cardHandler.UpdateCard)
			cards.DELETE("/:id", cardHandler.DeleteCard)
			cards.POST("/bulk-import", cardHandler.BulkImportCards)
		}

		// Quiz routes
		quizzes := api.Group("/quizzes")
		{
			quizzes.POST("", quizHandler.CreateQuiz)
			quizzes.GET("", quizHandler.GetUserQuizzes)
			quizzes.GET("/:id", quizHandler.GetQuiz)
			quizzes.POST("/answer", quizHandler.SubmitQuizAnswer)
			quizzes.POST("/complete", quizHandler.CompleteQuiz)
		}

		// Study/Spaced repetition routes
		study := api.Group("/study")
		{
			study.POST("/next-cards", studyHandler.GetNextCards)
			study.POST("/update-progress", studyHandler.UpdateCardProgress)
			study.GET("/stats", studyHandler.GetStudyStats)
		}
	}
}
