package main

import (
	"log"
	// "github.com/gin-contrib/cors"
	"FlashQuiz/internal/api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init(){
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

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	 // Router Initilization
	 router := gin.Default()

	 // Set trusted Proxies
	 router.SetTrustedProxies([]string{"127.0.0.1"})

	 router.GET("/", func(c *gin.Context){
		c.JSON(200, gin.H{
			"message" : "Welcome to QuizGo API"})
	 })

	 routes.SetupRoutes(router, db)

	 router.Run(":8080")
}
