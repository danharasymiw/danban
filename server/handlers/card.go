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

	title, err := getFormCardTitle(r, w)
	if err != nil {
		handleError(logEntry, "invalid card length", w, err)
		return
	}

	card, err := h.storage.AddCard(r.Context(), columnId, title)
	if err != nil {
		handleError(logEntry, "error adding card", w, err)
		return
	}

	components.CardComponent(boardName, columnId, card).Render(r.Context(), w)
}

func (h *Handler) EditCardView(w http.ResponseWriter, r *http.Request) {
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

	columns, err := h.storage.GetColumns(r.Context(), boardName)
	if err != nil {
		handleError(logEntry, "error getting board columns", w, err)
	}
	components.EditCardModal(boardName, columnId, card, columns).Render(r.Context(), w)
}

func (h *Handler) UpdateCard(w http.ResponseWriter, r *http.Request) {
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

	card.Title, err = getFormCardTitle(r, w)
	if err != nil {
		handleError(logEntry, "invalid title", w, err)
		return
	}

	card.Description, err = getFormCardDescription(r, w)
	if err != nil {
		handleError(logEntry, "invalid description", w, err)
		return
	}

	if err := h.storage.EditCard(r.Context(), card); err != nil {
		handleError(logEntry, "error editing card", w, err)
		return
	}

	if r.FormValue("columnChanged") == "true" {
		newColumnId := r.FormValue("toColumnId")
		err := h.storage.MoveCard(r.Context(), newColumnId, cardId, -1)
		if err != nil {
			handleError(logEntry, "error moving card from edit card modal", w, err)
		}
	}

	components.CardComponent(boardName, columnId, card).Render(r.Context(), w)
}

func getFormCardTitle(r *http.Request, w http.ResponseWriter) (string, error) {
	title := r.FormValue(`title`)
	if len(title) < 3 || len(title) >= 32 {
		return ``, store.NewBadRequestError(`title must be between 4 and 32 characters`)
	}

	return title, nil
}

func getFormCardDescription(r *http.Request, w http.ResponseWriter) (string, error) {
	description := r.FormValue(`description`)
	if len(description) > 2048 {
		return ``, store.NewBadRequestError(`description cannot exceed 2048 characters`)
	}
	return description, nil
}

func getFormCard(r *http.Request, w http.ResponseWriter) (*store.Card, error) {
	title, err := getFormCardTitle(r, w)
	if err != nil {
		return nil, err
	}

	description, err := getFormCardDescription(r, w)
	if err != nil {
		return nil, err
	}

	return &store.Card{
		Title:       title,
		Description: description,
	}, nil

}
