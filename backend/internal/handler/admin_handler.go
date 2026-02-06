package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/feedback/internal/models"
	"github.com/minisource/feedback/internal/usecase"
	"github.com/minisource/go-common/response"
)

// AdminHandler handles admin HTTP requests
type AdminHandler struct {
	adminUsecase    *usecase.AdminUsecase
	feedbackUsecase *usecase.FeedbackUsecase
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	adminUsecase *usecase.AdminUsecase,
	feedbackUsecase *usecase.FeedbackUsecase,
) *AdminHandler {
	return &AdminHandler{
		adminUsecase:    adminUsecase,
		feedbackUsecase: feedbackUsecase,
	}
}

// Categories

// CreateCategory creates a new category
// @Summary Create category
// @Tags Admin - Categories
// @Accept json
// @Produce json
// @Param body body models.CreateCategoryRequest true "Create category request"
// @Success 201 {object} models.Category
// @Router /admin/categories [post]
func (h *AdminHandler) CreateCategory(c *fiber.Ctx) error {
	var req models.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	tenantID := c.Locals("tenantId").(string)
	userID := c.Locals("userId").(string)

	category, err := h.adminUsecase.CreateCategory(c.Context(), req, tenantID, userID)
	if err != nil {
		if err == usecase.ErrCategoryNotFound {
			return response.NotFound(c, "Parent category not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, category)
}

// UpdateCategory updates a category
// @Summary Update category
// @Tags Admin - Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param body body models.UpdateCategoryRequest true "Update category request"
// @Success 200 {object} models.Category
// @Router /admin/categories/{id} [put]
func (h *AdminHandler) UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	category, err := h.adminUsecase.UpdateCategory(c.Context(), id, req)
	if err != nil {
		if err == usecase.ErrCategoryNotFound {
			return response.NotFound(c, "Category not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, category)
}

// DeleteCategory deletes a category
// @Summary Delete category
// @Tags Admin - Categories
// @Param id path string true "Category ID"
// @Success 204
// @Router /admin/categories/{id} [delete]
func (h *AdminHandler) DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.adminUsecase.DeleteCategory(c.Context(), id)
	if err != nil {
		if err == usecase.ErrCategoryNotFound {
			return response.NotFound(c, "Category not found")
		}
		return response.InternalError(c, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListCategories lists all categories
// @Summary List categories
// @Tags Categories
// @Produce json
// @Param active_only query bool false "Only active categories"
// @Success 200 {array} models.Category
// @Router /categories [get]
func (h *AdminHandler) ListCategories(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)
	activeOnly := c.QueryBool("active_only", true)

	categories, err := h.adminUsecase.ListCategories(c.Context(), tenantID, activeOnly)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, categories)
}

// GetCategory gets a category by ID
// @Summary Get category
// @Tags Categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} models.Category
// @Router /categories/{id} [get]
func (h *AdminHandler) GetCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	category, err := h.adminUsecase.GetCategory(c.Context(), id)
	if err != nil {
		if err == usecase.ErrCategoryNotFound {
			return response.NotFound(c, "Category not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, category)
}

// Settings

// GetSettings gets all settings
// @Summary Get all settings
// @Tags Admin - Settings
// @Produce json
// @Success 200 {array} models.Setting
// @Router /admin/settings [get]
func (h *AdminHandler) GetSettings(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)

	settings, err := h.adminUsecase.GetAllSettings(c.Context(), tenantID, false)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, settings)
}

// GetPublicSettings gets public settings
// @Summary Get public settings
// @Tags Settings
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /settings [get]
func (h *AdminHandler) GetPublicSettings(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)

	settings, err := h.adminUsecase.GetSettingsAsMap(c.Context(), tenantID, true)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, settings)
}

// UpdateSetting updates a setting
// @Summary Update setting
// @Tags Admin - Settings
// @Accept json
// @Produce json
// @Param body body models.UpdateSettingRequest true "Update setting request"
// @Success 200 {object} map[string]string
// @Router /admin/settings [put]
func (h *AdminHandler) UpdateSetting(c *fiber.Ctx) error {
	var req models.UpdateSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	tenantID := c.Locals("tenantId").(string)
	userID := c.Locals("userId").(string)

	err := h.adminUsecase.SetSetting(c.Context(), req, tenantID, userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Setting updated"})
}

// InitializeSettings initializes default settings
// @Summary Initialize default settings
// @Tags Admin - Settings
// @Success 200 {object} map[string]string
// @Router /admin/settings/initialize [post]
func (h *AdminHandler) InitializeSettings(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)

	err := h.adminUsecase.InitializeSettings(c.Context(), tenantID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Settings initialized"})
}

// Moderation

// ApproveFeedback approves a pending feedback
// @Summary Approve feedback
// @Tags Admin - Moderation
// @Param id path string true "Feedback ID"
// @Success 200 {object} models.Feedback
// @Router /admin/feedback/{id}/approve [post]
func (h *AdminHandler) ApproveFeedback(c *fiber.Ctx) error {
	id := c.Params("id")

	feedback, err := h.adminUsecase.ApproveFeedback(c.Context(), id)
	if err != nil {
		if err == usecase.ErrFeedbackNotFound {
			return response.NotFound(c, "Feedback not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, feedback)
}

// RejectFeedback rejects a feedback
// @Summary Reject feedback
// @Tags Admin - Moderation
// @Param id path string true "Feedback ID"
// @Success 200 {object} models.Feedback
// @Router /admin/feedback/{id}/reject [post]
func (h *AdminHandler) RejectFeedback(c *fiber.Ctx) error {
	id := c.Params("id")

	feedback, err := h.adminUsecase.RejectFeedback(c.Context(), id)
	if err != nil {
		if err == usecase.ErrFeedbackNotFound {
			return response.NotFound(c, "Feedback not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, feedback)
}

// UpdateStatus updates feedback status
// @Summary Update feedback status
// @Tags Admin - Moderation
// @Accept json
// @Produce json
// @Param id path string true "Feedback ID"
// @Param body body map[string]string true "Status update request"
// @Success 200 {object} models.Feedback
// @Router /admin/feedback/{id}/status [put]
func (h *AdminHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	userID := c.Locals("userId").(string)

	feedback, err := h.feedbackUsecase.UpdateStatus(c.Context(), id, models.FeedbackStatus(req.Status), userID)
	if err != nil {
		switch err {
		case usecase.ErrFeedbackNotFound:
			return response.NotFound(c, "Feedback not found")
		case usecase.ErrInvalidStatus:
			return response.BadRequest(c, "BAD_REQUEST", "Invalid status")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return response.OK(c, feedback)
}

// AddOfficialResponse adds an official response to feedback
// @Summary Add official response
// @Tags Admin - Moderation
// @Accept json
// @Produce json
// @Param id path string true "Feedback ID"
// @Param body body map[string]string true "Official response"
// @Success 200 {object} models.Feedback
// @Router /admin/feedback/{id}/response [post]
func (h *AdminHandler) AddOfficialResponse(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	userID := c.Locals("userId").(string)
	userName, _ := c.Locals("userName").(string)

	feedback, err := h.feedbackUsecase.AddOfficialResponse(c.Context(), id, req.Content, userID, userName)
	if err != nil {
		if err == usecase.ErrFeedbackNotFound {
			return response.NotFound(c, "Feedback not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, feedback)
}

// GetPendingFeedback gets all pending feedback
// @Summary Get pending feedback
// @Tags Admin - Moderation
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} map[string]interface{}
// @Router /admin/feedback/pending [get]
func (h *AdminHandler) GetPendingFeedback(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	feedbacks, total, err := h.adminUsecase.GetPendingFeedback(c.Context(), tenantID, page, perPage)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{
		"data":     feedbacks,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// Stats

// GetStats gets dashboard statistics
// @Summary Get dashboard statistics
// @Tags Admin - Stats
// @Produce json
// @Success 200 {object} models.FeedbackStats
// @Router /admin/stats [get]
func (h *AdminHandler) GetStats(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)

	stats, err := h.adminUsecase.GetFeedbackStats(c.Context(), tenantID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, stats)
}
