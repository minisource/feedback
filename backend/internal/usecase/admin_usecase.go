package usecase

import (
	"context"

	"github.com/minisource/feedback/internal/models"
	"github.com/minisource/feedback/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminUsecase handles admin business logic
type AdminUsecase struct {
	feedbackRepo *repository.FeedbackRepository
	categoryRepo *repository.CategoryRepository
	settingRepo  *repository.SettingRepository
}

// NewAdminUsecase creates a new admin usecase
func NewAdminUsecase(
	feedbackRepo *repository.FeedbackRepository,
	categoryRepo *repository.CategoryRepository,
	settingRepo *repository.SettingRepository,
) *AdminUsecase {
	return &AdminUsecase{
		feedbackRepo: feedbackRepo,
		categoryRepo: categoryRepo,
		settingRepo:  settingRepo,
	}
}

// Categories

// CreateCategory creates a new category
func (u *AdminUsecase) CreateCategory(ctx context.Context, req models.CreateCategoryRequest, tenantID, userID string) (*models.Category, error) {
	category := &models.Category{
		TenantID:    tenantID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		IsActive:    true,
	}

	// Parse parent ID if provided
	if req.ParentID != "" {
		parentID, err := primitive.ObjectIDFromHex(req.ParentID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		parent, err := u.categoryRepo.GetByID(ctx, parentID)
		if err != nil || parent == nil {
			return nil, ErrCategoryNotFound
		}
		category.ParentID = &parentID
	}

	if err := u.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// UpdateCategory updates a category
func (u *AdminUsecase) UpdateCategory(ctx context.Context, id string, req models.UpdateCategoryRequest) (*models.Category, error) {
	categoryID, err := primitive.ObjectIDFromHex(id)
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

	if req.Name != nil && *req.Name != "" {
		category.Name = *req.Name
	}
	if req.Slug != nil && *req.Slug != "" {
		category.Slug = *req.Slug
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.Color != nil {
		category.Color = *req.Color
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := u.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category
func (u *AdminUsecase) DeleteCategory(ctx context.Context, id string) error {
	categoryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrCategoryNotFound
	}

	category, err := u.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return ErrCategoryNotFound
	}

	return u.categoryRepo.Delete(ctx, categoryID)
}

// ListCategories lists all categories
func (u *AdminUsecase) ListCategories(ctx context.Context, tenantID string, activeOnly bool) ([]models.Category, error) {
	return u.categoryRepo.List(ctx, tenantID, activeOnly)
}

// GetCategory gets a category by ID
func (u *AdminUsecase) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	categoryID, err := primitive.ObjectIDFromHex(id)
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

	return category, nil
}

// Settings

// GetSetting gets a setting value
func (u *AdminUsecase) GetSetting(ctx context.Context, tenantID, key string) (*models.Setting, error) {
	return u.settingRepo.Get(ctx, tenantID, key)
}

// SetSetting sets a setting value
func (u *AdminUsecase) SetSetting(ctx context.Context, req models.UpdateSettingRequest, tenantID, userID string) error {
	setting := &models.Setting{
		TenantID:    tenantID,
		Key:         req.Key,
		Value:       req.Value,
		Type:        models.SettingType(req.Type),
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UpdatedBy:   userID,
	}

	return u.settingRepo.Set(ctx, setting)
}

// GetAllSettings gets all settings for a tenant
func (u *AdminUsecase) GetAllSettings(ctx context.Context, tenantID string, publicOnly bool) ([]models.Setting, error) {
	return u.settingRepo.GetAll(ctx, tenantID, publicOnly)
}

// GetSettingsAsMap gets settings as a key-value map
func (u *AdminUsecase) GetSettingsAsMap(ctx context.Context, tenantID string, publicOnly bool) (map[string]interface{}, error) {
	return u.settingRepo.GetAsMap(ctx, tenantID, publicOnly)
}

// InitializeSettings initializes default settings for a tenant
func (u *AdminUsecase) InitializeSettings(ctx context.Context, tenantID string) error {
	return u.settingRepo.InitializeDefaults(ctx, tenantID)
}

// Moderation

// ApproveFeedback approves a pending feedback
func (u *AdminUsecase) ApproveFeedback(ctx context.Context, id string) (*models.Feedback, error) {
	feedbackID, err := primitive.ObjectIDFromHex(id)
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

	feedback.Status = models.StatusApproved

	if err := u.feedbackRepo.Update(ctx, feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

// RejectFeedback rejects a feedback
func (u *AdminUsecase) RejectFeedback(ctx context.Context, id string) (*models.Feedback, error) {
	feedbackID, err := primitive.ObjectIDFromHex(id)
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

	feedback.Status = models.StatusRejected

	if err := u.feedbackRepo.Update(ctx, feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

// GetPendingFeedback gets all pending feedback
func (u *AdminUsecase) GetPendingFeedback(ctx context.Context, tenantID string, page, perPage int) ([]models.Feedback, int64, error) {
	filter := models.FeedbackFilter{
		TenantID: tenantID,
		Status:   []models.FeedbackStatus{models.StatusPending},
		Page:     page,
		PerPage:  perPage,
	}

	return u.feedbackRepo.List(ctx, filter)
}

// Stats

// GetFeedbackStats gets feedback statistics
func (u *AdminUsecase) GetFeedbackStats(ctx context.Context, tenantID string) (*models.FeedbackStats, error) {
	return u.feedbackRepo.GetStats(ctx, tenantID)
}
