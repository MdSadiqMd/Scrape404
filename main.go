package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/MdSadiqMd/Scrape404/package/middleware"
	"github.com/MdSadiqMd/Scrape404/package/utils"
	"github.com/MdSadiqMd/Scrape404/package/worker"
	"github.com/go-chi/chi/v5"
)

func main() {
	urlFlag := flag.String("url", "", "URL to scrape for dead links")
	depthFlag := flag.Int("depth", 5, "Maximum crawl depth (default: 5)")
	delayFlag := flag.Int("delay", 1000, "Delay between requests in milliseconds (default: 1000)")
	parallelismFlag := flag.Int("parallel", 2, "Number of parallel scrapers (default: 2)")
	timeoutFlag := flag.Int("timeout", 10, "Request timeout in seconds (default: 10)")
	userAgentFlag := flag.String("user-agent", "DeadLinkChecker/1.0", "User agent to use for requests")
	portFlag := flag.String("port", "8080", "Port to run the HTTP server on")
	flag.Parse()

	go startServer(*portFlag)

	if *urlFlag == "" {
		fmt.Println("Please provide a URL to scrape with --url flag")
		flag.Usage()
		os.Exit(1)
	}

	worker.ScrapeWebsite(*urlFlag, *depthFlag, *delayFlag, *parallelismFlag, *timeoutFlag, *userAgentFlag)
}

func startServer(port string) {
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
