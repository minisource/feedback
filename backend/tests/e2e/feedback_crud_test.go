//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/minisource/go-common/testing/e2e"
)

func TestFeedback_CRUDFlow(t *testing.T) {
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	h := e2e.Bearer(token)
	h["X-Tenant-ID"] = "default"

	c := e2e.NewClient(e2e.BaseURLFromEnv("FEEDBACK_BASE_URL", "http://127.0.0.1:5012"), h)
	c.RequireUp(t, "/health")

	resp, body, err := c.Do(http.MethodPost, "/api/v1/feedback", map[string]any{
		"title": "E2E CRUD feedback", "description": "full flow test",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusTooManyRequests)
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Skip("feedback rate limit exceeded")
	}

	var created map[string]any
	e2e.ParseJSON(t, body, &created)
	id := e2e.GetString(created, "data", "id")
	if id == "" {
		id = e2e.GetString(created, "data", "_id")
	}
	if id == "" {
		t.Fatalf("no feedback id in create response: %s", string(body))
	}

	resp, body, err = c.Do(http.MethodGet, "/api/v1/feedback/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPut, "/api/v1/feedback/"+id, map[string]any{
		"title": "E2E CRUD feedback updated",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/feedback/"+id+"/vote", map[string]any{
		"type": "up",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusBadRequest)

	resp, body, err = c.Do(http.MethodGet, "/api/v1/admin/stats", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusForbidden)
	if resp.StatusCode == http.StatusForbidden {
		t.Log("admin stats requires feedback admin scope")
	}

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/feedback/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}
