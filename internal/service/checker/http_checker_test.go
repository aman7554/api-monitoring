package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"pulsewatch/internal/domain"
)

func TestHTTPChecker_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	chk := NewHTTPChecker()
	m := &domain.Monitor{
		URL:                ts.URL,
		Method:             "GET",
		ExpectedStatusCode: 200,
		ResponseKeyword:    "ok",
		AuthConfig:         []byte(`{"type":"bearer","token":"secret-token"}`),
		TimeoutSeconds:     5,
	}

	res := chk.Execute(context.Background(), m)
	if res.Status != domain.MonitorStatusUp {
		t.Fatalf("expected status UP, got %s (error: %s)", res.Status, res.ErrorMessage)
	}

	if res.StatusCode != 200 {
		t.Fatalf("expected status code 200, got %d", res.StatusCode)
	}
}

func TestHTTPChecker_KeywordNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"maintenance"}`))
	}))
	defer ts.Close()

	chk := NewHTTPChecker()
	m := &domain.Monitor{
		URL:                ts.URL,
		Method:             "GET",
		ExpectedStatusCode: 200,
		ResponseKeyword:    "healthy",
		TimeoutSeconds:     5,
	}

	res := chk.Execute(context.Background(), m)
	if res.Status != domain.MonitorStatusDegraded {
		t.Fatalf("expected status DEGRADED for missing keyword, got %s", res.Status)
	}
}
