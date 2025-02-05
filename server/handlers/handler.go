package handlers

import (
	"context"
	"net/http"

	"github.com/danharasymiw/danban/server/logger"
	"github.com/danharasymiw/danban/server/store"
)

type Handler struct {
	storage store.Storage
}

func NewHandler(storage store.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func handleError(ctx context.Context, msg string, w http.ResponseWriter, err error) {
	log := logger.New(ctx)
	if err != nil {
		log.WithError(err).Errorf("%s: %w", msg, err)
		if _, ok := err.(store.BadRequestError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
