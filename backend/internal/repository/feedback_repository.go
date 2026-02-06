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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FeedbackRepository handles feedback data operations
type FeedbackRepository struct {
	db *database.MongoDB
}

// NewFeedbackRepository creates a new feedback repository
func NewFeedbackRepository(db *database.MongoDB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// Create creates a new feedback
func (r *FeedbackRepository) Create(ctx context.Context, feedback *models.Feedback) error {
	feedback.ID = primitive.NewObjectID()
	feedback.CreatedAt = time.Now()
	feedback.UpdatedAt = time.Now()

	_, err := r.db.Collection(database.CollectionFeedback).InsertOne(ctx, feedback)
	return err
}

// GetByID gets feedback by ID
func (r *FeedbackRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Feedback, error) {
	var feedback models.Feedback
	err := r.db.Collection(database.CollectionFeedback).FindOne(ctx, bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&feedback)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &feedback, nil
}

// Update updates a feedback
func (r *FeedbackRepository) Update(ctx context.Context, feedback *models.Feedback) error {
	feedback.UpdatedAt = time.Now()
	_, err := r.db.Collection(database.CollectionFeedback).ReplaceOne(ctx, bson.M{"_id": feedback.ID}, feedback)
	return err
}

// Delete soft deletes a feedback
func (r *FeedbackRepository) Delete(ctx context.Context, id primitive.ObjectID, deletedBy string) error {
	now := time.Now()
	_, err := r.db.Collection(database.CollectionFeedback).UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"deleted_by": deletedBy,
			"updated_at": now,
		},
	})
	return err
}

// List lists feedback with filtering and pagination
func (r *FeedbackRepository) List(ctx context.Context, filter models.FeedbackFilter) ([]models.Feedback, int64, error) {
	query := bson.M{
		"tenant_id":  filter.TenantID,
		"deleted_at": bson.M{"$exists": false},
	}

	if len(filter.Status) > 0 {
		query["status"] = bson.M{"$in": filter.Status}
	}
	if filter.CategoryID != "" {
		catID, err := primitive.ObjectIDFromHex(filter.CategoryID)
		if err == nil {
			query["category_id"] = catID
		}
	}
	if filter.AuthorID != "" {
		query["author_id"] = filter.AuthorID
	}
	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$in": filter.Tags}
	}
	if filter.Search != "" {
		query["$text"] = bson.M{"$search": filter.Search}
	}
	if filter.DateFrom != nil {
		query["created_at"] = bson.M{"$gte": *filter.DateFrom}
	}
	if filter.DateTo != nil {
		if _, exists := query["created_at"]; exists {
			query["created_at"].(bson.M)["$lte"] = *filter.DateTo
		} else {
			query["created_at"] = bson.M{"$lte": *filter.DateTo}
		}
	}

	// Count total
	total, err := r.db.Collection(database.CollectionFeedback).CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// Sort options
	sortField := "created_at"
	sortOrder := -1 // desc by default

	switch filter.SortBy {
	case "new":
		sortField = "created_at"
		sortOrder = -1
	case "top":
		sortField = "vote_score"
		sortOrder = -1
	case "trending":
		sortField = "trending_score"
		sortOrder = -1
	case "most_commented":
		sortField = "comment_count"
		sortOrder = -1
	}

	if filter.SortOrder == "asc" {
		sortOrder = 1
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	skip := int64((page - 1) * perPage)

	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(skip).
		SetLimit(int64(perPage))

	cursor, err := r.db.Collection(database.CollectionFeedback).Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var feedbacks []models.Feedback
	if err := cursor.All(ctx, &feedbacks); err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

// IncrementVotes increments vote counts
func (r *FeedbackRepository) IncrementVotes(ctx context.Context, id primitive.ObjectID, upvotes, downvotes int) error {
	_, err := r.db.Collection(database.CollectionFeedback).UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{
			"upvotes":    upvotes,
			"downvotes":  downvotes,
			"vote_score": upvotes - downvotes,
		},
		"$set": bson.M{"updated_at": time.Now()},
	})
	return err
}

// IncrementCommentCount increments comment count
func (r *FeedbackRepository) IncrementCommentCount(ctx context.Context, id primitive.ObjectID, delta int) error {
	_, err := r.db.Collection(database.CollectionFeedback).UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"comment_count": delta},
		"$set": bson.M{"updated_at": time.Now()},
	})
	return err
}

// IncrementViewCount increments view count
func (r *FeedbackRepository) IncrementViewCount(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionFeedback).UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"view_count": 1},
	})
	return err
}

// UpdateTrendingScore updates the trending score
func (r *FeedbackRepository) UpdateTrendingScore(ctx context.Context, id primitive.ObjectID, score float64) error {
	_, err := r.db.Collection(database.CollectionFeedback).UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"trending_score": score},
	})
	return err
}

// GetTrending gets trending feedback
func (r *FeedbackRepository) GetTrending(ctx context.Context, tenantID string, limit int) ([]models.Feedback, error) {
	query := bson.M{
		"tenant_id":  tenantID,
		"status":     models.StatusApproved,
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "trending_score", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.db.Collection(database.CollectionFeedback).Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var feedbacks []models.Feedback
	if err := cursor.All(ctx, &feedbacks); err != nil {
		return nil, err
	}

	return feedbacks, nil
}

// GetStats gets feedback statistics
func (r *FeedbackRepository) GetStats(ctx context.Context, tenantID string) (*models.FeedbackStats, error) {
	stats := &models.FeedbackStats{
		FeedbackByStatus:   make(map[string]int64),
		FeedbackByCategory: make(map[string]int64),
	}

	// Total feedback
	total, err := r.db.Collection(database.CollectionFeedback).CountDocuments(ctx, bson.M{
		"tenant_id":  tenantID,
		"deleted_at": bson.M{"$exists": false},
	})
	if err != nil {
		return nil, err
	}
	stats.TotalFeedback = total

	// Count by status
	pipeline := []bson.M{
		{"$match": bson.M{"tenant_id": tenantID, "deleted_at": bson.M{"$exists": false}}},
		{"$group": bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}},
	}

	cursor, err := r.db.Collection(database.CollectionFeedback).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var statusCounts []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	if err := cursor.All(ctx, &statusCounts); err != nil {
		return nil, err
	}

	for _, sc := range statusCounts {
		stats.FeedbackByStatus[sc.ID] = sc.Count
	}

	return stats, nil
}
