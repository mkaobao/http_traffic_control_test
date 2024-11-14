package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

type RequestContext struct {
	c        *gin.Context
	doneChan chan struct{}
}

func worker(queue chan RequestContext) {
	for reqCtx := range queue {
		fmt.Println("queue len: ", len(queue))

		reqCtx.c.Next()

		close(reqCtx.doneChan)
	}
}

func queueMiddleware(queue chan RequestContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		done := make(chan struct{})

		reqCtx := RequestContext{
			c:        c,
			doneChan: done,
		}

		select {
		case queue <- reqCtx:
			<-done

		default:
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "Too Many Requests",
			})
			c.Abort()
		}
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// queue setting
	const (
		queueSize   = 10
		workerCount = 1
	)

	requestQueue := make(chan RequestContext, queueSize)
	for i := 0; i < workerCount; i++ {
		go worker(requestQueue)
	}
	router.Use(queueMiddleware(requestQueue))

	router.GET("/", unstableHandlerFactory(false))
	router.GET("/fail", unstableHandlerFactory(true))

	fmt.Println("Generate Simple Http API Server on Port: [8080]")
	router.Run(":8080")

	// test command
	// siege -p -c10 -t3s http://localhost:8080
	// siege -p -c12 -t3s http://localhost:8080
}
