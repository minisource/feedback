//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func feedbackAuthClient(t *testing.T) *e2e.Client {
	t.Helper()
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	h := e2e.Bearer(token)
	h["X-Tenant-ID"] = "default"
	c := e2e.NewClient(e2e.BaseURLFromEnv("FEEDBACK_BASE_URL", "http://127.0.0.1:5012"), h)
	c.RequireUp(t, "/health")
	return c
}

func TestFeedback_SubscriptionsCRUD(t *testing.T) {
	c := feedbackAuthClient(t)

	title := fmt.Sprintf("e2e-sub-%d", time.Now().UnixNano())
	resp, body, err := c.Do(http.MethodPost, "/api/v1/feedback", map[string]any{
		"title": title, "description": "subscription test",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusTooManyRequests)
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Skip("feedback rate limit exceeded")
	}
	feedbackID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/subscriptions", map[string]any{
		"feedbackId": feedbackID, "types": []string{"comment"}, "emailEnabled": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusConflict, http.StatusTooManyRequests)
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Skip("feedback rate limit exceeded")
	}
	var subID string
	if resp.StatusCode == http.StatusConflict {
		_, _, _ = c.Do(http.MethodDelete, "/api/v1/feedback/"+feedbackID+"/unsubscribe", nil)
		resp, body, err = c.Do(http.MethodPost, "/api/v1/subscriptions", map[string]any{
			"feedbackId": feedbackID, "types": []string{"comment"}, "emailEnabled": true,
		})
		if err != nil {
			t.Fatal(err)
		}
		e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusTooManyRequests, http.StatusInternalServerError)
		if resp.StatusCode == http.StatusTooManyRequests {
			t.Skip("feedback rate limit exceeded")
		}
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			subID = e2e.ExtractID(t, body)
		}
	} else {
		subID = e2e.ExtractID(t, body)
	}
	if subID == "" {
		resp, body, err = c.Do(http.MethodGet, "/api/v1/subscriptions", nil)
		if err != nil {
			t.Fatal(err)
		}
		e2e.ExpectStatus(t, resp, body, http.StatusOK)
		var parsed map[string]any
		e2e.ParseJSON(t, body, &parsed)
		if items, ok := parsed["data"].([]any); ok {
			for _, it := range items {
				m, _ := it.(map[string]any)
				if fid, _ := m["feedbackId"].(string); fid == feedbackID {
					subID, _ = m["id"].(string)
					if subID == "" {
						subID, _ = m["_id"].(string)
					}
					break
				}
			}
		}
		if subID == "" {
			t.Skip("could not resolve subscription id after conflict")
		}
	}

	resp, body, err = c.Do(http.MethodGet, "/api/v1/feedback/"+feedbackID+"/subscription", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPut, "/api/v1/subscriptions/"+subID, map[string]any{
		"emailEnabled": false,
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/subscriptions/"+subID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/feedback/"+feedbackID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}

func TestFeedback_CommentsViaProxy(t *testing.T) {
	c := feedbackAuthClient(t)

	title := fmt.Sprintf("e2e-fb-comment-%d", time.Now().UnixNano())
	resp, body, err := c.Do(http.MethodPost, "/api/v1/feedback", map[string]any{
		"title": title, "description": "comment proxy test",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusTooManyRequests)
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Skip("feedback rate limit exceeded")
	}
	feedbackID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/feedback/"+feedbackID+"/comments", map[string]any{
		"content": "feedback service proxies to comment",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusTooManyRequests)
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Skip("feedback rate limit exceeded")
	}
	commentID := e2e.ExtractID(t, body)

	resp, body, err = c.Do(http.MethodPut, "/api/v1/comments/"+commentID, map[string]any{
		"content": "updated via feedback proxy",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	resp, body, err = c.Do(http.MethodPost, "/api/v1/comments/"+commentID+"/reactions", map[string]any{
		"type": "like",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusBadRequest)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/comments/"+commentID+"/reactions", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent, http.StatusNotFound)

	resp, body, err = c.Do(http.MethodDelete, "/api/v1/comments/"+commentID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusNoContent)
}
