package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"digital-signature-system/internal/config"
)

type Server struct {
	config *config.Config
	db     *gorm.DB
	router *gin.Engine
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	server := &Server{
		config: cfg,
		db:     db,
		router: gin.Default(),
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes will be added in subsequent tasks
	api := s.router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}