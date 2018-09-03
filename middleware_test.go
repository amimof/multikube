package multikube_test

import (
	"gitlab.com/amimof/multikube"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	w.Write([]byte(`{"alive": true}`))
}

func TestMiddleware(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(testHandler)
	handler.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

}

// Test the logging middleware. Should print output to the console
func TestMiddlewareWithLogging(t *testing.T) {
	handler := multikube.WithLogging(http.HandlerFunc(testHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Check the response body is what we expect.
	expected := `{"alive": true}`
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", w.Body.String(), expected)
	}

}

// Test the empty middleware. Shouldn't do anything
func TestMiddlewareWithEmpty(t *testing.T) {
	handler := multikube.WithEmpty(http.HandlerFunc(testHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Check the response body is what we expect.
	expected := `{"alive": true}`
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", w.Body.String(), expected)
	}

}
