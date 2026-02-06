package repository

import (
	"context"
	"errors"
	"time"

	"github.com/minisource/feedback/internal/database"
	"github.com/minisource/feedback/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// VoteRepository handles vote data operations
type VoteRepository struct {
	db *database.MongoDB
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository(db *database.MongoDB) *VoteRepository {
	return &VoteRepository{db: db}
}

// Create creates a new vote
func (r *VoteRepository) Create(ctx context.Context, vote *models.Vote) error {
	vote.ID = primitive.NewObjectID()
	vote.CreatedAt = time.Now()
	vote.UpdatedAt = time.Now()

	_, err := r.db.Collection(database.CollectionVotes).InsertOne(ctx, vote)
	return err
}

// GetByUserAndTarget gets a user's vote on a target
func (r *VoteRepository) GetByUserAndTarget(ctx context.Context, tenantID, userID, targetType string, targetID primitive.ObjectID) (*models.Vote, error) {
	var vote models.Vote
	err := r.db.Collection(database.CollectionVotes).FindOne(ctx, bson.M{
		"tenant_id":   tenantID,
		"user_id":     userID,
		"target_type": targetType,
		"target_id":   targetID,
	}).Decode(&vote)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// Update updates a vote
func (r *VoteRepository) Update(ctx context.Context, vote *models.Vote) error {
	vote.UpdatedAt = time.Now()
	_, err := r.db.Collection(database.CollectionVotes).ReplaceOne(ctx, bson.M{"_id": vote.ID}, vote)
	return err
}

// Delete deletes a vote
func (r *VoteRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionVotes).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// GetVoteSummary gets vote summary for a target
func (r *VoteRepository) GetVoteSummary(ctx context.Context, targetType string, targetID primitive.ObjectID) (*models.VoteSummary, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_type": targetType,
				"target_id":   targetID,
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"upvotes": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$vote_type", "up"}}, 1, 0},
					},
				},
				"downvotes": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$vote_type", "down"}}, 1, 0},
					},
				},
			},
		},
	}

	cursor, err := r.db.Collection(database.CollectionVotes).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Upvotes   int `bson:"upvotes"`
		Downvotes int `bson:"downvotes"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	summary := &models.VoteSummary{}
	if len(results) > 0 {
		summary.Upvotes = results[0].Upvotes
		summary.Downvotes = results[0].Downvotes
		summary.Score = summary.Upvotes - summary.Downvotes
	}

	return summary, nil
}

// GetUserVotesForTargets gets a user's votes for multiple targets
func (r *VoteRepository) GetUserVotesForTargets(ctx context.Context, userID, targetType string, targetIDs []primitive.ObjectID) (map[primitive.ObjectID]models.VoteType, error) {
	cursor, err := r.db.Collection(database.CollectionVotes).Find(ctx, bson.M{
		"user_id":     userID,
		"target_type": targetType,
		"target_id":   bson.M{"$in": targetIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var votes []models.Vote
	if err := cursor.All(ctx, &votes); err != nil {
		return nil, err
	}

	result := make(map[primitive.ObjectID]models.VoteType)
	for _, vote := range votes {
		result[vote.TargetID] = vote.VoteType
	}

	return result, nil
}

// CountByUser counts votes by a user
func (r *VoteRepository) CountByUser(ctx context.Context, tenantID, userID string) (int64, error) {
	return r.db.Collection(database.CollectionVotes).CountDocuments(ctx, bson.M{
		"tenant_id": tenantID,
		"user_id":   userID,
	})
}
