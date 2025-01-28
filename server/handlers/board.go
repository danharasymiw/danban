package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/danharasymiw/danban/server/logger"
	"github.com/danharasymiw/danban/server/store"
	"github.com/danharasymiw/danban/server/ui/views"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) HandleBoard(w http.ResponseWriter, r *http.Request) {
	boardName := chi.URLParam(r, "boardName")

	logEntry := logger.New(r.Context()).WithField("board", boardName)
	logEntry.Infof("Received board request")

	if len(boardName) > 32 {
		logEntry.Errorf("Board name too long")
		w.WriteHeader(400)
		w.Write([]byte("board name cannot be longer than 32 characters"))
		return
	}

	board, err := h.storage.GetBoard(r.Context(), boardName)
	if errors.Is(err, &store.NoQueryResultsError{}) {
		logEntry.Info("Board not found, creating new board")
		board, err = h.createNewBoard(r.Context(), boardName)
		if err != nil {
			logEntry.Errorf("Failed to create board")
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
	} else if err != nil {
		logEntry.Errorf("Unexpected error retrieving board")
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	logEntry.Info("Board found, returning")
	views.Board(board).Render(r.Context(), w)
}

func (h *Handler) createNewBoard(ctx context.Context, boardName string) (*store.Board, error) {
	board := &store.Board{
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
	return board, h.storage.AddBoard(ctx, board)
}
