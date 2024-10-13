package comunication

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.NilError(t, err)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(rootHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Body.String(), "Hello from Macpassd!")
}
