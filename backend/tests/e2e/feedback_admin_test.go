//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func feedbackAdminClient(t *testing.T) *e2e.Client {
	t.Helper()
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	h := e2e.Bearer(token)
	h["X-Tenant-ID"] = "default"
	c := e2e.NewClient(e2e.BaseURLFromEnv("FEEDBACK_BASE_URL", "http://127.0.0.1:5012"), h)
	c.RequireUp(t, "/health")
	return c
}

func TestFeedback_AdminStatsAndSettings(t *testing.T) {
	c := feedbackAdminClient(t)

	resp, body, err := c.Do(http.MethodGet, "/api/v1/admin/stats", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks feedback admin access")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/settings", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/feedback/pending", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)
}

func TestFeedback_AdminCategoryCRUD(t *testing.T) {
	c := feedbackAdminClient(t)
	suffix := time.Now().UnixNano()
	name := fmt.Sprintf("e2e-fb-cat-%d", suffix)

	resp, body, err := c.Do(http.MethodPost, "/api/v1/admin/categories", map[string]any{
		"name": name, "description": "e2e feedback category",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks feedback admin access")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)
	catID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/categories/"+catID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPut, "/api/v1/admin/categories/"+catID, map[string]any{
		"name": name, "description": "updated category",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/categories", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/admin/categories/"+catID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}

func TestFeedback_AdminModerationRejectAndStatus(t *testing.T) {
	c := feedbackAdminClient(t)
	suffix := time.Now().UnixNano()

	resp, body, err := c.Do(http.MethodPost, "/api/v1/feedback", map[string]any{
		"title": fmt.Sprintf("mod-reject-%d", suffix), "description": "moderation reject test",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusTooManyRequests)
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Skip("feedback rate limit exceeded")
	}
	fbID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/admin/feedback/"+fbID+"/reject", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusForbidden {
		t.Skip("admin lacks feedback admin access")
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPut, "/api/v1/admin/feedback/"+fbID+"/status", map[string]any{
		"status": "closed",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/feedback/"+fbID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}
