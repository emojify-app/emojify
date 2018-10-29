package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/gorilla/mux"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/emojify-api/emojify"
)

// CacheHandler returns images from the cache
type CacheHandler struct {
	logger hclog.Logger
	cache  emojify.Cache
}

// NewCacheHandler creates a new http.Handler for dealing with cache requests
func NewCacheHandler(l hclog.Logger, c emojify.Cache) *CacheHandler {
	return &CacheHandler{l, c}
}

// Handle handles requests for cache
func (c *CacheHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	c.logger.Info("Cache called", "method", r.Method, "URI", r.URL.String())

	if r.Method != "GET" {
		c.logger.Info("Method not allowed", "method", r.Method)
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

	// check the parameters contains a valid url
	vars := mux.Vars(r)
	f := vars["image"]
	if f == "" {
		c.logger.Error("Image is a required parameter", "file", f)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// the image is expected to be base64 encoded so lets decode it
	data, err := base64.StdEncoding.DecodeString(f)
	if err != nil {
		c.logger.Error("Image is not base64 encoded string", "file", f, "error", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch the file from the cache
	image, err := c.cache.Get(string(data))
	if err != nil {
		c.logger.Info("File not found in cache", "url", string(data), "error", err)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	fileType := http.DetectContentType(image)

	c.logger.Info("Found file, returning", "file", f)

	// all ok return the file
	rw.Header().Add("content-type", fileType)
	rw.Write(image)
}
