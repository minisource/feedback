package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ========================
// Request DTOs
// ========================

// CreateFeedbackRequest represents the request to create feedback
type CreateFeedbackRequest struct {
	TenantID    string                 `json:"tenantId"`
	Title       string                 `json:"title" validate:"required,max=200"`
	Description string                 `json:"description" validate:"required,max=10000"`
	CategoryID  string                 `json:"categoryId,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	IsAnonymous bool                   `json:"isAnonymous"`
	Attachments []AddAttachmentRequest `json:"attachments,omitempty"`
}

// UpdateFeedbackRequest represents the request to update feedback
type UpdateFeedbackRequest struct {
	Title       *string  `json:"title,omitempty" validate:"omitempty,max=200"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=10000"`
	CategoryID  *string  `json:"categoryId,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ChangeFeedbackStatusRequest represents a status change request
type ChangeFeedbackStatusRequest struct {
	Status  FeedbackStatus `json:"status" validate:"required"`
	Comment string         `json:"comment,omitempty"`
}

// AddOfficialResponseRequest represents the request to add official response
type AddOfficialResponseRequest struct {
	Content string `json:"content" validate:"required,max=10000"`
}

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	Content     string `json:"content" validate:"required,max=2000"`
	ParentID    string `json:"parentId,omitempty"`
	IsAnonymous bool   `json:"isAnonymous"`
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,max=2000"`
}

// VoteRequest represents a vote request
type VoteRequest struct {
	VoteType VoteType `json:"voteType" validate:"required,oneof=up down"`
}

// CreateCategoryRequest represents the request to create a category
type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty" validate:"max=500"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	ParentID    string `json:"parentId,omitempty"`
	SortOrder   int    `json:"sortOrder"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,max=100"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`
	SortOrder   *int    `json:"sortOrder,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`
}

// SubscribeRequest represents a subscription request
type SubscribeRequest struct {
	FeedbackID   string             `json:"feedbackId,omitempty"`
	CategoryID   string             `json:"categoryId,omitempty"`
	SubscribeAll bool               `json:"subscribeAll"`
	Types        []SubscriptionType `json:"types"`
	EmailEnabled bool               `json:"emailEnabled"`
	PushEnabled  bool               `json:"pushEnabled"`
}

// UpdateSettingRequest represents a setting update request
type UpdateSettingRequest struct {
	Key         string      `json:"key" validate:"required"`
	Value       interface{} `json:"value" validate:"required"`
	Type        string      `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	IsPublic    bool        `json:"isPublic"`
}

// UpdateSubscriptionRequest represents subscription update request
type UpdateSubscriptionRequest struct {
	Types        []SubscriptionType `json:"types,omitempty"`
	EmailEnabled *bool              `json:"emailEnabled,omitempty"`
	PushEnabled  *bool              `json:"pushEnabled,omitempty"`
}

// BulkUpdateSettingsRequest represents bulk settings update
type BulkUpdateSettingsRequest struct {
	Settings map[string]interface{} `json:"settings" validate:"required"`
}

// AddAttachmentRequest represents an attachment upload request
type AddAttachmentRequest struct {
	Name      string `json:"name" validate:"required"`
	StorageID string `json:"storageId" validate:"required"`
	URL       string `json:"url" validate:"required"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mimeType"`
}

// ========================
// Filter/Query DTOs
// ========================

// FeedbackFilter represents filter options for listing feedback
type FeedbackFilter struct {
	TenantID   string           `query:"tenantId"`
	Status     []FeedbackStatus `query:"status"`
	CategoryID string           `query:"categoryId"`
	AuthorID   string           `query:"authorId"`
	Tags       []string         `query:"tags"`
	Search     string           `query:"search"`
	SortBy     string           `query:"sortBy"`    // "new", "top", "trending", "most_commented"
	SortOrder  string           `query:"sortOrder"` // "asc", "desc"
	DateFrom   *time.Time       `query:"dateFrom"`
	DateTo     *time.Time       `query:"dateTo"`
	Page       int              `query:"page"`
	PerPage    int              `query:"perPage"`
}

