package v0

import (
	"net/http"

	"github.com/otiai10/marmoset"
)

// Status for `/v0/status`
func Status(w http.ResponseWriter, r *http.Request) {
	render := marmoset.Render(w)
	render.JSON(http.StatusOK, marmoset.P{
		"status":  "ok",
		"version": "0.0.11",
	})
}
