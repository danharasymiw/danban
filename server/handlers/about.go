package handlers

import (
	"net/http"

	"github.com/danharasymiw/danban/server/ui/views"
)

func (h *Handler) HandleAbout(w http.ResponseWriter, r *http.Request) {
	views.About().Render(r.Context(), w)
}
