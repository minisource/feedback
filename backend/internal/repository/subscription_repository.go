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

// SubscriptionRepository handles subscription data operations
type SubscriptionRepository struct {
	db *database.MongoDB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *database.MongoDB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create creates a new subscription
func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	sub.ID = primitive.NewObjectID()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	sub.IsActive = true

	_, err := r.db.Collection(database.CollectionSubscriptions).InsertOne(ctx, sub)
	return err
}

// GetByID gets subscription by ID
func (r *SubscriptionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.Collection(database.CollectionSubscriptions).FindOne(ctx, bson.M{"_id": id}).Decode(&sub)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// GetByUserAndFeedback gets a user's subscription to a specific feedback
func (r *SubscriptionRepository) GetByUserAndFeedback(ctx context.Context, userID string, feedbackID primitive.ObjectID) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.Collection(database.CollectionSubscriptions).FindOne(ctx, bson.M{
		"user_id":     userID,
		"feedback_id": feedbackID,
		"is_active":   true,
	}).Decode(&sub)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// Update updates a subscription
func (r *SubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	sub.UpdatedAt = time.Now()
	_, err := r.db.Collection(database.CollectionSubscriptions).ReplaceOne(ctx, bson.M{"_id": sub.ID}, sub)
	return err
}

// Unsubscribe deactivates a subscription
func (r *SubscriptionRepository) Unsubscribe(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.db.Collection(database.CollectionSubscriptions).UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"is_active":       false,
			"unsubscribed_at": now,
			"updated_at":      now,
		},
	})
	return err
}

// ListByUser lists subscriptions for a user
func (r *SubscriptionRepository) ListByUser(ctx context.Context, tenantID, userID string) ([]models.Subscription, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.db.Collection(database.CollectionSubscriptions).Find(ctx, bson.M{
		"tenant_id": tenantID,
		"user_id":   userID,
		"is_active": true,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subs []models.Subscription
	if err := cursor.All(ctx, &subs); err != nil {
		return nil, err
	}

	return subs, nil
}

// GetSubscribersForFeedback gets all active subscribers for a feedback
func (r *SubscriptionRepository) GetSubscribersForFeedback(ctx context.Context, feedbackID primitive.ObjectID, eventType models.SubscriptionType) ([]models.Subscription, error) {
	query := bson.M{
		"is_active": true,
		"$or": []bson.M{
			{"feedback_id": feedbackID},
			{"subscribe_all": true},
		},
	}

	// Filter by event type if specified
	if eventType != "" && eventType != models.SubAll {
		query["$or"] = []bson.M{
			{"feedback_id": feedbackID, "types": bson.M{"$in": []models.SubscriptionType{eventType, models.SubAll}}},
			{"subscribe_all": true, "types": bson.M{"$in": []models.SubscriptionType{eventType, models.SubAll}}},
		}
	}

	cursor, err := r.db.Collection(database.CollectionSubscriptions).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subs []models.Subscription
	if err := cursor.All(ctx, &subs); err != nil {
		return nil, err
	}

	return subs, nil
}

// GetSubscribersForCategory gets all active subscribers for a category
func (r *SubscriptionRepository) GetSubscribersForCategory(ctx context.Context, categoryID primitive.ObjectID, eventType models.SubscriptionType) ([]models.Subscription, error) {
	query := bson.M{
		"category_id": categoryID,
		"is_active":   true,
	}

	if eventType != "" && eventType != models.SubAll {
		query["types"] = bson.M{"$in": []models.SubscriptionType{eventType, models.SubAll}}
	}

	cursor, err := r.db.Collection(database.CollectionSubscriptions).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subs []models.Subscription
	if err := cursor.All(ctx, &subs); err != nil {
		return nil, err
	}

	return subs, nil
}

// IsSubscribed checks if a user is subscribed to a feedback
func (r *SubscriptionRepository) IsSubscribed(ctx context.Context, userID string, feedbackID primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection(database.CollectionSubscriptions).CountDocuments(ctx, bson.M{
		"user_id":     userID,
		"feedback_id": feedbackID,
		"is_active":   true,
	})
	return count > 0, err
}
