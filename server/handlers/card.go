package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/danharasymiw/danban/server/constants"
	"github.com/danharasymiw/danban/server/store"
	"github.com/danharasymiw/danban/server/ui/components"
)

func (h *Handler) AddCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	boardName := chi.URLParam(r, "boardName")
	columnId := chi.URLParam(r, "columnId")

	title, err := getFormCardTitle(r, w)
	if thatWasAnError(ctx, w, "invalid card length", err) {
		return
	}

	card, err := h.storage.AddCard(r.Context(), columnId, title)
	if thatWasAnError(ctx, w, "error adding card", err) {
		return
	}

	components.CardComponent(boardName, columnId, card).Render(r.Context(), w)
}

func (h *Handler) EditCardView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	boardName := chi.URLParam(r, "boardName")
	columnId := chi.URLParam(r, "columnId")
	cardId := chi.URLParam(r, "cardId")

	card, err := h.storage.GetCard(r.Context(), cardId)
	if thatWasAnError(ctx, w, "error getting card from storage", err) {
		return
	}

	columns, err := h.storage.GetColumns(r.Context(), boardName)
	if thatWasAnError(ctx, w, "error getting board columns", err) {
		return
	}
	components.EditCardModal(boardName, columnId, card, columns).Render(r.Context(), w)
}

func (h *Handler) UpdateCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	boardName := chi.URLParam(r, "boardName")
	columnId := chi.URLParam(r, "columnId")
	cardId := chi.URLParam(r, "cardId")

	card, err := h.storage.GetCard(r.Context(), cardId)
	if thatWasAnError(ctx, w, "error getting card from storage", err) {
		return
	}

	card.Title, err = getFormCardTitle(r, w)
	if thatWasAnError(ctx, w, "invalid title", err) {
		return
	}

	card.Description, err = getFormCardDescription(r, w)
	if thatWasAnError(ctx, w, "invalid description", err) {
		return
	}

	err = h.storage.EditCard(r.Context(), card)
	if thatWasAnError(ctx, w, "error editing card", err) {
		return
	}

	// User updated the card, we need to move it.
	if r.FormValue("columnChanged") == "true" {
		newColumnId := r.FormValue("toColumnId")
		err := h.storage.MoveCard(r.Context(), newColumnId, cardId, -1)
		if thatWasAnError(ctx, w, "error moving card from edit card modal", err) {
			return
		}
		components.MovedCardComponent(boardName, newColumnId, card).Render(ctx, w)
	} else {
		components.CardComponent(boardName, columnId, card).Render(ctx, w)
	}

}

func getFormCardTitle(r *http.Request, w http.ResponseWriter) (string, error) {
	title := r.FormValue(`title`)
	if len(title) < constants.MinTitleLength || len(title) > constants.MaxTitleLength {
		return ``, store.NewBadRequestError(fmt.Sprintf(`title must be between %d and %d characters`, constants.MinTitleLength, constants.MaxTitleLength))
	}

	return title, nil
}

func getFormCardDescription(r *http.Request, w http.ResponseWriter) (string, error) {
	description := r.FormValue(`description`)
	if len(description) > constants.MaxDescriptionLength {
		return ``, store.NewBadRequestError(fmt.Sprintf(`description cannot exceed %d characters`, constants.MaxDescriptionLength))
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

func (h *Handler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	columnId := chi.URLParam(r, "columnId")
	cardId := chi.URLParam(r, "cardId")

	card, err := h.storage.GetCard(r.Context(), cardId)
	if thatWasAnError(ctx, w, "error getting card from storage", err) {
		return
	}

	err = h.storage.DeleteCard(ctx, columnId, cardId, card.Index)
	if thatWasAnError(ctx, w, "error deleting card", err) {
		return
	}

	w.WriteHeader(http.StatusOK)
}
