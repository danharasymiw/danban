package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/danharasymiw/danban/server/logger"
	"github.com/danharasymiw/danban/server/store"
	"github.com/danharasymiw/danban/server/ui/views"
)

func (h *Handler) HandleBoard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	boardName := chi.URLParam(r, "boardName")

	log := logger.New(r.Context())
	log.Infof("Received get board request")

	if len(boardName) <= 3 || len(boardName) > 32 {
		thatWasAnError(r.Context(), w, "invalid board name length", store.NewBadRequestError("board name must be between 3 and 32 characters"))
		return
	}

	board, err := h.storage.GetBoard(ctx, boardName)
	if err != nil {
		if _, ok := err.(*store.NotFoundError); ok {
			log.Info("Board not found, creating new board")
			board, err = h.createNewBoard(ctx, boardName)
			if thatWasAnError(ctx, w, "failed to create board", err) {
				return
			}
		} else {
			thatWasAnError(ctx, w, "failed to get board", err)
			return
		}
	}

	log.Info("Board found, returning")
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

type moveCardRequest struct {
	CardId     string `json:"cardId"`
	NewIndex   int    `json:"newIndex"`
	ToColumnId string `json:"toColumnId"`
}

func (h *Handler) HandleMoveCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logEntry := logger.New(ctx)

	var req moveCardRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if thatWasAnError(ctx, w, "error decoding request", err) {
		return
	}
	logEntry = logEntry.WithFields(logrus.Fields{
		"to column": req.ToColumnId,
		"card id":   req.CardId,
		"new index": req.NewIndex,
	})
	logEntry.Info("Found move card args")

	err = h.storage.MoveCard(ctx, req.ToColumnId, req.CardId, req.NewIndex)
	if thatWasAnError(ctx, w, "error moving card in storage", err) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
