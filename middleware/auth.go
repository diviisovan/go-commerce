package middleware

import (
	"net/http"
	"strings"

	"go-ecommerce/internal/auth"
	"go-ecommerce/models"

	"github.com/gin-gonic/gin"
)

// Keys under which the authenticated identity is stored on the Gin context.
// Handlers must read the user ID from here and never from a request body or
// query parameter, which the caller controls and can forge.
const (
	ContextUserID   = "auth_user_id"
	ContextUserRole = "auth_user_role"
)

// RequireAuth rejects any request without a valid Bearer access token.
func RequireAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			unauthorized(c, "authorization header required")
			return
		}

		// Expect exactly "Bearer <token>".
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c, "authorization header must be in the form: Bearer <token>")
			return
		}

		claims, err := tokens.ParseAccessToken(parts[1])
		if err != nil {
			unauthorized(c, "invalid or expired token")
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}

// RequireRole allows the request only if the authenticated user holds one of the
// given roles. Chain it after RequireAuth.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role := UserRole(c)
		if _, ok := allowed[role]; !ok {
			// 403, not 404: the caller is authenticated, just not permitted.
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error: "insufficient permissions for this resource",
			})
			return
		}
		c.Next()
	}
}

// UserID returns the authenticated user's ID from the context.
func UserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get(ContextUserID)
	if !exists {
		return 0, false
	}
	id, ok := value.(uint)
	return id, ok
}

// UserRole returns the authenticated user's role, or "" when unauthenticated.
func UserRole(c *gin.Context) string {
	value, exists := c.Get(ContextUserRole)
	if !exists {
		return ""
	}
	role, _ := value.(string)
	return role
}

// unauthorized aborts with 401 and the WWW-Authenticate header the HTTP spec
// requires on that status.
func unauthorized(c *gin.Context, message string) {
	c.Header("WWW-Authenticate", `Bearer realm="api"`)
	c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Error: message})
}
