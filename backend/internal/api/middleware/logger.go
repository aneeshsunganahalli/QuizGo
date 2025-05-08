package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc{
	return func(c *gin.Context){
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Printf("Request: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
	}
}