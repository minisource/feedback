package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeedbackStatus represents the status of a feedback item
type FeedbackStatus string

const (
	StatusPending   FeedbackStatus = "pending"
	StatusApproved  FeedbackStatus = "approved"
	StatusRejected  FeedbackStatus = "rejected"
	StatusUnderReview FeedbackStatus = "under_review"
	StatusPlanned   FeedbackStatus = "planned"
	StatusInProgress FeedbackStatus = "in_progress"
	StatusCompleted FeedbackStatus = "completed"
	StatusClosed    FeedbackStatus = "closed"
)

// Feedback represents a user feedback/suggestion
type Feedback struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Status      FeedbackStatus     `bson:"status" json:"status"`
	
	// Author information
	AuthorID    string `bson:"author_id" json:"authorId"`
	AuthorName  string `bson:"author_name" json:"authorName"`
	AuthorEmail string `bson:"author_email,omitempty" json:"authorEmail,omitempty"`
	AuthorAvatar string `bson:"author_avatar,omitempty" json:"authorAvatar,omitempty"`
	IsAnonymous bool   `bson:"is_anonymous" json:"isAnonymous"`
	
	// Category and tags
	CategoryID   *primitive.ObjectID `bson:"category_id,omitempty" json:"categoryId,omitempty"`
	CategoryName string              `bson:"category_name,omitempty" json:"categoryName,omitempty"`
	Tags         []string            `bson:"tags,omitempty" json:"tags,omitempty"`
	
	// Attachments (uploaded via storage service)
	Attachments []Attachment `bson:"attachments,omitempty" json:"attachments,omitempty"`
	
	// Voting
	Upvotes      int `bson:"upvotes" json:"upvotes"`
	Downvotes    int `bson:"downvotes" json:"downvotes"`
	VoteScore    int `bson:"vote_score" json:"voteScore"` // upvotes - downvotes
	
	// Engagement metrics
	ViewCount    int `bson:"view_count" json:"viewCount"`
	CommentCount int `bson:"comment_count" json:"commentCount"`
	
	// Trending score (calculated based on votes, comments, views, and time decay)
	TrendingScore float64 `bson:"trending_score" json:"trendingScore"`
	
	// Admin response
	OfficialResponse *OfficialResponse `bson:"official_response,omitempty" json:"officialResponse,omitempty"`
	
	// Timestamps
	CreatedAt   time.Time  `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `bson:"updated_at" json:"updatedAt"`
	ApprovedAt  *time.Time `bson:"approved_at,omitempty" json:"approvedAt,omitempty"`
	CompletedAt *time.Time `bson:"completed_at,omitempty" json:"completedAt,omitempty"`
	
	// Soft delete
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deletedAt,omitempty"`
	DeletedBy string     `bson:"deleted_by,omitempty" json:"deletedBy,omitempty"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID        string    `bson:"id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	URL       string    `bson:"url" json:"url"`
	Size      int64     `bson:"size" json:"size"`
	MimeType  string    `bson:"mime_type" json:"mimeType"`
	StorageID string    `bson:"storage_id" json:"storageId"` // ID from storage service
	UploadedAt time.Time `bson:"uploaded_at" json:"uploadedAt"`
}

// OfficialResponse represents an admin/official response to feedback
type OfficialResponse struct {
	Content     string    `bson:"content" json:"content"`
	RespondedBy string    `bson:"responded_by" json:"respondedBy"`
	RespondedByName string `bson:"responded_by_name" json:"respondedByName"`
	RespondedAt time.Time `bson:"responded_at" json:"respondedAt"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updatedAt"`
}
