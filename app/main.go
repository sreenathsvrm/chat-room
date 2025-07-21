package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sreenathsvrm/chat-room/app/internal/chat"
	"github.com/sreenathsvrm/chat-room/app/internal/config"
	"github.com/sreenathsvrm/chat-room/app/internal/models"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Verify DB connection
	sqldb, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database object:", err)
	}
	if err := sqldb.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(&models.Message{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize chat room
	messageRepo := chat.NewMessageRepository(db)
	chatRoom := chat.NewChatRoom(messageRepo, 1000)
	chatRoom.Run()

	// Set up Gin router
	router := gin.Default()

	// Custom logger middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s\" %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.ErrorMessage,
		)
	}))
	// Recovery middleware
	router.Use(gin.Recovery())

	// API routes
	api := router.Group("/api")
	{
		api.POST("/join", joinHandler(chatRoom))
		api.POST("/leave", leaveHandler(chatRoom))
		api.POST("/send", sendHandler(chatRoom))
		api.GET("/messages", messagesHandler(chatRoom))
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server closed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Forced shutdown:", err)
	}
}

func joinHandler(room *chat.ChatRoom) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ClientID string `json:"client_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		client, err := room.Join(req.ClientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"client_id": client.ID,
		})
	}
}

func leaveHandler(room *chat.ChatRoom) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ClientID string `json:"client_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		room.Leave(req.ClientID)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func sendHandler(room *chat.ChatRoom) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ClientID string `json:"client_id" binding:"required"`
			Message  string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := room.Send(req.ClientID, req.Message)
		if err != nil {
			if err.Error() == "client not found" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Please join the room before sending messages."})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func messagesHandler(room *chat.ChatRoom) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.Query("client_id")
		if clientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
			return
		}

		sinceStr := c.Query("since")
		var since time.Time
		if sinceStr != "" {
			unixTime, err := strconv.ParseInt(sinceStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid since parameter"})
				return
			}
			since = time.Unix(unixTime, 0)
		} else {
			since = time.Time{} // zero time, will return all messages
		}

		// Long polling with timeout
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				c.JSON(http.StatusOK, gin.H{"messages": []string{}})
				return
			case <-ticker.C:
				messages, err := room.GetMessages(clientID, since)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if len(messages) > 0 {
					c.JSON(http.StatusOK, gin.H{"messages": messages})
					return
				}
			}
		}
	}
}
