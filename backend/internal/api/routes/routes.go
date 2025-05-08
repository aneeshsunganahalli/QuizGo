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

	// Initialise Handlers
	authHandler := handlers.NewAuthHandler(db)

	// Auth Routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.RegisterUser)
		authRoutes.POST("/login", authHandler.Login)
	}

	protectedRoutes := router.Group("/")
	protectedRoutes.Use(middleware.AuthMiddleware())
}