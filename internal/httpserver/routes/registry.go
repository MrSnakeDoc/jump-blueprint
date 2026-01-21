package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MrSnakeDoc/jump/internal/httpserver/deps"
)

type (
	Registrar  func(r chi.Router, d deps.Deps)
	Middleware = func(http.Handler) http.Handler
)

type entry struct {
	reg Registrar
	mws []Middleware
}

var registry []entry

// Register a registrar with optional per-route middlewares.
func Register(reg Registrar, mws ...Middleware) {
	registry = append(registry, entry{reg: reg, mws: mws})
}

// Called once from server.New()
func RegisterAll(r chi.Router, d deps.Deps) {
	for _, e := range registry {
		if len(e.mws) == 0 {
			e.reg(r, d)
			continue
		}
		sub := r.With(e.mws...) // apply per-route middlewares
		e.reg(sub, d)
	}
}
