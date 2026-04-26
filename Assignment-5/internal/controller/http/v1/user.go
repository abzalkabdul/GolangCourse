package v1

import (
	"net/http"

	"assignment_5/internal/entity"
	"assignment_5/internal/usecase"
	"assignment_5/pkg/logger"
	"assignment_5/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userRoutes struct {
	t usecase.UserInterface
	l logger.Interface
}

func newUserRoutes(handler *gin.RouterGroup, t usecase.UserInterface, l logger.Interface, rl *utils.RateLimiter) {
	r := &userRoutes{t, l}

	h := handler.Group("/users")
	h.Use(utils.RateLimiterMiddleware(rl))
	{
		h.POST("/", r.RegisterUser)
		h.POST("/login", r.LoginUser)

		protected := h.Group("/")
		protected.Use(utils.JWTAuthMiddleware())
		{
			protected.GET("/me", r.GetMe)
			protected.GET("/protected/hello", r.ProtectedFunc)

			adminRoutes := protected.Group("/")
			adminRoutes.Use(utils.RoleMiddleware("admin"))
			{
				adminRoutes.PATCH("/promote/:id", r.PromoteUser)
			}
		}
	}
}

func (r *userRoutes) RegisterUser(c *gin.Context) {
	var dto entity.CreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	role := "user"
	if dto.Role != "" {
		role = dto.Role
	}

	user := entity.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: hashedPassword,
		Role:     role,
	}

	createdUser, sessionID, err := r.t.RegisterUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "User registered successfully. Please check your email for verification code.",
		"session_id": sessionID,
		"user":       createdUser,
	})
}

func (r *userRoutes) LoginUser(c *gin.Context) {
	var input entity.LoginUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := r.t.LoginUser(&input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetMe returns the authenticated user's profile.
// Requires: valid JWT in Authorization header. No other input needed.
func (r *userRoutes) GetMe(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID in token"})
		return
	}

	user, err := r.t.GetMe(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// PromoteUser promotes a user to admin role.
// Requires: admin JWT token. Target user ID is in the URL path.
func (r *userRoutes) PromoteUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := r.t.PromoteUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User successfully promoted to admin",
		"user":    user,
	})
}

func (r *userRoutes) ProtectedFunc(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
