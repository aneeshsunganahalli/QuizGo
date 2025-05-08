package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Verifies the JWT Token and passes user information into the context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context){
		// Debug Output
		fmt.Println("Auth middleware processing request: ", c.Request.URL.Path)

		// Getting Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Println("No Authorization Header found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization Header Missing"})
			c.Abort()
			return
		}

		// Check if its a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			fmt.Println("Invalid Authorization format: ", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Header must be in the format Bearer <token>"})
			c.Abort()
			return
		}

		// Getting the token
		tokenString := parts[1]
		fmt.Println("Token recieved: ", tokenString[:10], "...")

		// Parse and validate token
		claims := &jwt.MapClaims{}

		secretKey := os.Getenv("JWT_SECRET")
		
		// Parsing the token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected siging method")
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			fmt.Println("Token parsing error: ", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if !token.Valid{
			fmt.Println("Invalid Token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}

		// Set user info into context
		userID, ok := (*claims)["user_id"]
		if !ok {
			fmt.Println("No user_id found in token claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Convert userId into uint and pass into context
		userIDValue := uint(userID.(float64))
		fmt.Println("Setting user_id in context: ", userIDValue)
		c.Set("user_id", userIDValue)
		c.Set("username", (*claims)["username"])
		c.Next()
	}
}