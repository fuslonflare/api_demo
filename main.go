package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/fuslonflare/api_demo/auth"
	"github.com/fuslonflare/api_demo/todo"
)

var (
	buildCommit = "dev"
	buildTime   = time.Now().String()
)

func main() {
	_, err := os.Create("tmp/live")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("tmp/live")

	err = godotenv.Load("local.env")
	if err != nil {
		log.Printf("Please consider environment variables: %s\n", err)
	}

	db, err := gorm.Open(mysql.Open(os.Getenv("DB_CONN")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database.")
	}

	db.AutoMigrate(&todo.Todo{})

	engine := gin.Default()

	// handle-cors
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:8080",
	}
	config.AllowHeaders = []string{
		"Origin",
		"Authorization",
		"TransactionId",
	}
	engine.Use(cors.New(config))

	engine.GET("/healthz", func(context *gin.Context) {
		context.Status(200)
	})

	engine.GET("limitz", limitedHandler)
	engine.GET("/x", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"buildCommit": buildCommit,
			"buildTime":   buildTime,
		})
	})

	engine.GET("/tokenz", auth.AccessToken(os.Getenv("SIGN")))
	protected := engine.Group("", auth.Protect([]byte(os.Getenv("SIGN"))))

	handler := todo.NewTodoHandler(db)
	protected.POST("/todos", handler.NewTask)
	protected.GET("/todos", handler.List)
	protected.DELETE("/todos/:id", handler.Remove)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := &http.Server{
		Addr:           ":" + os.Getenv("PORT"),
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()
	fmt.Println("shutting down gracefully, press ctrl+c again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}

var limiter = rate.NewLimiter(5, 5)

func limitedHandler(context *gin.Context) {
	if !limiter.Allow() {
		context.AbortWithStatus(http.StatusTooManyRequests)
		return
	}

	context.JSON(200, gin.H{
		"message": "dr.pong",
	})
}
