package database

import (
	"context"
	"fmt"
	"time"

	"github.com/minisource/feedback/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collections
const (
	CollectionFeedback      = "feedback"
	CollectionComments      = "comments"
	CollectionVotes         = "votes"
	CollectionCategories    = "categories"
	CollectionSettings      = "settings"
	CollectionSubscriptions = "subscriptions"
)

// MongoDB holds the MongoDB client and database
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(cfg config.MongoDBConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime)

	// Connect
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.Database)

	mongodb := &MongoDB{
		Client:   client,
		Database: db,
	}

	// Create indexes
	if err := mongodb.CreateIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return mongodb, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

// Collection returns a collection by name
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// CreateIndexes creates all necessary indexes
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// Feedback indexes
	feedbackIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "category_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "author_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "vote_score", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "trending_score", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "title", Value: "text"},
				{Key: "description", Value: "text"},
			},
		},
		{
			Keys: bson.D{{Key: "tags", Value: 1}},
		},
	}

	_, err := m.Collection(CollectionFeedback).Indexes().CreateMany(ctx, feedbackIndexes)
	if err != nil {
		return fmt.Errorf("failed to create feedback indexes: %w", err)
	}

	// Comment indexes
	commentIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "feedback_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "feedback_id", Value: 1},
				{Key: "parent_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "feedback_id", Value: 1},
				{Key: "vote_score", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "author_id", Value: 1},
			},
		},
	}

	_, err = m.Collection(CollectionComments).Indexes().CreateMany(ctx, commentIndexes)
	if err != nil {
		return fmt.Errorf("failed to create comment indexes: %w", err)
	}

	// Vote indexes
	voteIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "target_type", Value: 1},
				{Key: "target_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
			},
		},
	}

	_, err = m.Collection(CollectionVotes).Indexes().CreateMany(ctx, voteIndexes)
	if err != nil {
		return fmt.Errorf("failed to create vote indexes: %w", err)
	}

	// Category indexes
	categoryIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "slug", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "sort_order", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "is_active", Value: 1},
			},
		},
	}

	_, err = m.Collection(CollectionCategories).Indexes().CreateMany(ctx, categoryIndexes)
	if err != nil {
		return fmt.Errorf("failed to create category indexes: %w", err)
	}

	// Settings indexes
	settingsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "key", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err = m.Collection(CollectionSettings).Indexes().CreateMany(ctx, settingsIndexes)
	if err != nil {
		return fmt.Errorf("failed to create settings indexes: %w", err)
	}

	// Subscription indexes
	subscriptionIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "feedback_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "category_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "feedback_id", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"feedback_id": bson.M{"$exists": true}}),
		},
	}

	_, err = m.Collection(CollectionSubscriptions).Indexes().CreateMany(ctx, subscriptionIndexes)
	if err != nil {
		return fmt.Errorf("failed to create subscription indexes: %w", err)
	}

	return nil
}
