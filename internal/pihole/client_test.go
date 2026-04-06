package pihole

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClient_Get_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			writeJSON(w, authResponse{Session: sessionInfo{Valid: true, SID: "test-sid"}})
		case "/api/dns/blocking":
			if r.Header.Get("X-FTL-SID") != "test-sid" {
				t.Errorf("missing SID header")
			}
			writeJSON(w, BlockingStatus{Blocking: "enabled"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "test", WithHTTPClient(srv.Client()))

	var status BlockingStatus
	err := c.Get(context.Background(), "/dns/blocking", &status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Blocking != "enabled" {
		t.Errorf("Blocking = %q, want %q", status.Blocking, "enabled")
	}
}

func TestClient_AuthRetryOn401(t *testing.T) {
	var authCount atomic.Int32
	var reqCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			n := authCount.Add(1)
			writeJSON(w, authResponse{Session: sessionInfo{
				Valid: true,
				SID:   fmt.Sprintf("sid-%d", n),
			}})
		case "/api/dns/blocking":
			n := reqCount.Add(1)
			if n == 1 {
				// First request returns 401 to trigger re-auth.
				w.WriteHeader(http.StatusUnauthorized)
				writeJSON(w, errorResponse{Error: errorDetail{Key: "unauthorized", Message: "expired"}})
				return
			}
			writeJSON(w, BlockingStatus{Blocking: "enabled"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "test", WithHTTPClient(srv.Client()))

	var status BlockingStatus
	err := c.Get(context.Background(), "/dns/blocking", &status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Blocking != "enabled" {
		t.Errorf("Blocking = %q, want %q", status.Blocking, "enabled")
	}
	if authCount.Load() != 2 {
		t.Errorf("auth count = %d, want 2 (initial + retry)", authCount.Load())
	}
}

func TestClient_404_ReturnsNotFoundError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			writeJSON(w, authResponse{Session: sessionInfo{Valid: true, SID: "sid"}})
		default:
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, errorResponse{Error: errorDetail{Key: "not_found", Message: "not found"}})
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "test", WithHTTPClient(srv.Client()))

	err := c.Get(context.Background(), "/groups/nonexistent", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	var nfe *NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestClient_400_ReturnsValidationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			writeJSON(w, authResponse{Session: sessionInfo{Valid: true, SID: "sid"}})
		default:
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, errorResponse{Error: errorDetail{Key: "bad_request", Message: "invalid"}})
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "test", WithHTTPClient(srv.Client()))

	err := c.Post(context.Background(), "/domains/allow/exact", map[string]string{}, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestClient_Delete_204(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			writeJSON(w, authResponse{Session: sessionInfo{Valid: true, SID: "sid"}})
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "test", WithHTTPClient(srv.Client()))

	err := c.Delete(context.Background(), "/groups/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_AuthFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, authResponse{Session: sessionInfo{Valid: false, Message: "wrong password"}})
	}))
	defer srv.Close()

	c := New(srv.URL, "wrong", WithHTTPClient(srv.Client()))

	err := c.Get(context.Background(), "/dns/blocking", nil)
	if err == nil {
		t.Fatal("expected auth error")
	}

	var ae *AuthError
	if !errors.As(err, &ae) {
		t.Errorf("expected AuthError, got %T: %v", err, err)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
