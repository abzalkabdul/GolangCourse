package utils

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ----- JWT Secret (lazy-loaded) -----

var (
	jwtSecretOnce sync.Once
	jwtSecret     []byte
)

func getJWTSecret() []byte {
	jwtSecretOnce.Do(func() {
		jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	})
	return jwtSecret
}

// ----- Password helpers -----

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// ----- JWT -----

func GenerateJWT(userID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// JWTAuthMiddleware requires a valid JWT token; aborts with 401 if missing/invalid.
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return getJWTSecret(), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		claims, _ := token.Claims.(jwt.MapClaims)
		c.Set("userID", claims["user_id"].(string))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}

// ----- Role Middleware -----

// RoleMiddleware checks that the JWT role claim matches requiredRole.
// Must be used after JWTAuthMiddleware.
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No role information found"})
			return
		}
		if role.(string) != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		c.Next()
	}
}

// ----- Rate Limiter -----

type visitorInfo struct {
	count   int
	resetAt time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitorInfo
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitorInfo),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]
	if !exists || now.After(v.resetAt) {
		rl.visitors[key] = &visitorInfo{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	if v.count >= rl.limit {
		return false
	}
	v.count++
	return true
}

// RateLimiterMiddleware limits requests per user (JWT userID) or per IP (anonymous).
// Authenticated identification: UserID from JWT claim.
// Anonymous identification: ClientIP from Gin context.
func RateLimiterMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string

		// Check if JWTAuthMiddleware already set the userID
		if userID, exists := c.Get("userID"); exists {
			key = "user:" + userID.(string)
		} else {
			// Try to parse the token without requiring it
			tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
			if tokenStr != "" {
				token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
					return getJWTSecret(), nil
				})
				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if uid, ok := claims["user_id"].(string); ok {
							key = "user:" + uid
						}
					}
				}
			}
			if key == "" {
				key = "ip:" + c.ClientIP()
			}
		}

		if !rl.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}
		c.Next()
	}
}
