package logger

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func LogEntryMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(chi.URLParam(r, "id"))

		log := New(r.Context())

		if boardName := chi.URLParam(r, "boardName"); boardName != `` {
			log = log.WithField("board name", boardName)
		}

		if columnId := chi.URLParam(r, "columnId"); columnId != `` {
			log = log.WithField("column id", columnId)
		}

		if cardId := chi.URLParam(r, "cardId"); cardId != `` {
			log = log.WithField("card id", cardId)
		}

		r.WithContext(context.WithValue(r.Context(), ctxLogger, log))

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
