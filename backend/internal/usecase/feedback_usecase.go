package usecase

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	"github.com/minisource/feedback/config"
	"github.com/minisource/feedback/internal/models"
	"github.com/minisource/feedback/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrFeedbackNotFound  = errors.New("feedback not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrInvalidVoteType   = errors.New("invalid vote type")
	ErrAlreadyVoted      = errors.New("already voted with same type")
	ErrVotingDisabled    = errors.New("voting is disabled")
	ErrAnonymousDisabled = errors.New("anonymous feedback is disabled")
	ErrApprovalRequired  = errors.New("approval required")
	ErrInvalidStatus     = errors.New("invalid status")
)

// NotifierClient sends in-app notifications via the Notifier service.
type NotifierClient interface {
	SendNotification(ctx context.Context, notification NotificationRequest) error
}

// NotificationRequest represents a notification to send.
type NotificationRequest struct {
	Type       string
	Recipients []string
	Title      string
	Body       string
	Data       map[string]string
}

// FeedbackUsecase handles feedback business logic
type FeedbackUsecase struct {
	feedbackRepo     *repository.FeedbackRepository
	voteRepo         *repository.VoteRepository
	categoryRepo     *repository.CategoryRepository
	settingRepo      *repository.SettingRepository
	subscriptionRepo *repository.SubscriptionRepository
	notifier         NotifierClient
	cfg              *config.Config
}

// NewFeedbackUsecase creates a new feedback usecase
func NewFeedbackUsecase(
	feedbackRepo *repository.FeedbackRepository,
	voteRepo *repository.VoteRepository,
	categoryRepo *repository.CategoryRepository,
	settingRepo *repository.SettingRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	notifier NotifierClient,
	cfg *config.Config,
) *FeedbackUsecase {
	return &FeedbackUsecase{
		feedbackRepo:     feedbackRepo,
		voteRepo:         voteRepo,
		categoryRepo:     categoryRepo,
		settingRepo:      settingRepo,
		subscriptionRepo: subscriptionRepo,
		notifier:         notifier,
		cfg:              cfg,
	}
}

