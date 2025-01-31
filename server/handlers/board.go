package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/danharasymiw/danban/server/logger"
	"github.com/danharasymiw/danban/server/store"
	"github.com/danharasymiw/danban/server/ui/views"
)

func (h *Handler) HandleBoard(w http.ResponseWriter, r *http.Request) {
	boardName := chi.URLParam(r, "boardName")

	logEntry := logger.New(r.Context()).WithField("board", boardName)
	logEntry.Infof("Received get board request")

	if len(boardName) <= 3 || len(boardName) > 32 {
		handleError(logEntry, "invalid board name length", w, store.NewBadRequestError("board name must be between 3 and 32 characters"))
		return
	}

	board, err := h.storage.GetBoard(r.Context(), boardName)
	if err != nil {
		if _, ok := err.(*store.NotFoundError); ok {
			logEntry.Info("Board not found, creating new board")
			board, err = h.createNewBoard(r.Context(), boardName)
			if err != nil {
				handleError(logEntry, "failed to create board", w, err)
				return
			}
		} else {
			handleError(logEntry, "failed to get board", w, err)
			return
		}
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

type moveCardRequest struct {
	CardId     string `json:"cardId"`
	NewIndex   int    `json:"newIndex"`
	ToColumnId string `json:"toColumnId"`
}

func (h *Handler) HandleMoveCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	boardName := chi.URLParam(r, "boardName")

	logEntry := logger.New(ctx).WithField("board", boardName)
	logEntry.Infof("Received move card request")

	var req moveCardRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		handleError(logEntry, "error decoding request", w, store.NewBadRequestError(fmt.Sprintf("error decoding request: %s", err)))
		return
	}
	logEntry = logEntry.WithFields(logrus.Fields{
		"to column": req.ToColumnId,
		"card id":   req.CardId,
		"new index": req.NewIndex,
	})
	logEntry.Info("Found move card args")

	if err = h.storage.MoveCard(ctx, req.ToColumnId, req.CardId, req.NewIndex); err != nil {
		handleError(logEntry, "error moving card in storage", w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
