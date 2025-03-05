package utils

import (
	"fmt"
	"net/http"
)

func HandleCheckURL(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing URL parameter", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Starting scan for URL: %s", url)
}

func HandleSubmitURL(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "Missing URL parameter", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Submitted URL for scanning: %s", url)
}
