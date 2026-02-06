package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/response"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns health status
// @Summary Health check
// @Tags Health
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"status":  "healthy",
		"service": "feedback",
	})
}

// Ready returns readiness status
// @Summary Readiness check
// @Tags Health
// @Success 200 {object} map[string]string
// @Router /ready [get]
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"status": "ready",
	})
}
