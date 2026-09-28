package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/processor/stub"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := New(
		stub.NewNotImplementedPhotoProcessor(),
		stub.NewNotImplementedPDFMerger(),
		nil, // storage.Client: unused by the routes exercised in these tests
		time.Hour,
		15,
	)
	r := gin.New()
	h.Register(r)
	return r
}

func doRequest(t *testing.T, r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthz(t *testing.T) {
	r := newTestRouter()
	w := doRequest(t, r, http.MethodGet, "/healthz")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPresets(t *testing.T) {
	r := newTestRouter()
	w := doRequest(t, r, http.MethodGet, "/api/v1/presets")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, name := range []string{"ocsc", "passport", "teacher"} {
		if !strings.Contains(body, name) {
			t.Errorf("expected preset %q in response body: %s", name, body)
		}
	}
}

func TestSelftest_ReturnsNotImplementedWhileProcessorIsStub(t *testing.T) {
	r := newTestRouter()
	w := doRequest(t, r, http.MethodGet, "/api/v1/selftest")
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", w.Code, w.Body.String())
	}
}
