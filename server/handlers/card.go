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
	if err != nil {
		handleError(ctx, "invalid card length", w, err)
		return
	}

	card, err := h.storage.AddCard(r.Context(), columnId, title)
	if err != nil {
		handleError(ctx, "error adding card", w, err)
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
	if err != nil {
		handleError(ctx, "error getting card from storage", w, err)
		return
	}

	columns, err := h.storage.GetColumns(r.Context(), boardName)
	if err != nil {
		handleError(ctx, "error getting board columns", w, err)
	}
	components.EditCardModal(boardName, columnId, card, columns).Render(r.Context(), w)
}

func (h *Handler) UpdateCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	boardName := chi.URLParam(r, "boardName")
	columnId := chi.URLParam(r, "columnId")
	cardId := chi.URLParam(r, "cardId")

	card, err := h.storage.GetCard(r.Context(), cardId)
	if err != nil {
		handleError(ctx, "error getting card from storage", w, err)
		return
	}

	card.Title, err = getFormCardTitle(r, w)
	if err != nil {
		handleError(ctx, "invalid title", w, err)
		return
	}

	card.Description, err = getFormCardDescription(r, w)
	if err != nil {
		handleError(ctx, "invalid description", w, err)
		return
	}

	if err := h.storage.EditCard(r.Context(), card); err != nil {
		handleError(ctx, "error editing card", w, err)
		return
	}

	// User updated the card, we need to move it.
	if r.FormValue("columnChanged") == "true" {
		newColumnId := r.FormValue("toColumnId")
		err := h.storage.MoveCard(r.Context(), newColumnId, cardId, -1)
		if err != nil {
			handleError(ctx, "error moving card from edit card modal", w, err)
		}
	} else {
		components.CardComponent(boardName, columnId, card).Render(r.Context(), w)
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
