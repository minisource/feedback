//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestFeedback_API(t *testing.T) {
	c := e2e.NewClient(e2e.BaseURLFromEnv("FEEDBACK_BASE_URL", "http://127.0.0.1:5012"), nil)
	c.RequireUp(t, "/health")

	c.RunCases(t, []e2e.Case{
		{Name: "health", Method: http.MethodGet, Path: "/health", WantCode: []int{http.StatusOK}},
		{Name: "ready", Method: http.MethodGet, Path: "/ready", WantCode: []int{http.StatusOK}},
		{Name: "feedback_list", Method: http.MethodGet, Path: "/api/v1/feedback", WantCode: []int{http.StatusOK, http.StatusTooManyRequests}},
		{Name: "feedback_trending", Method: http.MethodGet, Path: "/api/v1/feedback/trending", WantCode: []int{http.StatusOK, http.StatusTooManyRequests}},
		{Name: "feedback_stats", Method: http.MethodGet, Path: "/api/v1/feedback/stats", WantCode: []int{http.StatusOK, http.StatusTooManyRequests}},
		{Name: "categories", Method: http.MethodGet, Path: "/api/v1/categories", WantCode: []int{http.StatusOK, http.StatusTooManyRequests}},
		{Name: "settings", Method: http.MethodGet, Path: "/api/v1/settings", WantCode: []int{http.StatusOK, http.StatusTooManyRequests}},
	})
}

func TestFeedback_ProtectedEndpoints(t *testing.T) {
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	c := e2e.NewClient(e2e.BaseURLFromEnv("FEEDBACK_BASE_URL", "http://127.0.0.1:5012"), e2e.Bearer(token))
	c.RequireUp(t, "/health")

	c.RunCases(t, []e2e.Case{
		{Name: "create_feedback", Method: http.MethodPost, Path: "/api/v1/feedback", Body: map[string]any{
			"title": "E2E feedback", "description": "from e2e test", "categoryId": "",
		}, WantCode: []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest, http.StatusTooManyRequests}},
		{Name: "admin_stats", Method: http.MethodGet, Path: "/api/v1/admin/stats", WantCode: []int{http.StatusOK, http.StatusForbidden, http.StatusUnauthorized, http.StatusNotFound}},
		{Name: "admin_categories", Method: http.MethodGet, Path: "/api/v1/categories", WantCode: []int{http.StatusOK, http.StatusUnauthorized, http.StatusTooManyRequests}},
	})
}
