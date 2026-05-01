package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
    // 1. Create a dummy request to pass to our handler
    req := httptest.NewRequest(http.MethodGet, "/health", nil)

    // 2. We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
    rr := httptest.NewRecorder()

    // 3. Call the handler directly, passing in the recorder and request
    health(rr, req)

    // 4. Check the status code is what we expect
    if rr.Code != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
    }

    // 5. Check the response body is what we expect
    expectedBody := "OK"
    bodyBytes, err := io.ReadAll(rr.Body)
    if err != nil {
        t.Fatalf("could not read response body: %v", err)
    }

    if string(bodyBytes) != expectedBody {
        t.Errorf("handler returned unexpected body: got %v want %v", string(bodyBytes), expectedBody)
    }

    // 6. Check the Content-Type header
    expectedContentType := "text/plain"
    if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
        t.Errorf("handler returned wrong content-type: got %v want %v", contentType, expectedContentType)
    }
}