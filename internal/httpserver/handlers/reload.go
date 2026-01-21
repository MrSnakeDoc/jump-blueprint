package handlers

import (
	"net/http"

	"github.com/MrSnakeDoc/jump/internal/httpserver/deps"
	"github.com/MrSnakeDoc/jump/internal/logger"
)

// Reload triggers a manual reload of services and bookmarks
func Reload(d deps.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Trigger immediate reload for services
		servicesTriggered := false
		select {
		case d.ReloadTrigger <- struct{}{}:
			servicesTriggered = true
			d.Logger.Info("manual services reload triggered via endpoint",
				logger.String("remote_ip", r.RemoteAddr))
		default:
			d.Logger.Warn("services reload already in progress",
				logger.String("remote_ip", r.RemoteAddr))
		}

		// Trigger immediate reload for bookmarks (if enabled)
		bookmarksTriggered := false
		if d.BookmarkReloadTrigger != nil {
			select {
			case d.BookmarkReloadTrigger <- struct{}{}:
				bookmarksTriggered = true
				d.Logger.Info("manual bookmarks reload triggered via endpoint",
					logger.String("remote_ip", r.RemoteAddr))
			default:
				d.Logger.Warn("bookmarks reload already in progress",
					logger.String("remote_ip", r.RemoteAddr))
			}
		}

		// Determine response based on what was triggered
		if servicesTriggered || bookmarksTriggered {
			w.WriteHeader(http.StatusAccepted)
			if _, err := w.Write([]byte("✅ Reload triggered successfully\n")); err != nil {
				d.Logger.Debug("failed to write response", logger.Error(err))
			}
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write([]byte("⏳ Reload already in progress, please wait\n")); err != nil {
				d.Logger.Debug("failed to write response", logger.Error(err))
			}
		}
	}
}
