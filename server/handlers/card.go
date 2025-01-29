package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/danharasymiw/danban/server/logger"
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
		logEntry.WithField("title length", len(title)).Info("Card title too short or too long")
		w.WriteHeader(400)
		w.Write([]byte("card title must be between 3 and 250 characters"))
		return
	}

	card, err := h.storage.AddCard(r.Context(), boardName, columnId, title)
	if err != nil {
		logEntry.Errorf(fmt.Sprintf("Unexpected error adding card: %w", err))
		http.Error(w, "Error adding card", http.StatusInternalServerError)
		return
	}

	components.CardComponent(card).Render(r.Context(), w)
}

func (h *Handler) EditCard(w http.ResponseWriter, r *http.Request) {

}