// Create creates a new feedback
func (u *FeedbackUsecase) Create(ctx context.Context, req models.CreateFeedbackRequest, tenantID, userID, userName, userEmail string) (*models.Feedback, error) {
	// Check if anonymous is allowed
	if userID == "" {
		setting, _ := u.settingRepo.Get(ctx, tenantID, "allow_anonymous")
		if setting != nil && setting.Value == false {
			return nil, ErrAnonymousDisabled
		}
	}

	// Parse category ID
	var categoryID *primitive.ObjectID
	if req.CategoryID != "" {
		catID, err := primitive.ObjectIDFromHex(req.CategoryID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		category, err := u.categoryRepo.GetByID(ctx, catID)
		if err != nil || category == nil {
			return nil, ErrCategoryNotFound
		}
		categoryID = &catID
	}

	// Determine initial status
	status := models.StatusApproved
	setting, _ := u.settingRepo.Get(ctx, tenantID, "require_approval")
	if setting != nil && setting.Value == true {
		status = models.StatusPending
	}

	// Create feedback
	feedback := &models.Feedback{
		TenantID:    tenantID,
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  categoryID,
		Tags:        req.Tags,
		Status:      status,
		AuthorID:    userID,
		AuthorName:  userName,
		AuthorEmail: userEmail,
		IsAnonymous: req.IsAnonymous,
	}

	// Handle attachments
	if len(req.Attachments) > 0 {
		for _, att := range req.Attachments {
			feedback.Attachments = append(feedback.Attachments, models.Attachment{
				ID:         primitive.NewObjectID().Hex(),
				Name:       att.Name,
				URL:        att.URL,
				Size:       att.Size,
				MimeType:   att.MimeType,
				StorageID:  att.StorageID,
				UploadedAt: time.Now(),
			})
		}
	}

	if err := u.feedbackRepo.Create(ctx, feedback); err != nil {
		return nil, err
	}

	// Update category count
	if categoryID != nil {
		_ = u.categoryRepo.IncrementFeedbackCount(ctx, *categoryID, 1)
	}

	// Auto-subscribe author
	if userID != "" {
		sub := &models.Subscription{
			TenantID:   tenantID,
			UserID:     userID,
			UserEmail:  userEmail,
			FeedbackID: &feedback.ID,
			Types:      []models.SubscriptionType{models.SubAll},
		}
		_ = u.subscriptionRepo.Create(ctx, sub)
	}

	u.sendNewFeedbackNotification(ctx, tenantID, feedback)

	return feedback, nil
}

func (u *FeedbackUsecase) sendNewFeedbackNotification(ctx context.Context, tenantID string, feedback *models.Feedback) {
	if u.notifier == nil || u.cfg == nil || !u.cfg.Notifier.Enabled || u.cfg.Notifier.AdminUserID == "" {
		return
	}

	setting, _ := u.settingRepo.Get(ctx, tenantID, models.SettingNotifyOnNewFeedback)
	if setting != nil && setting.Value == false {
		return
	}

	notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	title := "New Feedback"
	if feedback.Status == models.StatusPending {
		title = "Feedback Pending Approval"
	}

	notification := NotificationRequest{
		Type:       "feedback.new",
		Recipients: []string{u.cfg.Notifier.AdminUserID},
		Title:      title,
		Body:       truncateFeedbackText(feedback.Title, 100),
		Data: map[string]string{
			"feedback_id": feedback.ID.Hex(),
			"tenant_id":   feedback.TenantID,
			"author_id":   feedback.AuthorID,
			"status":      string(feedback.Status),
		},
	}

	if err := u.notifier.SendNotification(notifyCtx, notification); err != nil {
		log.Printf("Failed to send feedback notification: %v", err)
	}
}

func truncateFeedbackText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// GetByID gets feedback by ID
func (u *FeedbackUsecase) GetByID(ctx context.Context, id string, userID string) (*models.FeedbackResponse, error) {
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

	// Increment view count
	_ = u.feedbackRepo.IncrementViewCount(ctx, feedbackID)

	// Get user's vote if logged in
	var userVote *models.VoteType
	if userID != "" {
		vote, _ := u.voteRepo.GetByUserAndTarget(ctx, feedback.TenantID, userID, "feedback", feedbackID)
		if vote != nil {
			userVote = &vote.VoteType
		}
	}

	// Check subscription
	isSubscribed := false
	if userID != "" {
		isSubscribed, _ = u.subscriptionRepo.IsSubscribed(ctx, userID, feedbackID)
	}

	return u.toResponse(feedback, userVote, isSubscribed), nil
}

// List lists feedback with filtering
func (u *FeedbackUsecase) List(ctx context.Context, filter models.FeedbackFilter, userID string) ([]models.FeedbackResponse, int64, error) {
	feedbacks, total, err := u.feedbackRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Get user votes for all feedbacks
	var userVotes map[primitive.ObjectID]models.VoteType
	if userID != "" {
		ids := make([]primitive.ObjectID, len(feedbacks))
		for i, f := range feedbacks {
			ids[i] = f.ID
		}
		userVotes, _ = u.voteRepo.GetUserVotesForTargets(ctx, userID, "feedback", ids)
	}

	responses := make([]models.FeedbackResponse, len(feedbacks))
	for i, f := range feedbacks {
		var userVote *models.VoteType
		if v, ok := userVotes[f.ID]; ok {
			userVote = &v
		}
		responses[i] = *u.toResponse(&f, userVote, false)
	}

	return responses, total, nil
}

// Update updates a feedback
func (u *FeedbackUsecase) Update(ctx context.Context, id string, req models.UpdateFeedbackRequest, userID string, isAdmin bool) (*models.Feedback, error) {
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

	// Check authorization
	if !isAdmin && feedback.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	// Update fields
	if req.Title != nil && *req.Title != "" {
		feedback.Title = *req.Title
	}
	if req.Description != nil && *req.Description != "" {
		feedback.Description = *req.Description
	}
	if req.CategoryID != nil && *req.CategoryID != "" {
		catID, _ := primitive.ObjectIDFromHex(*req.CategoryID)
		feedback.CategoryID = &catID
	}
	if len(req.Tags) > 0 {
		feedback.Tags = req.Tags
	}

	if err := u.feedbackRepo.Update(ctx, feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

// UpdateStatus updates feedback status (admin only)
func (u *FeedbackUsecase) UpdateStatus(ctx context.Context, id string, status models.FeedbackStatus, userID string) (*models.Feedback, error) {
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

	// Validate status
	validStatuses := []models.FeedbackStatus{
		models.StatusPending, models.StatusApproved, models.StatusRejected,
		models.StatusPlanned, models.StatusInProgress, models.StatusCompleted, models.StatusClosed,
	}
	valid := false
	for _, s := range validStatuses {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidStatus
	}

	feedback.Status = status

	if err := u.feedbackRepo.Update(ctx, feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

// Delete deletes a feedback
func (u *FeedbackUsecase) Delete(ctx context.Context, id string, userID string, isAdmin bool) error {
	feedbackID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrFeedbackNotFound
	}

	feedback, err := u.feedbackRepo.GetByID(ctx, feedbackID)
	if err != nil {
		return err
	}
	if feedback == nil {
		return ErrFeedbackNotFound
	}

	// Check authorization
	if !isAdmin && feedback.AuthorID != userID {
		return ErrUnauthorized
	}

	// Update category count
	if feedback.CategoryID != nil {
		_ = u.categoryRepo.IncrementFeedbackCount(ctx, *feedback.CategoryID, -1)
	}

	return u.feedbackRepo.Delete(ctx, feedbackID, userID)
}

// Vote handles voting on feedback
func (u *FeedbackUsecase) Vote(ctx context.Context, id string, voteType models.VoteType, tenantID, userID string) error {
	// Check if voting is enabled
	setting, _ := u.settingRepo.Get(ctx, tenantID, "enable_voting")
	if setting != nil && setting.Value == false {
		return ErrVotingDisabled
	}

	feedbackID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrFeedbackNotFound
	}

	feedback, err := u.feedbackRepo.GetByID(ctx, feedbackID)
	if err != nil {
		return err
	}
	if feedback == nil {
		return ErrFeedbackNotFound
	}

	// Validate vote type
	if voteType != models.VoteUp && voteType != models.VoteDown {
		return ErrInvalidVoteType
	}

	// Check existing vote
	existingVote, err := u.voteRepo.GetByUserAndTarget(ctx, tenantID, userID, "feedback", feedbackID)
	if err != nil {
		return err
	}

	if existingVote != nil {
		if existingVote.VoteType == voteType {
			// Remove vote
			if err := u.voteRepo.Delete(ctx, existingVote.ID); err != nil {
				return err
			}
			// Update feedback counts
			if voteType == models.VoteUp {
				return u.feedbackRepo.IncrementVotes(ctx, feedbackID, -1, 0)
			}
			return u.feedbackRepo.IncrementVotes(ctx, feedbackID, 0, -1)
		}

		// Change vote
		existingVote.VoteType = voteType
		if err := u.voteRepo.Update(ctx, existingVote); err != nil {
			return err
		}
		// Update feedback counts
		if voteType == models.VoteUp {
			return u.feedbackRepo.IncrementVotes(ctx, feedbackID, 1, -1)
		}
		return u.feedbackRepo.IncrementVotes(ctx, feedbackID, -1, 1)
	}

	// Create new vote
	vote := &models.Vote{
		TenantID:   tenantID,
		UserID:     userID,
		TargetType: "feedback",
		TargetID:   feedbackID,
		VoteType:   voteType,
	}
	if err := u.voteRepo.Create(ctx, vote); err != nil {
		return err
	}

	// Update feedback counts
	if voteType == models.VoteUp {
		return u.feedbackRepo.IncrementVotes(ctx, feedbackID, 1, 0)
	}
	return u.feedbackRepo.IncrementVotes(ctx, feedbackID, 0, 1)
}

// AddOfficialResponse adds an official response (admin only)
func (u *FeedbackUsecase) AddOfficialResponse(ctx context.Context, id string, content, responderID, responderName string) (*models.Feedback, error) {
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

	feedback.OfficialResponse = &models.OfficialResponse{
		Content:         content,
		RespondedBy:     responderID,
		RespondedByName: responderName,
		RespondedAt:     time.Now(),
	}

	if err := u.feedbackRepo.Update(ctx, feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

// GetTrending gets trending feedback
func (u *FeedbackUsecase) GetTrending(ctx context.Context, tenantID string, limit int) ([]models.FeedbackResponse, error) {
	feedbacks, err := u.feedbackRepo.GetTrending(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]models.FeedbackResponse, len(feedbacks))
	for i, f := range feedbacks {
		responses[i] = *u.toResponse(&f, nil, false)
	}

	return responses, nil
}

// UpdateTrendingScores recalculates trending scores
func (u *FeedbackUsecase) UpdateTrendingScores(ctx context.Context, tenantID string) error {
	filter := models.FeedbackFilter{
		TenantID: tenantID,
		Status:   []models.FeedbackStatus{models.StatusApproved},
		PerPage:  1000,
		Page:     1,
	}

	feedbacks, _, err := u.feedbackRepo.List(ctx, filter)
	if err != nil {
		return err
	}

	for _, f := range feedbacks {
		score := u.calculateTrendingScore(&f)
		_ = u.feedbackRepo.UpdateTrendingScore(ctx, f.ID, score)
	}

	return nil
}

// calculateTrendingScore calculates trending score using a time-decay algorithm
func (u *FeedbackUsecase) calculateTrendingScore(f *models.Feedback) float64 {
	// Base score from votes
	score := float64(f.VoteScore)

	// Add comment activity
	score += float64(f.CommentCount) * 0.5

	// Time decay (half-life of 7 days)
	hoursSinceCreation := time.Since(f.CreatedAt).Hours()
	decay := math.Pow(0.5, hoursSinceCreation/(7*24))

	return score * decay
}

// GetStats gets feedback statistics
func (u *FeedbackUsecase) GetStats(ctx context.Context, tenantID string) (*models.FeedbackStats, error) {
	return u.feedbackRepo.GetStats(ctx, tenantID)
}

// toResponse converts Feedback to FeedbackResponse
func (u *FeedbackUsecase) toResponse(f *models.Feedback, userVote *models.VoteType, isSubscribed bool) *models.FeedbackResponse {
	resp := &models.FeedbackResponse{
		ID:               f.ID.Hex(),
		Title:            f.Title,
		Description:      f.Description,
		Tags:             f.Tags,
		Status:           f.Status,
		Upvotes:          f.Upvotes,
		Downvotes:        f.Downvotes,
		VoteScore:        f.VoteScore,
		CommentCount:     f.CommentCount,
		ViewCount:        f.ViewCount,
		Attachments:      f.Attachments,
		OfficialResponse: f.OfficialResponse,
		CreatedAt:        f.CreatedAt,
		UpdatedAt:        f.UpdatedAt,
		IsSubscribed:     isSubscribed,
	}

	if f.CategoryID != nil {
		resp.CategoryID = f.CategoryID.Hex()
	}

	if !f.IsAnonymous {
		resp.AuthorID = f.AuthorID
		resp.AuthorName = f.AuthorName
	}

	if userVote != nil {
		voteStr := string(*userVote)
		resp.UserVote = &voteStr
	}

	return resp
}
