package repository

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/minisource/feedback/internal/database"
	"github.com/minisource/feedback/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CategoryRepository handles category data operations
type CategoryRepository struct {
	db *database.MongoDB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *database.MongoDB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	// Generate slug if not provided
	if category.Slug == "" {
		category.Slug = r.generateSlug(category.Name)
	}

	_, err := r.db.Collection(database.CollectionCategories).InsertOne(ctx, category)
	return err
}

// GetByID gets category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Category, error) {
	var category models.Category
	err := r.db.Collection(database.CollectionCategories).FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

// GetBySlug gets category by slug
func (r *CategoryRepository) GetBySlug(ctx context.Context, tenantID, slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.Collection(database.CollectionCategories).FindOne(ctx, bson.M{
		"tenant_id": tenantID,
		"slug":      slug,
	}).Decode(&category)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	category.UpdatedAt = time.Now()
	_, err := r.db.Collection(database.CollectionCategories).ReplaceOne(ctx, bson.M{"_id": category.ID}, category)
	return err
}

// Delete deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionCategories).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// List lists all categories for a tenant
func (r *CategoryRepository) List(ctx context.Context, tenantID string, activeOnly bool) ([]models.Category, error) {
	query := bson.M{"tenant_id": tenantID}
	if activeOnly {
		query["is_active"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}, {Key: "name", Value: 1}})

	cursor, err := r.db.Collection(database.CollectionCategories).Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// IncrementFeedbackCount increments feedback count
func (r *CategoryRepository) IncrementFeedbackCount(ctx context.Context, id primitive.ObjectID, delta int) error {
	_, err := r.db.Collection(database.CollectionCategories).UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"feedback_count": delta},
		"$set": bson.M{"updated_at": time.Now()},
	})
	return err
}

// generateSlug generates a URL-friendly slug from a name
func (r *CategoryRepository) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	return slug
}
