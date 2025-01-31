package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/danharasymiw/danban/server/logger"
	"github.com/danharasymiw/danban/server/store"
	"github.com/danharasymiw/danban/server/ui/components"
)

func (h *Handler) AddCard(w http.ResponseWriter, r *http.Request) {
	boardName := chi.URLParam(r, "boardName")
	columnId := chi.URLParam(r, "columnId")

	logEntry := logger.New(r.Context()).WithFields(
		logrus.Fields{
			"board name": boardName,
			"column id":  columnId,
		})
	logEntry.Infof("Received add card request")

	title := r.FormValue("title")
	if len(title) < 3 || len(title) > 250 {
		handleError(logEntry, "invalid card length", w, store.NewBadRequestError("card title must be between 3 and 250 characters"))
		return
	}

	card, err := h.storage.AddCard(r.Context(), columnId, title)
	if err != nil {
		handleError(logEntry, "error adding card", w, err)
		return
	}

	components.CardComponent(card).Render(r.Context(), w)
}

func (h *Handler) EditCard(w http.ResponseWriter, r *http.Request) {
	boardName := chi.URLParam(r, "boardName")
	columnId := chi.URLParam(r, "columnId")
	cardId := chi.URLParam(r, "cardId")

	logEntry := logger.New(r.Context()).WithFields(
		logrus.Fields{
			"board name": boardName,
			"column id":  columnId,
			"card id":    cardId,
		})

	logEntry.Infof("Received get edit card modal request")
	card, err := h.storage.GetCard(r.Context(), cardId)
	if err != nil {
		handleError(logEntry, "error getting card from storage", w, err)
		return
	}

	columns, err := h.storage.GetColumns(r.Context(), columnId)
	components.EditCardModal(card, boardName, columnId, columns).Render(r.Context(), w)
}
