package handlers

import (
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
