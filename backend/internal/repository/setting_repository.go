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

// SettingRepository handles setting data operations
type SettingRepository struct {
	db *database.MongoDB
}

// NewSettingRepository creates a new setting repository
func NewSettingRepository(db *database.MongoDB) *SettingRepository {
	return &SettingRepository{db: db}
}

// Get gets a setting by key
func (r *SettingRepository) Get(ctx context.Context, tenantID, key string) (*models.Setting, error) {
	var setting models.Setting
	err := r.db.Collection(database.CollectionSettings).FindOne(ctx, bson.M{
		"tenant_id": tenantID,
		"key":       key,
	}).Decode(&setting)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

// Set sets a setting value
func (r *SettingRepository) Set(ctx context.Context, setting *models.Setting) error {
	setting.UpdatedAt = time.Now()

	filter := bson.M{
		"tenant_id": setting.TenantID,
		"key":       setting.Key,
	}

	update := bson.M{
		"$set": bson.M{
			"value":       setting.Value,
			"type":        setting.Type,
			"description": setting.Description,
			"is_public":   setting.IsPublic,
			"updated_at":  setting.UpdatedAt,
			"updated_by":  setting.UpdatedBy,
		},
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"tenant_id":  setting.TenantID,
			"key":        setting.Key,
			"created_at": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.db.Collection(database.CollectionSettings).UpdateOne(ctx, filter, update, opts)
	return err
}

// GetAll gets all settings for a tenant
func (r *SettingRepository) GetAll(ctx context.Context, tenantID string, publicOnly bool) ([]models.Setting, error) {
	query := bson.M{"tenant_id": tenantID}
	if publicOnly {
		query["is_public"] = true
	}

	cursor, err := r.db.Collection(database.CollectionSettings).Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settings []models.Setting
	if err := cursor.All(ctx, &settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// GetAsMap gets all settings as a key-value map
func (r *SettingRepository) GetAsMap(ctx context.Context, tenantID string, publicOnly bool) (map[string]interface{}, error) {
	settings, err := r.GetAll(ctx, tenantID, publicOnly)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	return result, nil
}

// InitializeDefaults initializes default settings for a tenant
func (r *SettingRepository) InitializeDefaults(ctx context.Context, tenantID string) error {
	defaults := models.DefaultSettings(tenantID)

	for _, setting := range defaults {
		// Check if already exists
		existing, _ := r.Get(ctx, tenantID, setting.Key)
		if existing == nil {
			if err := r.Set(ctx, &setting); err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete deletes a setting
func (r *SettingRepository) Delete(ctx context.Context, tenantID, key string) error {
	_, err := r.db.Collection(database.CollectionSettings).DeleteOne(ctx, bson.M{
		"tenant_id": tenantID,
		"key":       key,
	})
	return err
}
