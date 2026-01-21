package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/MrSnakeDoc/jump/internal/httpserver/deps"
)

type readyzResponse struct {
	Ready bool `json:"ready"`
}

func Readyz(d deps.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(readyzResponse{
			Ready: true,
		})
	}
}
