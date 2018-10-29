package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandlerReturnsOK(t *testing.T) {
	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)

	handler := HealthHandler{}
	handler.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "OK", string(rw.Body.Bytes()))
}
