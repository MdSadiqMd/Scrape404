package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MdSadiqMd/Scrape404/package/server"
	"github.com/MdSadiqMd/Scrape404/package/worker"
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

	go server.StartServer(*portFlag)

	if *urlFlag == "" {
		fmt.Println("Please provide a URL to scrape with --url flag")
		flag.Usage()
		os.Exit(1)
	}

	worker.ScrapeWebsite(*urlFlag, *depthFlag, *delayFlag, *parallelismFlag, *timeoutFlag, *userAgentFlag)
}
