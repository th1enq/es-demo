package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type HTTPServer interface {
	Start(ctx context.Context) error
}

type Config struct {
	Host string
	Port int
}

type httpServer struct {
	cfg        Config
	controller *Controller
	logger     *zap.Logger
}

func NewHTTPServer(
	cfg Config,
	controller *Controller,
	logger *zap.Logger,
) HTTPServer {
	return &httpServer{
		cfg:        cfg,
		controller: controller,
		logger:     logger,
	}
}

func (s *httpServer) RegisRouter() *gin.Engine {
	router := gin.Default()

	// CORS middleware for frontend
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := router.Group("/api/v1")
	{

		// Public bank account creation (for registration)
		apiV1.POST("/bank_accounts", s.controller.CreateBankAccount)

		// Public bank account lookup (for testing)
		apiV1.GET("/bank_accounts/:id", s.controller.GetBankAccountByID)

		// Events history endpoint
		apiV1.GET("/bank_accounts/:id/events", s.controller.GetEventsHistory)

		// Temporarily make all operations public for testing
		apiV1.POST("/bank_accounts/:id/deposite", s.controller.DepositeBalance)
		apiV1.POST("/bank_accounts/:id/withdraw", s.controller.WithdrawBalance)

	}

	return router
}

func (s *httpServer) Start(ctx context.Context) error {
	s.logger.Info("Starting HTTP Server", zap.String("host", s.cfg.Host), zap.Int("port", s.cfg.Port))

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Handler: s.RegisRouter(),
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("HTTP Server failed", zap.Error(err))
	}
	return nil
}
