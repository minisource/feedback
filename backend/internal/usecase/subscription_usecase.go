package usecase

import (
	"context"
	"errors"

	"github.com/minisource/feedback/internal/models"
	"github.com/minisource/feedback/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrAlreadySubscribed    = errors.New("already subscribed")
)

// SubscriptionUsecase handles subscription business logic
type SubscriptionUsecase struct {
	subscriptionRepo *repository.SubscriptionRepository
	feedbackRepo     *repository.FeedbackRepository
	categoryRepo     *repository.CategoryRepository
}

// NewSubscriptionUsecase creates a new subscription usecase
func NewSubscriptionUsecase(
	subscriptionRepo *repository.SubscriptionRepository,
	feedbackRepo *repository.FeedbackRepository,
	categoryRepo *repository.CategoryRepository,
) *SubscriptionUsecase {
	return &SubscriptionUsecase{
		subscriptionRepo: subscriptionRepo,
		feedbackRepo:     feedbackRepo,
		categoryRepo:     categoryRepo,
	}
}

// Subscribe creates a new subscription
func (u *SubscriptionUsecase) Subscribe(ctx context.Context, req models.SubscribeRequest, tenantID, userID, userEmail string) (*models.Subscription, error) {
	sub := &models.Subscription{
		TenantID:     tenantID,
		UserID:       userID,
		UserEmail:    userEmail,
		Types:        req.Types,
		SubscribeAll: req.SubscribeAll,
		EmailEnabled: req.EmailEnabled,
		PushEnabled:  req.PushEnabled,
	}

	// Validate and set feedback ID
	if req.FeedbackID != "" {
		feedbackID, err := primitive.ObjectIDFromHex(req.FeedbackID)
		if err != nil {
			return nil, ErrFeedbackNotFound
		}

		feedback, err := u.feedbackRepo.GetByID(ctx, feedbackID)
		if err != nil {
			return nil, err
		}
		if feedback == nil {
			return nil, ErrFeedbackNotFound
		}

		// Check if already subscribed
		existing, _ := u.subscriptionRepo.GetByUserAndFeedback(ctx, userID, feedbackID)
		if existing != nil {
			return nil, ErrAlreadySubscribed
		}

		sub.FeedbackID = &feedbackID
	}

	// Validate and set category ID
	if req.CategoryID != "" {
		categoryID, err := primitive.ObjectIDFromHex(req.CategoryID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}

		category, err := u.categoryRepo.GetByID(ctx, categoryID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, ErrCategoryNotFound
		}

		sub.CategoryID = &categoryID
	}

	// Set default types if not specified
	if len(sub.Types) == 0 {
		sub.Types = []models.SubscriptionType{models.SubAll}
	}

	if err := u.subscriptionRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// Unsubscribe removes a subscription
func (u *SubscriptionUsecase) Unsubscribe(ctx context.Context, id string, userID string) error {
	subID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrSubscriptionNotFound
	}

	sub, err := u.subscriptionRepo.GetByID(ctx, subID)
	if err != nil {
		return err
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	// Verify ownership
	if sub.UserID != userID {
		return ErrUnauthorized
	}

	return u.subscriptionRepo.Unsubscribe(ctx, subID)
}

// UnsubscribeFromFeedback removes subscription to a specific feedback
func (u *SubscriptionUsecase) UnsubscribeFromFeedback(ctx context.Context, feedbackIDStr, userID string) error {
	feedbackID, err := primitive.ObjectIDFromHex(feedbackIDStr)
	if err != nil {
		return ErrFeedbackNotFound
	}

	sub, err := u.subscriptionRepo.GetByUserAndFeedback(ctx, userID, feedbackID)
	if err != nil {
		return err
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}

	return u.subscriptionRepo.Unsubscribe(ctx, sub.ID)
}

// ListByUser lists all subscriptions for a user
func (u *SubscriptionUsecase) ListByUser(ctx context.Context, tenantID, userID string) ([]models.SubscriptionResponse, error) {
	subs, err := u.subscriptionRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]models.SubscriptionResponse, len(subs))
	for i, sub := range subs {
		responses[i] = *u.toResponse(&sub)
	}

	return responses, nil
}

// GetSubscribersForNotification gets subscribers who should be notified
func (u *SubscriptionUsecase) GetSubscribersForNotification(ctx context.Context, feedbackID primitive.ObjectID, eventType models.SubscriptionType) ([]models.Subscription, error) {
	return u.subscriptionRepo.GetSubscribersForFeedback(ctx, feedbackID, eventType)
}

// IsSubscribed checks if a user is subscribed to a feedback
func (u *SubscriptionUsecase) IsSubscribed(ctx context.Context, userID, feedbackIDStr string) (bool, error) {
	feedbackID, err := primitive.ObjectIDFromHex(feedbackIDStr)
	if err != nil {
		return false, nil
	}

	return u.subscriptionRepo.IsSubscribed(ctx, userID, feedbackID)
}

// UpdatePreferences updates subscription preferences
func (u *SubscriptionUsecase) UpdatePreferences(ctx context.Context, id string, req models.UpdateSubscriptionRequest, userID string) (*models.Subscription, error) {
	subID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}

	sub, err := u.subscriptionRepo.GetByID(ctx, subID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}

	// Verify ownership
	if sub.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Update fields
	if len(req.Types) > 0 {
		sub.Types = req.Types
	}
	if req.EmailEnabled != nil {
		sub.EmailEnabled = *req.EmailEnabled
	}
	if req.PushEnabled != nil {
		sub.PushEnabled = *req.PushEnabled
	}

	if err := u.subscriptionRepo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// toResponse converts Subscription to SubscriptionResponse
func (u *SubscriptionUsecase) toResponse(s *models.Subscription) *models.SubscriptionResponse {
	resp := &models.SubscriptionResponse{
		ID:           s.ID.Hex(),
		Types:        s.Types,
		SubscribeAll: s.SubscribeAll,
		EmailEnabled: s.EmailEnabled,
		PushEnabled:  s.PushEnabled,
		CreatedAt:    s.CreatedAt,
	}

	if s.FeedbackID != nil {
		resp.FeedbackID = s.FeedbackID.Hex()
	}
	if s.CategoryID != nil {
		resp.CategoryID = s.CategoryID.Hex()
	}

	return resp
}
