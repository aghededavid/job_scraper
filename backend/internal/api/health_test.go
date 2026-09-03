package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestRouter returns a minimal gin router wired with just the health endpoint.
// This mirrors the setup in main.go without requiring a real database or Redis.
func newTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "remotescope-api"})
	})

	return r
}

// TestHealthEndpoint verifies the /health route returns HTTP 200 and the
// expected JSON body.
func TestHealthEndpoint(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
	if body["service"] != "remotescope-api" {
		t.Errorf("expected service=remotescope-api, got %q", body["service"])
	}
}

// TestHealthEndpointMethod verifies POST to /health returns 405 Method Not Allowed.
func TestHealthEndpointMethod(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
		t.Fatalf("expected 405 or 404 for wrong method, got %d", w.Code)
	}
}

// TestNotFoundRoute verifies unknown routes return 404.
func TestNotFoundRoute(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown route, got %d", w.Code)
	}
}
