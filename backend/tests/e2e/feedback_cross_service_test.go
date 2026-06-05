//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/minisource/go-common/testing/e2e"
)

func TestFeedback_CommentCrossService(t *testing.T) {
	authURL := e2e.BaseURLFromEnv("AUTH_BASE_URL", "http://127.0.0.1:9001")
	feedbackURL := e2e.BaseURLFromEnv("FEEDBACK_BASE_URL", "http://127.0.0.1:5012")
	commentURL := e2e.BaseURLFromEnv("COMMENT_BASE_URL", "http://127.0.0.1:5010")

	token := e2e.LoginAuth(t, authURL, "admin@example.com", "AdminPass123!")
	h := e2e.Bearer(token)
	h["X-Tenant-ID"] = "default"

	feedback := e2e.NewClient(feedbackURL, h)
	feedback.RequireUp(t, "/health")

	title := fmt.Sprintf("e2e-comment-%d", time.Now().UnixNano())
	resp, body, err := feedback.Do(http.MethodPost, "/api/v1/feedback", map[string]any{
		"title": title, "description": "feedback for comment cross-service test",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated, http.StatusTooManyRequests)
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Skip("feedback rate limit exceeded")
	}
	feedbackID := e2e.ExtractID(t, body)

	resp, body, err = feedback.Do(http.MethodPost, "/api/v1/feedback/"+feedbackID+"/comments", map[string]any{
		"content": "e2e comment via feedback→comment service",
	})
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK, http.StatusCreated)

	resp, body, err = feedback.Do(http.MethodGet, "/api/v1/feedback/"+feedbackID+"/comments", nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)

	comment := e2e.NewClient(commentURL, h)
	comment.RequireUp(t, "/health")
	resp, body, err = comment.Do(http.MethodGet, "/api/v1/comments?resourceType=feedback&resourceId="+feedbackID, nil)
	if err != nil {
		t.Fatal(err)
	}
	e2e.ExpectStatus(t, resp, body, http.StatusOK)
}
