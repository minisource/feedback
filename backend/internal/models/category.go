package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Category represents a feedback category
type Category struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Name        string             `bson:"name" json:"name"`
	Slug        string             `bson:"slug" json:"slug"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Color       string             `bson:"color,omitempty" json:"color,omitempty"`
	Icon        string             `bson:"icon,omitempty" json:"icon,omitempty"`
	SortOrder   int                `bson:"sort_order" json:"sortOrder"`
	IsActive    bool               `bson:"is_active" json:"isActive"`

	// Stats
	FeedbackCount int `bson:"feedback_count" json:"feedbackCount"`

	// Parent category for hierarchical structure
	ParentID *primitive.ObjectID `bson:"parent_id,omitempty" json:"parentId,omitempty"`

	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`

	// For response - nested children (not stored in DB)
	Children []Category `bson:"-" json:"children,omitempty"`
}
