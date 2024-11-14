package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vulcand/oxy/v2/cbreaker"
)

func unstableHandlerFactory(isUnstable bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// sleep 0.1 sec
		time.Sleep(100 * time.Millisecond)

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

func GenerateCircuitBreaker(router http.Handler) *cbreaker.CircuitBreaker {
	cb, err := cbreaker.New(router, "ResponseCodeRatio(500, 600, 0, 600) > 0.1",
		cbreaker.FallbackDuration(5*time.Second),
		cbreaker.RecoveryDuration(5*time.Second),
		cbreaker.CheckPeriod(20*time.Second),
	)
	if err != nil {
		fmt.Println("failed to crease cbreaker")
		os.Exit(1)
	}

	return cb
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.GET("/", unstableHandlerFactory(false))
	router.GET("/fail", unstableHandlerFactory(true))

	fmt.Println("Generate Simple Http API Server on Port: [8080]")
	cb := GenerateCircuitBreaker(router)
	http.ListenAndServe(":8080", cb)

	// test command
	// siege -p -c1 -t10m -d0.01 http://localhost:8080
	// siege -p -c1 -t3s -d0.01 http://localhost:8080/fail
}
