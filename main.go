package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/MdSadiqMd/Scrape404/package/middleware"
	"github.com/go-chi/chi"
)

type Config struct {
	Pages     []Page     `yaml:"pages"`
	Status    []Status   `yaml:"statuses"`
	Redirects []Redirect `yaml:"redirects"`
}

type Link struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Redirect struct {
	Path string `yaml:"path"`
	To   string `yaml:"to"`
}

type Page struct {
	Path  string `yaml:"path"`
	Title string `yaml:"title"`
	Links []Link `yaml:"links"`
	Extra *Link  `yaml:"extra"`
}

type Status struct {
	Path       string `yaml:"path"`
	StatusCode int    `yaml:"status"`
}

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func handleError(tmpl *template.Template, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		if wrapped.statusCode >= 400 {
			tmpl.ExecuteTemplate(w, "error.html", map[string]string{
				"ErrorMessage": http.StatusText(wrapped.statusCode),
				"Status":       strconv.Itoa(wrapped.statusCode),
			})
		}
	})
}

func main() {
	fmt.Println("Hello World")
	r := chi.NewRouter()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	r.Use(middleware.Logging(logger))
}
