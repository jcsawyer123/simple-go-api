package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/jcsawyer123/simple-go-api/internal/auth"
)

type Handlers struct {
	auth    auth.Service
	bufPool *sync.Pool // buffer pool for JSON encoding
}

func New(auth auth.Service) *Handlers {
	return &Handlers{
		auth: auth,
		bufPool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	buf := h.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		h.bufPool.Put(buf)
	}()

	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	return err
}

func (h *Handlers) GetData(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "data endpoint",
		"timestamp": time.Now().UTC(),
	}

	if err := h.writeJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "healthcheck endpoint",
		"timestamp": time.Now().UTC(),
	}

	if err := h.writeJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

type TestPermissionsResponse struct {
	HasPermission bool   `json:"has_permission"`
	Message       string `json:"message"`
}

func (h *Handlers) TestPermissions(w http.ResponseWriter, r *http.Request) {
	response := TestPermissionsResponse{
		HasPermission: true,
		Message:       "Permission check completed",
	}
	h.writeJSON(w, http.StatusOK, response)
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
