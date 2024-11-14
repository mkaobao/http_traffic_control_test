package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginLimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memoryStore "github.com/ulule/limiter/v3/drivers/store/memory"
)

func unstableHandlerFactory(isUnstable bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isUnstable {
			c.JSON(http.StatusInternalServerError, gin.H{
				"time": time.Now().Format("2006-01-02 15:04:05"),
				"code": http.StatusInternalServerError,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"time": time.Now().Format("2006-01-02 15:04:05"),
				"code": http.StatusOK,
			})
		}
	}
}

func GenerateRateLimitMiddleware() gin.HandlerFunc {
	store := memoryStore.NewStore()
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  200,
	}

	limiterInstance := limiter.New(store, rate)
	limiterMiddleware := ginLimiter.NewMiddleware(limiterInstance)

	return limiterMiddleware
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// rate limit middleware
	router.Use(GenerateRateLimitMiddleware())

	router.GET("/", unstableHandlerFactory(false))
	router.GET("/fail", unstableHandlerFactory(true))

	fmt.Println("Generate Simple Http API Server on Port: [8080]")
	router.Run(":8080")

	// test command
	// siege -p -c2 -t3s -d0.01 http://localhost:8080
}
