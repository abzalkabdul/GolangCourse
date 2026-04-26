package v1

import (
	"time"

	"assignment_5/internal/usecase"
	"assignment_5/pkg/logger"
	"assignment_5/utils"

	"github.com/gin-gonic/gin"
)

// NewRouter registers all routes on the provided gin.Engine.
// Rate limiter: 10 requests per minute per user/IP.
func NewRouter(handler *gin.Engine, t usecase.UserInterface, l logger.Interface) {
	rateLimiter := utils.NewRateLimiter(10, time.Minute)

	handler.Use(gin.Recovery())
	handler.Use(gin.Logger())

	v1 := handler.Group("/api/v1")
	{
		newUserRoutes(v1, t, l, rateLimiter)
	}
}
