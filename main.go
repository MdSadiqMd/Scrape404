package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MdSadiqMd/Scrape404/package/middleware"
	"github.com/MdSadiqMd/Scrape404/package/types"
	"github.com/MdSadiqMd/Scrape404/package/utils"
	"github.com/MdSadiqMd/Scrape404/package/worker"
	"github.com/fatih/color"
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

func checkLink(link, currentPage, linkType string, deadLinks *[]types.DeadLink, infoColor, successColor, errorColor *color.Color) {
	infoColor.Printf("  Found %s: %s\n", linkType, link)

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Use HEAD request first (faster), fall back to GET if needed
	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		*deadLinks = append(*deadLinks, types.DeadLink{
			URL:        link,
			StatusCode: 0,
			FoundOn:    currentPage,
			Type:       linkType,
		})
		errorColor.Printf("❌ Dead %s found: %s (Request Error: %s)\n", linkType, link, err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		*deadLinks = append(*deadLinks, types.DeadLink{
			URL:        link,
			StatusCode: 0,
			FoundOn:    currentPage,
			Type:       linkType,
		})
		errorColor.Printf("❌ Dead %s found: %s (Network Error: %s)\n", linkType, link, err)
		return
	}
	defer resp.Body.Close()

	// Some servers don't support HEAD requests, try GET if we get Method Not Allowed
	if resp.StatusCode == http.StatusMethodNotAllowed {
		req, err = http.NewRequest("GET", link, nil)
		if err != nil {
			*deadLinks = append(*deadLinks, types.DeadLink{
				URL:        link,
				StatusCode: 0,
				FoundOn:    currentPage,
				Type:       linkType,
			})
			errorColor.Printf("❌ Dead %s found: %s (Request Error: %s)\n", linkType, link, err)
			return
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		resp, err = client.Do(req)
		if err != nil {
			*deadLinks = append(*deadLinks, types.DeadLink{
				URL:        link,
				StatusCode: 0,
				FoundOn:    currentPage,
				Type:       linkType,
			})
			errorColor.Printf("❌ Dead %s found: %s (Network Error: %s)\n", linkType, link, err)
			return
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode >= 400 {
		*deadLinks = append(*deadLinks, types.DeadLink{
			URL:        link,
			StatusCode: resp.StatusCode,
			FoundOn:    currentPage,
			Type:       linkType,
		})
		errorColor.Printf("❌ Dead %s found: %s (Status: %d)\n", linkType, link, resp.StatusCode)
	} else {
		successColor.Printf("✓ Valid %s: %s\n", linkType, link)
	}
}

func parseURL(rawURL string) (*url.URL, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}
	return url.Parse(rawURL)
}

func sameHost(link, baseURL string) bool {
	linkURL, err := parseURL(link)
	if err != nil {
		return false
	}

	baseURLParsed, err := parseURL(baseURL)
	if err != nil {
		return false
	}

	return linkURL.Hostname() == baseURLParsed.Hostname()
}

func printResults(deadLinks []types.DeadLink, visitedLinks map[string]bool, visitedPages int, duration time.Duration, titleColor, errorColor *color.Color) {
	titleColor.Printf("\n=== Scan Summary ===\n")
	fmt.Printf("Pages visited: %d\n", visitedPages)
	fmt.Printf("Total links checked: %d\n", len(visitedLinks))
	fmt.Printf("Scan duration: %s\n", duration)
	fmt.Printf("Dead links found: %d\n", len(deadLinks))

	if len(deadLinks) == 0 {
		titleColor.Println("\n✓ No dead links found!")
		return
	}

	titleColor.Printf("\n=== Dead Links (%d) ===\n\n", len(deadLinks))

	fmt.Println("+----------------------+--------+----------+----------------------+")
	fmt.Println("| Dead Link            | Status | Type     | Found On             |")
	fmt.Println("+----------------------+--------+----------+----------------------+")

	for _, link := range deadLinks {
		statusText := "ERROR"
		if link.StatusCode > 0 {
			statusText = strconv.Itoa(link.StatusCode)
		}
		deadLinkDisplay := truncateString(link.URL, 20)
		foundOnDisplay := truncateString(link.FoundOn, 20)
		fmt.Printf("| %-20s | %-6s | %-8s | %-20s |\n", deadLinkDisplay, statusText, link.Type, foundOnDisplay)
	}
	fmt.Println("+----------------------+--------+----------+----------------------+")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
