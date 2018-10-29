package handlers

import "net/http"

// HealthHandler is a http handler for health checking the service
type HealthHandler struct{}

func (h *HealthHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("OK"))
}
