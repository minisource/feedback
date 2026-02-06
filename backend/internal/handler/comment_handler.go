package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/feedback/internal/models"
	"github.com/minisource/go-common/response"
	"github.com/minisource/go-sdk/comment"
)

// CommentHandler handles comment HTTP requests using comment microservice
type CommentHandler struct {
	commentClient *comment.Client
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentClient *comment.Client) *CommentHandler {
	return &CommentHandler{commentClient: commentClient}
}

// Create creates a new comment on a feedback
// @Summary Create comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param feedback_id path string true "Feedback ID"
// @Param body body models.CreateCommentRequest true "Create comment request"
// @Success 201 {object} comment.Comment
// @Router /feedback/{feedback_id}/comments [post]
func (h *CommentHandler) Create(c *fiber.Ctx) error {
	feedbackID := c.Params("feedback_id")

	var req models.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	tenantID := c.Locals("tenantId").(string)
	userName, _ := c.Locals("userName").(string)

	// Create comment via comment microservice
	createReq := comment.CreateCommentRequest{
		TenantID:     tenantID,
		ResourceType: "feedback",
		ResourceID:   feedbackID,
		ParentID:     req.ParentID,
		Content:      req.Content,
		AuthorName:   userName,
		IsAnonymous:  req.IsAnonymous,
	}

	result, err := h.commentClient.Create(c.Context(), createReq)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, result)
}

// List lists comments for a feedback
// @Summary List comments
// @Tags Comments
// @Produce json
// @Param feedback_id path string true "Feedback ID"
// @Param parent_id query string false "Parent comment ID"
// @Param sort_by query string false "Sort by (created_at, like_count)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} comment.ListCommentsResponse
// @Router /feedback/{feedback_id}/comments [get]
func (h *CommentHandler) List(c *fiber.Ctx) error {
	feedbackID := c.Params("feedback_id")
	tenantID := c.Locals("tenantId").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	listReq := comment.ListCommentsRequest{
		TenantID:     tenantID,
		ResourceType: "feedback",
		ResourceID:   feedbackID,
		ParentID:     c.Query("parent_id"),
		SortBy:       c.Query("sort_by", "created_at"),
		SortOrder:    c.Query("sort_order", "desc"),
		Page:         page,
		PageSize:     pageSize,
	}

	result, err := h.commentClient.List(c.Context(), listReq)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, result)
}

// GetByID gets a comment by ID
// @Summary Get comment by ID
// @Tags Comments
// @Produce json
// @Param id path string true "Comment ID"
// @Success 200 {object} comment.Comment
// @Router /comments/{id} [get]
func (h *CommentHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := h.commentClient.Get(c.Context(), id)
	if err != nil {
		if err == comment.ErrCommentNotFound {
			return response.NotFound(c, "Comment not found")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, result)
}

// GetReplies gets replies to a comment
// @Summary Get replies to a comment
// @Tags Comments
// @Produce json
// @Param id path string true "Comment ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} comment.RepliesResponse
// @Router /comments/{id}/replies [get]
func (h *CommentHandler) GetReplies(c *fiber.Ctx) error {
	id := c.Params("id")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	result, err := h.commentClient.GetReplies(c.Context(), id, page, pageSize)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, result)
}

// Update updates a comment
// @Summary Update comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param body body models.UpdateCommentRequest true "Update comment request"
// @Success 200 {object} comment.Comment
// @Router /comments/{id} [put]
func (h *CommentHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	updateReq := comment.UpdateCommentRequest{
		Content: req.Content,
	}

	result, err := h.commentClient.Update(c.Context(), id, updateReq)
	if err != nil {
		if err == comment.ErrCommentNotFound {
			return response.NotFound(c, "Comment not found")
		}
		if err == comment.ErrUnauthorized {
			return response.Forbidden(c, "Not authorized to update this comment")
		}
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, result)
}

// Delete deletes a comment
// @Summary Delete comment
// @Tags Comments
// @Param id path string true "Comment ID"
// @Success 204
// @Router /comments/{id} [delete]
func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.commentClient.Delete(c.Context(), id)
	if err != nil {
		if err == comment.ErrCommentNotFound {
			return response.NotFound(c, "Comment not found")
		}
		if err == comment.ErrUnauthorized {
			return response.Forbidden(c, "Not authorized to delete this comment")
		}
		return response.InternalError(c, err.Error())
	}

	return response.NoContent(c)
}

// AddReaction adds a reaction to a comment
// @Summary Add reaction to comment
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param body body map[string]string true "Reaction request"
// @Success 200 {object} map[string]string
// @Router /comments/{id}/reactions [post]
func (h *CommentHandler) AddReaction(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Type string `json:"type"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	err := h.commentClient.AddReaction(c.Context(), id, comment.ReactionType(req.Type))
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"message": "Reaction added"})
}

// RemoveReaction removes a reaction from a comment
// @Summary Remove reaction from comment
// @Tags Comments
// @Param id path string true "Comment ID"
// @Success 204
// @Router /comments/{id}/reactions [delete]
func (h *CommentHandler) RemoveReaction(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.commentClient.RemoveReaction(c.Context(), id)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.NoContent(c)
}

// GetStats gets comment statistics for a feedback
// @Summary Get comment statistics
// @Tags Comments
// @Produce json
// @Param feedback_id path string true "Feedback ID"
// @Success 200 {object} comment.CommentStats
// @Router /feedback/{feedback_id}/comments/stats [get]
func (h *CommentHandler) GetStats(c *fiber.Ctx) error {
	feedbackID := c.Params("feedback_id")
	tenantID := c.Locals("tenantId").(string)

	stats, err := h.commentClient.GetStats(c.Context(), tenantID, "feedback", feedbackID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, stats)
}
