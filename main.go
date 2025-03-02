package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/MdSadiqMd/Scrape404/package/middleware"
	"github.com/go-chi/chi"
)

func main() {
	fmt.Println("Hello World")
	r := chi.NewRouter()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	r.Use(middleware.Logging(logger))
}
