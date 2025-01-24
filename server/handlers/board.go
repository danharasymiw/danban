package handlers

import (
	"errors"
	"net/http"

	"github.com/danharasymiw/danban/server/store"
	"github.com/danharasymiw/danban/server/ui/views"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) HandleBoard(w http.ResponseWriter, r *http.Request) {
	boardName := chi.URLParam(r, "boardName")

	board, err := h.storage.GetBoard(r.Context(), boardName)
	if errors.Is(err, &store.NoQueryResultsError{}) {
		board = &store.Board{
			Name: boardName,
			Columns: []*store.Column{
				{
					Index: 0,
					Name:  "To do",
					Cards: []*store.Card{
						{
							Id:          "id",
							Index:       0,
							Title:       "Make cards",
							Description: "This is a new board, make some cards!",
						},
					},
				},
				{
					Index: 1,
					Name:  "In Progress",
					Cards: []*store.Card{},
				},
				{
					Index: 2,
					Name:  "Done",
					Cards: []*store.Card{},
				},
			},
		}
		h.storage.AddBoard(r.Context(), board)
	}

	views.Board(board).Render(r.Context(), w)
}
