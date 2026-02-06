package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/feedback/internal/models"
	"github.com/minisource/feedback/internal/usecase"
	"github.com/minisource/go-common/response"
)

// FeedbackHandler handles feedback HTTP requests
type FeedbackHandler struct {
	feedbackUsecase *usecase.FeedbackUsecase
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(feedbackUsecase *usecase.FeedbackUsecase) *FeedbackHandler {
	return &FeedbackHandler{feedbackUsecase: feedbackUsecase}
}

// Create creates a new feedback
// @Summary Create feedback
// @Tags Feedback
// @Accept json
// @Produce json
// @Param body body models.CreateFeedbackRequest true "Create feedback request"
// @Success 201 {object} models.Feedback
// @Router /feedback [post]
func (h *FeedbackHandler) Create(c *fiber.Ctx) error {
	var req models.CreateFeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	tenantID := c.Locals("tenantId").(string)
	userID, _ := c.Locals("userId").(string)
	userName, _ := c.Locals("userName").(string)
	userEmail, _ := c.Locals("userEmail").(string)

	feedback, err := h.feedbackUsecase.Create(c.Context(), req, tenantID, userID, userName, userEmail)
	if err != nil {
		switch err {
		case usecase.ErrAnonymousDisabled:
			return response.BadRequest(c, "BAD_REQUEST", "Anonymous feedback is disabled")
		case usecase.ErrCategoryNotFound:
			return response.NotFound(c, "Category not found")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return response.Created(c, feedback)
}

// GetByID gets feedback by ID
// @Summary Get feedback by ID
// @Tags Feedback
// @Produce json
// @Param id path string true "Feedback ID"
// @Success 200 {object} models.FeedbackResponse
// @Router /feedback/{id} [get]
func (h *FeedbackHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	feedback, err := h.feedbackUsecase.GetByID(c.Context(), id, userID)
	if err != nil {
		if err == usecase.ErrFeedbackNotFound {
			return response.NotFound(c, "Feedback not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, feedback)
}

// List lists feedback with filtering
// @Summary List feedback
// @Tags Feedback
// @Produce json
// @Param status query []string false "Filter by status"
// @Param category_id query string false "Filter by category"
// @Param tags query []string false "Filter by tags"
// @Param search query string false "Search text"
// @Param sort_by query string false "Sort by (new, top, trending, most_commented)"
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Success 200 {object} map[string]interface{}
// @Router /feedback [get]
func (h *FeedbackHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)
	userID, _ := c.Locals("userId").(string)

	filter := models.FeedbackFilter{
		TenantID:   tenantID,
		CategoryID: c.Query("category_id"),
		AuthorID:   c.Query("author_id"),
		Search:     c.Query("search"),
		SortBy:     c.Query("sort_by", "new"),
		SortOrder:  c.Query("sort_order", "desc"),
		Page:       c.QueryInt("page", 1),
		PerPage:    c.QueryInt("per_page", 20),
	}

	// Parse status filter
	if statuses := c.Query("status"); statuses != "" {
		filter.Status = []models.FeedbackStatus{models.FeedbackStatus(statuses)}
	}

	feedbacks, total, err := h.feedbackUsecase.List(c.Context(), filter, userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{
		"data":     feedbacks,
		"total":    total,
		"page":     filter.Page,
		"per_page": filter.PerPage,
	})
}

// Update updates a feedback
// @Summary Update feedback
// @Tags Feedback
// @Accept json
// @Produce json
// @Param id path string true "Feedback ID"
// @Param body body models.UpdateFeedbackRequest true "Update feedback request"
// @Success 200 {object} models.Feedback
// @Router /feedback/{id} [put]
func (h *FeedbackHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateFeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	userID := c.Locals("userId").(string)
	isAdmin, _ := c.Locals("isAdmin").(bool)

	feedback, err := h.feedbackUsecase.Update(c.Context(), id, req, userID, isAdmin)
	if err != nil {
		switch err {
		case usecase.ErrFeedbackNotFound:
			return response.NotFound(c, "Feedback not found")
		case usecase.ErrUnauthorized:
			return response.Forbidden(c, "Not authorized to update this feedback")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return response.OK(c, feedback)
}

// Delete deletes a feedback
// @Summary Delete feedback
// @Tags Feedback
// @Param id path string true "Feedback ID"
// @Success 204
// @Router /feedback/{id} [delete]
func (h *FeedbackHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userId").(string)
	isAdmin, _ := c.Locals("isAdmin").(bool)

	err := h.feedbackUsecase.Delete(c.Context(), id, userID, isAdmin)
	if err != nil {
		switch err {
		case usecase.ErrFeedbackNotFound:
			return response.NotFound(c, "Feedback not found")
		case usecase.ErrUnauthorized:
			return response.Forbidden(c, "Not authorized to delete this feedback")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Vote votes on a feedback
// @Summary Vote on feedback
// @Tags Feedback
// @Accept json
// @Produce json
// @Param id path string true "Feedback ID"
// @Param body body models.VoteRequest true "Vote request"
// @Success 200 {object} map[string]string
// @Router /feedback/{id}/vote [post]
func (h *FeedbackHandler) Vote(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.VoteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	tenantID := c.Locals("tenantId").(string)
	userID := c.Locals("userId").(string)

	err := h.feedbackUsecase.Vote(c.Context(), id, models.VoteType(req.VoteType), tenantID, userID)
	if err != nil {
		switch err {
		case usecase.ErrFeedbackNotFound:
			return response.NotFound(c, "Feedback not found")
		case usecase.ErrInvalidVoteType:
			return response.BadRequest(c, "BAD_REQUEST", "Invalid vote type")
		case usecase.ErrVotingDisabled:
			return response.BadRequest(c, "BAD_REQUEST", "Voting is disabled")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return response.OK(c, fiber.Map{"message": "Vote recorded"})
}

// GetTrending gets trending feedback
// @Summary Get trending feedback
// @Tags Feedback
// @Produce json
// @Param limit query int false "Limit"
// @Success 200 {array} models.FeedbackResponse
// @Router /feedback/trending [get]
func (h *FeedbackHandler) GetTrending(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)
	limit := c.QueryInt("limit", 10)

	feedbacks, err := h.feedbackUsecase.GetTrending(c.Context(), tenantID, limit)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, feedbacks)
}

// GetStats gets feedback statistics
// @Summary Get feedback statistics
// @Tags Feedback
// @Produce json
// @Success 200 {object} models.FeedbackStats
// @Router /feedback/stats [get]
func (h *FeedbackHandler) GetStats(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)

	stats, err := h.feedbackUsecase.GetStats(c.Context(), tenantID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, stats)
}
