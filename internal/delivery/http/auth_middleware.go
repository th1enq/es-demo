package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/es-demo/internal/dto"
	"github.com/th1enq/es-demo/internal/service"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	authService service.AuthService
	logger      *zap.Logger
}

func NewAuthMiddleware(authService service.AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// JWTAuth middleware validates JWT tokens
func (m *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing authorization header")
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				dto.CodeUnauthorized,
				"authorization header required",
				"missing authorization header",
			))
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			m.logger.Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				dto.CodeUnauthorized,
				"invalid authorization header format",
				"format should be 'Bearer <token>'",
			))
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate token
		token, err := m.authService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			m.logger.Warn("Invalid or expired token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				dto.CodeInvalidToken,
				"invalid or expired token",
				err.Error(),
			))
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*service.CustomClaims)
		if !ok {
			m.logger.Error("Failed to extract token claims")
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
				dto.CodeInvalidToken,
				"invalid token claims",
				"failed to parse token claims",
			))
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		m.logger.Debug("JWT authentication successful",
			zap.String("user_id", claims.UserID),
			zap.String("email", claims.Email),
		)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			m.logger.Error("User role not found in context")
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(
				dto.CodeForbidden,
				"access denied",
				"user role not available",
			))
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok || role != requiredRole {
			userID, _ := c.Get("user_id")
			m.logger.Warn("Insufficient permissions",
				zap.Any("user_id", userID),
				zap.String("required_role", requiredRole),
				zap.Any("user_role", userRole),
			)
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(
				dto.CodeForbidden,
				"insufficient permissions",
				"required role: "+requiredRole,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware validates JWT token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		// Validate token
		token, err := m.authService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*service.CustomClaims)
		if ok {
			// Set user context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)
		}

		c.Next()
	}
}
