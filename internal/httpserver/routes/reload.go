package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/MrSnakeDoc/jump/internal/httpserver/deps"
	"github.com/MrSnakeDoc/jump/internal/httpserver/handlers"
	"github.com/MrSnakeDoc/jump/internal/httpserver/mw"
)

func init() { Register(registerReload) }

func registerReload(r chi.Router, d deps.Deps) {
	r.With(mw.AllowOnlyCIDRS(d.AllowedCIDRS, d.TrustProxy, d.Logger), mw.EnforceHost(d.AllowedHosts, d.Logger)).Post("/reload", handlers.Reload(d))
}
