package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VoteType represents the type of vote
type VoteType string

const (
	VoteUp   VoteType = "up"
	VoteDown VoteType = "down"
)

// Vote represents a user's vote on feedback or comment
type Vote struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID   string             `bson:"tenant_id" json:"tenantId"`
	TargetType string             `bson:"target_type" json:"targetType"` // "feedback" or "comment"
	TargetID   primitive.ObjectID `bson:"target_id" json:"targetId"`
	UserID     string             `bson:"user_id" json:"userId"`
	VoteType   VoteType           `bson:"vote_type" json:"voteType"`
	CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updatedAt"`
}

// VoteSummary represents the vote counts for a target
type VoteSummary struct {
	Upvotes   int      `json:"upvotes"`
	Downvotes int      `json:"downvotes"`
	Score     int      `json:"score"`
	UserVote  VoteType `json:"userVote,omitempty"` // Current user's vote, if any
}
