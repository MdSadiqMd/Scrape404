package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/MdSadiqMd/Scrape404/package/server"
	"github.com/MdSadiqMd/Scrape404/package/utils"
	"github.com/MdSadiqMd/Scrape404/package/worker"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	url := utils.PromptString(scanner, "Enter URL to scrape for dead links", "")
	if url == "" {
		fmt.Println("Error: URL cannot be empty")
		os.Exit(1)
	}

	depth := utils.PromptInt(scanner, "Enter maximum crawl depth", 5)
	delay := utils.PromptInt(scanner, "Enter delay between requests in milliseconds", 1000)
	parallel := utils.PromptInt(scanner, "Enter number of parallel scrapers", 2)
	timeout := utils.PromptInt(scanner, "Enter request timeout in seconds", 30)
	userAgent := utils.PromptString(scanner, "Enter user agent", "DeadLinkChecker/1.0")
	port := utils.PromptString(scanner, "Enter port for HTTP server", "8080")
	jsInput := utils.PromptString(scanner, "Use Playwright for JavaScript-enabled websites? (y/n)", "n")
	usePlaywright := strings.ToLower(jsInput) == "y" || strings.ToLower(jsInput) == "yes"

	go server.StartServer(port)

	if usePlaywright {
		worker.ScrapeWithPlaywright(url, depth, delay, parallel, timeout, userAgent)
	} else {
		worker.ScrapeWebsite(url, depth, delay, parallel, timeout, userAgent)
	}
}
