package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/emojify-api/emojify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var fileURL = "http://something.com/a.jpg"
var base64URL string

func setupCacheHandler() (*httptest.ResponseRecorder, *http.Request, *mux.Router) {
	mockCache = emojify.MockCache{}
	logger := hclog.Default()
	base64URL = base64.StdEncoding.EncodeToString([]byte(fileURL))

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(
		"GET",
		"/"+base64URL,
		nil,
	)

	h := &CacheHandler{logger, &mockCache}
	router := mux.NewRouter()
	router.HandleFunc("/{image}", h.ServeHTTP).Methods("GET")

	return rw, r, router
}

func TestReturns405WhenNotGet(t *testing.T) {
	rw, _, h := setupCacheHandler()
	r := httptest.NewRequest("POST", "/", nil)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusNotFound, rw.Code)
}

func TestReturns400WhenInvalidFileParameter(t *testing.T) {
	rw, _, h := setupCacheHandler()
	r := httptest.NewRequest("GET", "/", nil)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusNotFound, rw.Code)
}

func TestReturns404WhenNoImageFoundInCache(t *testing.T) {
	rw, r, h := setupCacheHandler()
	mockCache.On("Get", mock.Anything).Return([]byte{}, fmt.Errorf("Not found"))
	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusNotFound, rw.Code)
}

func TestReturns200WhenImageFound(t *testing.T) {
	rw, r, h := setupCacheHandler()
	mockCache.On("Get", mock.Anything).Return([]byte("abc"), nil)

	h.ServeHTTP(rw, r)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "abc", rw.Body.String())
}
