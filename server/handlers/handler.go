package handlers

import (
	"context"
	"errors"
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

func thatWasAnError(ctx context.Context, w http.ResponseWriter, msg string, err error) bool {
	log := logger.New(ctx)
	if err != nil {
		log.WithError(err).Errorf("%s: %w", msg, err)

		var badRequest *store.BadRequestError
		var notFound *store.NotFoundError

		if errors.As(err, &badRequest) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errors.As(err, &notFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return true
	}
	return false
}
