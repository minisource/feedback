package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-sdk/auth"
)

// AuthConfig holds auth middleware configuration
type AuthConfig struct {
	AuthClient   *auth.Client
	SkipPaths    []string
	RequireAdmin []string
}

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(cfg AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip auth for certain paths
		path := c.Path()
		for _, skipPath := range cfg.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// Allow optional auth - set defaults and continue
			c.Locals("userId", "")
			c.Locals("userName", "")
			c.Locals("userEmail", "")
			c.Locals("isAdmin", false)
			return c.Next()
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
		}

		// Validate token
		result, err := cfg.AuthClient.ValidateToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Failed to validate token",
			})
		}

		if !result.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Token is not valid",
			})
		}

		// Check admin requirement for certain paths
		isAdmin := hasAdminScope(result.Scopes)
		for _, adminPath := range cfg.RequireAdmin {
			if strings.HasPrefix(path, adminPath) {
				if !isAdmin {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error":   "forbidden",
						"message": "Admin access required",
					})
				}
				break
			}
		}

		// Set user info in context
		c.Locals("userId", result.ClientID)
		c.Locals("userName", result.ServiceName)
		c.Locals("userEmail", "")
		c.Locals("isAdmin", isAdmin)
		c.Locals("scopes", result.Scopes)

		return c.Next()
	}
}

// RequireAuthMiddleware requires authentication
func RequireAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userId").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
		}
		return c.Next()
	}
}

// hasAdminScope checks if scopes include admin
func hasAdminScope(scopes []string) bool {
	for _, scope := range scopes {
		if scope == "feedback:admin" || scope == "admin" {
			return true
		}
	}
	return false
}
