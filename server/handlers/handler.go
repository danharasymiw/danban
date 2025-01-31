package handlers

import (
	"net/http"

	"github.com/danharasymiw/danban/server/store"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	storage store.Storage
}

func NewHandler(storage store.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func handleError(log *logrus.Entry, msg string, w http.ResponseWriter, err error) {
	if err != nil {
		logrus.WithError(err).Errorf("%s: %w", msg, err)
		if _, ok := err.(store.BadRequestError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
