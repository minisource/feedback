package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/feedback/internal/models"
	"github.com/minisource/feedback/internal/usecase"
	"github.com/minisource/go-common/response"
)

// SubscriptionHandler handles subscription HTTP requests
type SubscriptionHandler struct {
	subscriptionUsecase *usecase.SubscriptionUsecase
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(subscriptionUsecase *usecase.SubscriptionUsecase) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionUsecase: subscriptionUsecase}
}

// Subscribe creates a new subscription
// @Summary Subscribe to feedback/category
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param body body models.SubscribeRequest true "Subscribe request"
// @Success 201 {object} models.Subscription
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Subscribe(c *fiber.Ctx) error {
	var req models.SubscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	tenantID := c.Locals("tenantId").(string)
	userID := c.Locals("userId").(string)
	userEmail, _ := c.Locals("userEmail").(string)

	subscription, err := h.subscriptionUsecase.Subscribe(c.Context(), req, tenantID, userID, userEmail)
	if err != nil {
		switch err {
		case usecase.ErrFeedbackNotFound:
			return response.NotFound(c, "Feedback not found")
		case usecase.ErrCategoryNotFound:
			return response.NotFound(c, "Category not found")
		case usecase.ErrAlreadySubscribed:
			return response.Conflict(c, "Already subscribed")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return response.Created(c, subscription)
}

// Unsubscribe removes a subscription
// @Summary Unsubscribe
// @Tags Subscriptions
// @Param id path string true "Subscription ID"
// @Success 204
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Unsubscribe(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userId").(string)

	err := h.subscriptionUsecase.Unsubscribe(c.Context(), id, userID)
	if err != nil {
		switch err {
		case usecase.ErrSubscriptionNotFound:
			return response.NotFound(c, "Subscription not found")
		case usecase.ErrUnauthorized:
			return response.Forbidden(c, "Not authorized to unsubscribe")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UnsubscribeFromFeedback removes subscription to a specific feedback
// @Summary Unsubscribe from feedback
// @Tags Subscriptions
// @Param feedback_id path string true "Feedback ID"
// @Success 204
// @Router /feedback/{feedback_id}/unsubscribe [delete]
func (h *SubscriptionHandler) UnsubscribeFromFeedback(c *fiber.Ctx) error {
	feedbackID := c.Params("feedback_id")
	userID := c.Locals("userId").(string)

	err := h.subscriptionUsecase.UnsubscribeFromFeedback(c.Context(), feedbackID, userID)
	if err != nil {
		switch err {
		case usecase.ErrFeedbackNotFound:
			return response.NotFound(c, "Feedback not found")
		case usecase.ErrSubscriptionNotFound:
			return response.NotFound(c, "Subscription not found")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListByUser lists all subscriptions for current user
// @Summary List user subscriptions
// @Tags Subscriptions
// @Produce json
// @Success 200 {array} models.SubscriptionResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListByUser(c *fiber.Ctx) error {
	tenantID := c.Locals("tenantId").(string)
	userID := c.Locals("userId").(string)

	subscriptions, err := h.subscriptionUsecase.ListByUser(c.Context(), tenantID, userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, subscriptions)
}

// UpdatePreferences updates subscription preferences
// @Summary Update subscription preferences
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param body body models.UpdateSubscriptionRequest true "Update subscription request"
// @Success 200 {object} models.Subscription
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdatePreferences(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "BAD_REQUEST", "Invalid request body")
	}

	userID := c.Locals("userId").(string)

	subscription, err := h.subscriptionUsecase.UpdatePreferences(c.Context(), id, req, userID)
	if err != nil {
		switch err {
		case usecase.ErrSubscriptionNotFound:
			return response.NotFound(c, "Subscription not found")
		case usecase.ErrUnauthorized:
			return response.Forbidden(c, "Not authorized to update this subscription")
		default:
			return response.InternalError(c, err.Error())
		}
	}

	return response.OK(c, subscription)
}

// CheckSubscription checks if user is subscribed to a feedback
// @Summary Check subscription status
// @Tags Subscriptions
// @Produce json
// @Param feedback_id path string true "Feedback ID"
// @Success 200 {object} map[string]bool
// @Router /feedback/{feedback_id}/subscription [get]
func (h *SubscriptionHandler) CheckSubscription(c *fiber.Ctx) error {
	feedbackID := c.Params("feedback_id")
	userID := c.Locals("userId").(string)

	isSubscribed, err := h.subscriptionUsecase.IsSubscribed(c.Context(), userID, feedbackID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, fiber.Map{"is_subscribed": isSubscribed})
}
