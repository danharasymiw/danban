package handlers

import (
	"math/rand"
	"net/http"

	"github.com/danharasymiw/danban/server/ui/views"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	// Create a slice to hold the random characters
	var boardName []byte

	// Loop to generate random characters
	for i := 0; i < 10; i++ {
		// Generate a random index in the charset
		randomIndex := rand.Intn(len(charset))
		// Append the randomly selected character to the result
		boardName = append(boardName, charset[randomIndex])
	}

	views.Home(string(boardName)).Render(r.Context(), w)
}
