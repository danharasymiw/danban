package handlers

import (
	"net/http"

	"github.com/danharasymiw/danban/server/ui/components"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) AddColumn(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) EditColumn(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) MoveColumn(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) GetColumn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	boardName := chi.URLParam(r, "boardName")
	columnId := chi.URLParam(r, "columnId")
	column, err := h.storage.GetColumn(ctx, columnId)
	if thatWasAnError(ctx, w, "unable to get column", err) {
		return
	}

	components.ColumnComponent(boardName, column).Render(r.Context(), w)
}
