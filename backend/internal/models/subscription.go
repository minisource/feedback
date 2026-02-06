package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionType represents what event types to subscribe to
type SubscriptionType string

const (
	SubNewFeedback      SubscriptionType = "new_feedback"
	SubFeedbackUpdate   SubscriptionType = "feedback_update"
	SubStatusChange     SubscriptionType = "status_change"
	SubNewComment       SubscriptionType = "new_comment"
	SubOfficialResponse SubscriptionType = "official_response"
	SubAll              SubscriptionType = "all"
)

// Subscription represents a user's subscription to feedback updates
type Subscription struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID  string             `bson:"tenant_id" json:"tenantId"`
	UserID    string             `bson:"user_id" json:"userId"`
	UserEmail string             `bson:"user_email" json:"userEmail"`

	// What to subscribe to
	FeedbackID   *primitive.ObjectID `bson:"feedback_id,omitempty" json:"feedbackId,omitempty"` // Subscribe to specific feedback
	CategoryID   *primitive.ObjectID `bson:"category_id,omitempty" json:"categoryId,omitempty"` // Subscribe to category
	SubscribeAll bool                `bson:"subscribe_all" json:"subscribeAll"`                 // Subscribe to all new feedback

	// Subscription types
	Types []SubscriptionType `bson:"types" json:"types"`

	// Notification preferences
	EmailEnabled bool `bson:"email_enabled" json:"emailEnabled"`
	PushEnabled  bool `bson:"push_enabled" json:"pushEnabled"`

	// Status
	IsActive bool `bson:"is_active" json:"isActive"`

	// Timestamps
	CreatedAt      time.Time  `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time  `bson:"updated_at" json:"updatedAt"`
	UnsubscribedAt *time.Time `bson:"unsubscribed_at,omitempty" json:"unsubscribedAt,omitempty"`
}