// CommentFilter represents filter options for listing comments
type CommentFilter struct {
	FeedbackID string `query:"feedbackId"`
	ParentID   string `query:"parentId"`
	AuthorID   string `query:"authorId"`
	SortBy     string `query:"sortBy"` // "new", "top", "oldest"
	Page       int    `query:"page"`
	PerPage    int    `query:"perPage"`
}

// ========================
// Response DTOs
// ========================

// FeedbackResponse represents feedback in API response
type FeedbackResponse struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	CategoryID       string            `json:"categoryId,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Status           FeedbackStatus    `json:"status"`
	AuthorID         string            `json:"authorId,omitempty"`
	AuthorName       string            `json:"authorName,omitempty"`
	Upvotes          int               `json:"upvotes"`
	Downvotes        int               `json:"downvotes"`
	VoteScore        int               `json:"voteScore"`
	CommentCount     int               `json:"commentCount"`
	ViewCount        int               `json:"viewCount"`
	Attachments      []Attachment      `json:"attachments,omitempty"`
	OfficialResponse *OfficialResponse `json:"officialResponse,omitempty"`
	Category         *CategorySummary  `json:"category,omitempty"`
	UserVote         *string           `json:"userVote,omitempty"`
	IsSubscribed     bool              `json:"isSubscribed"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

// CategorySummary represents a category summary
type CategorySummary struct {
	ID    primitive.ObjectID `json:"id"`
	Name  string             `json:"name"`
	Slug  string             `json:"slug"`
	Color string             `json:"color,omitempty"`
	Icon  string             `json:"icon,omitempty"`
}

// FeedbackStats represents feedback statistics
type FeedbackStats struct {
	TotalFeedback      int64              `json:"totalFeedback"`
	TotalComments      int64              `json:"totalComments"`
	TotalVotes         int64              `json:"totalVotes"`
	FeedbackByStatus   map[string]int64   `json:"feedbackByStatus"`
	FeedbackByCategory map[string]int64   `json:"feedbackByCategory"`
	TopContributors    []ContributorStats `json:"topContributors"`
	RecentActivity     int64              `json:"recentActivity"` // Last 24 hours
}

// ContributorStats represents a contributor's statistics
type ContributorStats struct {
	UserID        string `json:"userId"`
	UserName      string `json:"userName"`
	FeedbackCount int64  `json:"feedbackCount"`
	CommentCount  int64  `json:"commentCount"`
	VoteCount     int64  `json:"voteCount"`
}

// TrendingFeedback represents feedback with trending info
type TrendingFeedback struct {
	*Feedback
	TrendRank int `json:"trendRank"`
}

// CommentResponse represents comment in API response
type CommentResponse struct {
	ID         string    `json:"id"`
	FeedbackID string    `json:"feedbackId"`
	ParentID   string    `json:"parentId,omitempty"`
	Content    string    `json:"content"`
	AuthorID   string    `json:"authorId,omitempty"`
	AuthorName string    `json:"authorName,omitempty"`
	Upvotes    int       `json:"upvotes"`
	Downvotes  int       `json:"downvotes"`
	VoteScore  int       `json:"voteScore"`
	Depth      int       `json:"depth"`
	ReplyCount int       `json:"replyCount"`
	IsPinned   bool      `json:"isPinned"`
	IsEdited   bool      `json:"isEdited"`
	UserVote   *string   `json:"userVote,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// SubscriptionResponse represents subscription in API response
type SubscriptionResponse struct {
	ID           string             `json:"id"`
	FeedbackID   string             `json:"feedbackId,omitempty"`
	CategoryID   string             `json:"categoryId,omitempty"`
	Types        []SubscriptionType `json:"types"`
	SubscribeAll bool               `json:"subscribeAll"`
	EmailEnabled bool               `json:"emailEnabled"`
	PushEnabled  bool               `json:"pushEnabled"`
	CreatedAt    time.Time          `json:"createdAt"`
}
