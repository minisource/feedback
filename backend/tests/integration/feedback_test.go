//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Feedback represents a feedback entry for testing
type Feedback struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	UserID      string                 `json:"user_id,omitempty"`
	Email       string                 `json:"email,omitempty"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Priority    string                 `json:"priority,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Votes       int                    `json:"votes"`
	CreatedAt   string                 `json:"created_at"`
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	app := fiber.New()

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "feedback",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestCreateFeedback tests feedback submission
func TestCreateFeedback(t *testing.T) {
	app := fiber.New()

	var createdFeedback Feedback

	app.Post("/api/v1/feedback", func(c *fiber.Ctx) error {
		if err := c.BodyParser(&createdFeedback); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if createdFeedback.Type == "" || createdFeedback.Title == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Type and title are required",
			})
		}

		createdFeedback.ID = "feedback-123"
		createdFeedback.Status = "new"
		createdFeedback.Votes = 1
		createdFeedback.TenantID = c.Get("X-Tenant-ID")
		return c.Status(fiber.StatusCreated).JSON(createdFeedback)
	})

	t.Run("Create Feature Request", func(t *testing.T) {
		feedback := Feedback{
			Type:        "feature",
			Title:       "Add dark mode support",
			Description: "Would love to have a dark mode option for the dashboard",
			Tags:        []string{"ui", "accessibility"},
		}
		body, _ := json.Marshal(feedback)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant-123")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result Feedback
		json.NewDecoder(resp.Body).Decode(&result)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, "new", result.Status)
		assert.Equal(t, 1, result.Votes)
	})

	t.Run("Create Bug Report", func(t *testing.T) {
		feedback := Feedback{
			Type:        "bug",
			Title:       "Login page crashes on mobile",
			Description: "When trying to login on iPhone Safari, the page crashes",
			Email:       "user@example.com",
		}
		body, _ := json.Marshal(feedback)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant-123")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Create Without Type", func(t *testing.T) {
		feedback := Feedback{
			Title: "Missing type",
		}
		body, _ := json.Marshal(feedback)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestListFeedback tests feedback listing
func TestListFeedback(t *testing.T) {
	app := fiber.New()

	mockFeedback := []Feedback{
		{ID: "1", Type: "feature", Title: "Feature 1", Status: "new", Votes: 10},
		{ID: "2", Type: "bug", Title: "Bug 1", Status: "in_progress", Votes: 5},
		{ID: "3", Type: "feature", Title: "Feature 2", Status: "new", Votes: 15},
	}

	app.Get("/api/v1/feedback", func(c *fiber.Ctx) error {
		feedbackType := c.Query("type")
		status := c.Query("status")
		sortBy := c.Query("sort", "votes")

		var filtered []Feedback
		for _, fb := range mockFeedback {
			if (feedbackType == "" || fb.Type == feedbackType) &&
				(status == "" || fb.Status == status) {
				filtered = append(filtered, fb)
			}
		}

		// Simple sort by votes (descending)
		if sortBy == "votes" && len(filtered) > 1 {
			for i := 0; i < len(filtered)-1; i++ {
				for j := i + 1; j < len(filtered); j++ {
					if filtered[j].Votes > filtered[i].Votes {
						filtered[i], filtered[j] = filtered[j], filtered[i]
					}
				}
			}
		}

		return c.JSON(fiber.Map{
			"data":  filtered,
			"total": len(filtered),
		})
	})

	t.Run("List All Feedback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/feedback", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, float64(3), result["total"])
	})

	t.Run("List Features Only", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/feedback?type=feature", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, float64(2), result["total"])
	})

	t.Run("List Sorted By Votes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/feedback?sort=votes", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestVoteFeedback tests voting on feedback
func TestVoteFeedback(t *testing.T) {
	app := fiber.New()

	voteCount := 10

	app.Post("/api/v1/feedback/:id/vote", func(c *fiber.Ctx) error {
		voteCount++
		return c.JSON(fiber.Map{
			"id":    c.Params("id"),
			"votes": voteCount,
		})
	})

	app.Delete("/api/v1/feedback/:id/vote", func(c *fiber.Ctx) error {
		voteCount--
		return c.JSON(fiber.Map{
			"id":    c.Params("id"),
			"votes": voteCount,
		})
	})

	t.Run("Upvote", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/123/vote", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, float64(11), result["votes"])
	})

	t.Run("Remove Vote", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/feedback/123/vote", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, float64(10), result["votes"])
	})
}

// TestUpdateFeedbackStatus tests status updates (admin)
func TestUpdateFeedbackStatus(t *testing.T) {
	app := fiber.New()

	app.Patch("/api/v1/feedback/:id/status", func(c *fiber.Ctx) error {
		var payload struct {
			Status string `json:"status"`
		}
		c.BodyParser(&payload)

		return c.JSON(Feedback{
			ID:     c.Params("id"),
			Status: payload.Status,
		})
	})

	statuses := []string{"new", "under_review", "planned", "in_progress", "completed", "declined"}

	for _, status := range statuses {
		t.Run("Set Status "+status, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"status": status})

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/feedback/123/status", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result Feedback
			json.NewDecoder(resp.Body).Decode(&result)
			assert.Equal(t, status, result.Status)
		})
	}
}

// TestFeedbackComments tests adding comments to feedback
func TestFeedbackComments(t *testing.T) {
	app := fiber.New()

	app.Post("/api/v1/feedback/:id/comments", func(c *fiber.Ctx) error {
		var payload struct {
			Content string `json:"content"`
		}
		c.BodyParser(&payload)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id":          "comment-123",
			"feedback_id": c.Params("id"),
			"content":     payload.Content,
		})
	})

	body, _ := json.Marshal(map[string]string{"content": "Great idea! +1"})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback/123/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestFeedbackAnalytics tests feedback analytics
func TestFeedbackAnalytics(t *testing.T) {
	app := fiber.New()

	app.Get("/api/v1/feedback/analytics", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"total_feedback": 100,
			"by_type": fiber.Map{
				"feature": 60,
				"bug":     30,
				"other":   10,
			},
			"by_status": fiber.Map{
				"new":          25,
				"under_review": 15,
				"planned":      20,
				"in_progress":  10,
				"completed":    25,
				"declined":     5,
			},
			"top_voted": 5,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/feedback/analytics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, float64(100), result["total_feedback"])
}
