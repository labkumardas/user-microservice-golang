package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"user-microservice-golang/utils"
)

const (
	AuthUserIDKey = "auth_user_id"
	AuthEmailKey  = "auth_email"
	AuthRoleKey   = "auth_role"
)

// Authenticate validates the Bearer JWT and injects claims into the context
func Authenticate(jwtSecret string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			utils.RespondError(c, 401, "missing or malformed authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			logger.Debug("invalid token", zap.Error(err))
			utils.RespondError(c, 401, "invalid or expired token")
			return
		}

		c.Set(AuthUserIDKey, claims.UserID)
		c.Set(AuthEmailKey, claims.Email)
		c.Set(AuthRoleKey, claims.Role)
		c.Next()
	}
}

// RequireRole ensures the authenticated user has one of the allowed roles
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, _ := c.Get(AuthRoleKey)
		roleStr, _ := role.(string)
		if _, ok := allowed[roleStr]; !ok {
			utils.RespondError(c, 403, "insufficient permissions")
			return
		}
		c.Next()
	}
}

// SelfOrAdmin allows access only if the requesting user matches {id} param, or is admin
func SelfOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		authID, _ := c.Get(AuthUserIDKey)
		role, _ := c.Get(AuthRoleKey)

		paramID := c.Param("id")
		if authID.(string) != paramID && role.(string) != "admin" {
			utils.RespondError(c, 403, "access denied")
			return
		}
		c.Next()
	}
}
