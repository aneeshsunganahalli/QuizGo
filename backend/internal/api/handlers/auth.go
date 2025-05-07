package handlers

import (
	"FlashQuiz/internal/models"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// RegisterRequest -> Struct for user registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

// LoginRequest -> Struct for user login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func generateJWT(user models.User) (string,error) {

	secretKey := os.Getenv("JWT_SECRET_KEY")
		fmt.Println("----JWT secret key", secretKey)
	
	claims := JWTClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Token expires in 24 hours
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	// Payload within the JWT now contains the username and expiration time
	// We can create the JWT token using the claims and the secret key now

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString , err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}


