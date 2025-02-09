package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/danharasymiw/danban/server/handlers"
	"github.com/danharasymiw/danban/server/store/mdb"
)

func main() {
	storage := mdb.New()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	handler := handlers.NewHandler(storage)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var boardName []byte

		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		// Loop to generate random characters
		for i := 0; i < 12; i++ {
			// Generate a random index in the charset
			randomIndex := rand.Intn(len(charset))
			// Append the randomly selected character to the result
			boardName = append(boardName, charset[randomIndex])
		}

		redirectURL := fmt.Sprintf("/board/%s", boardName)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	r.Get("/board", func(w http.ResponseWriter, r *http.Request) {
		// Parse the query parameter 'name'
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
			return
		}

		redirectURL := fmt.Sprintf("/board/%s", name)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})
	r.Get("/board/{boardName}", handler.HandleBoard)

	r.Post("/board/{boardName}/moveCard", handler.HandleMoveCard)

	r.Post("/board/{boardName}/column/{columnId}/cards/add", handler.AddCard)

	r.Get("/board/{boardName}/column/{columnId}/card/{cardId}/edit", handler.EditCardView)
	r.Put("/board/{boardName}/column/{columnId}/card/{cardId}/edit", handler.UpdateCard)

	r.Delete("/board/{boardName}/column/{columnId}/card/{cardId}", handler.DeleteCard)

	r.Get("/about", handler.HandleAbout)

	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	isDeployed := os.Getenv("RAILWAY_PUBLIC_DOMAIN") != ``
	domain := "localhost"
	if isDeployed {
		domain = ""
	}

	println("Listening on: 8080")
	err := http.ListenAndServe(fmt.Sprintf("%s:8080", domain), r)
	if err != nil {
		panic(err)
	}
}
