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

	if len(boardName) > 32 {
		logEntry.Errorf("Board name too long")
		w.WriteHeader(400)
		w.Write([]byte("board name cannot be longer than 32 characters"))
		return
	}

	board, err := h.storage.GetBoard(r.Context(), boardName)
	if err != nil {
		if _, ok := err.(*store.NotFoundError); ok {
			logEntry.Info("Board not found, creating new board")
			board, err = h.createNewBoard(r.Context(), boardName)
			if err != nil {
				logEntry.Errorf("Failed to create board ", err)
				http.Error(w, "Failed to create board", http.StatusInternalServerError)
				return
			}
		} else {
			logEntry.Errorf("Failed to get board: %s", err)
			http.Error(w, "Failed to get board", http.StatusInternalServerError)
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
	CardId       string `json:"cardId"`
	NewIndex     int    `json:"newIndex"`
	FromColumnId string `json:"fromColumnId"`
	ToColumnId   string `json:"toColumnId"`
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
		http.Error(w, fmt.Sprintf("Error decoding request: %s", err), http.StatusBadRequest)
		return
	}
	logEntry = logEntry.WithFields(logrus.Fields{
		"from column": req.FromColumnId,
		"to column":   req.ToColumnId,
		"card id":     req.CardId,
		"new index":   req.NewIndex,
	})
	logEntry.Info("Found move card args")

	if err = h.storage.MoveCard(ctx, boardName, req.FromColumnId, req.ToColumnId, req.CardId, req.NewIndex); err != nil {
		logEntry.Error("error moving card: ", err)
		if _, ok := err.(store.BadRequestError); ok {
			http.Error(w, fmt.Sprintf("Error moving card: ", err), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal error moving card", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
