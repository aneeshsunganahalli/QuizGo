package main

import (
	"FlashQuiz/internal/api/routes"
	"FlashQuiz/internal/models"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	// Load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Error loading .env file")
	} else {
		log.Println("Loaded Environment Variables from .env file")
	}

}

func initDB() (*gorm.DB, error) {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate all models
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Deck{},
		&models.FlashCard{},
		&models.CardProgress{},
		&models.Quiz{},
		&models.QuizQuestion{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("Database migrations completed successfully")

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Router Initilization
	router := gin.Default()

	// Config CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	// Set trusted Proxies
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to QuizGo API"})
	})

	routes.SetupRoutes(router, db)

	log.Printf("Server starting on port %s", "8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
