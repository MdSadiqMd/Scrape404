package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/MdSadiqMd/Scrape404/package/middleware"
	"github.com/MdSadiqMd/Scrape404/package/utils"
	"github.com/go-chi/chi/v5"
)

func StartServer(port string) {
	r := chi.NewRouter()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	r.Use(middleware.Logging(logger))
	r.Use(middleware.NoCache)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Dead Link Checker API"))
	})
	r.Route("/api", func(r chi.Router) {
		r.Get("/check", utils.HandleCheckURL)
		r.Post("/check", utils.HandleSubmitURL)
	})

	fmt.Printf("Starting server on port %s...\n", port)
	http.ListenAndServe(":"+port, r)
}
