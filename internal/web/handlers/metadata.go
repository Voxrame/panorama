package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lord-server/panorama/internal/config"
)

func Metadata(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"title":      config.Web.Title,
			"zoomLevels": config.Renderer.ZoomLevels,
		})
	}
}
