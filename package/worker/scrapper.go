package worker

import (
	"strings"
	"sync"
	"time"

	"github.com/MdSadiqMd/Scrape404/package/types"
	"github.com/fatih/color"
	"github.com/gocolly/colly"
)

func ScrapeWebsite(urlStr string, maxDepth, delayMs, parallelism, timeoutSec int, userAgent string) {
	titleColor := color.New(color.FgCyan, color.Bold)
	successColor := color.New(color.FgGreen)
	errorColor := color.New(color.FgRed)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgBlue)

	titleColor.Println("\n=== Dead Link Checker ===")
	infoColor.Printf("Starting scan for: %s\n", urlStr)
	infoColor.Printf("Max depth: %d, Delay: %dms, Parallel workers: %d\n\n", maxDepth, delayMs, parallelism)

	baseURL, err := parseURL(urlStr)
	if err != nil {
		errorColor.Printf("Error parsing URL: %s\n", err)
		return
	}

	domain := baseURL.Hostname()
	infoColor.Printf("Domain to scan: %s\n", domain)

	c := colly.NewCollector(
		colly.AllowedDomains(domain),
		colly.MaxDepth(maxDepth),
		colly.Async(true),
		colly.UserAgent(userAgent),
	)

	// rate limiter
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       time.Duration(delayMs) * time.Millisecond,
		RandomDelay: time.Duration(delayMs/2) * time.Millisecond,
		Parallelism: parallelism,
	})
	if err != nil {
		errorColor.Println("Failed to set rate limiter:", err)
		return
	}

	// Synchronize access to shared data
	var mu sync.Mutex
	visitedLinks := make(map[string]bool)
	deadLinks := make([]types.DeadLink, 0)
	visitedPages := 0
	currentPage := ""
	startTime := time.Now()

	c.SetRequestTimeout(time.Duration(timeoutSec) * time.Second)
	c.OnError(func(r *colly.Response, err error) {
		mu.Lock()
		defer mu.Unlock()

		if r.StatusCode == 403 || r.StatusCode == 429 || strings.Contains(err.Error(), "cloudflare") {
			warningColor.Printf("⚠️  SKIPPING %s (Blocked: %d - Likely Cloudflare protection)\n", r.Request.URL, r.StatusCode)
		} else {
			errorColor.Printf("⚠️  Error visiting %s: %s\n", r.Request.URL, err)
		}
	})

	// Save the current page
	c.OnRequest(func(r *colly.Request) {
		mu.Lock()
		currentPage = r.URL.String()
		visitedPages++
		infoColor.Printf("🔍 [%d] Visiting: %s\n", visitedPages, currentPage)
		mu.Unlock()
	})

	c.OnResponse(func(r *colly.Response) {
		successColor.Printf("✓ Page loaded: %s (Status: %d)\n", r.Request.URL, r.StatusCode)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" || strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if visitedLinks[link] {
			return
		}
		visitedLinks[link] = true
		checkLink(link, currentPage, "link", &deadLinks, infoColor, successColor, errorColor)
		if sameHost(link, urlStr) {
			e.Request.Visit(link)
		}
	})

	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		imgSrc := e.Request.AbsoluteURL(e.Attr("src"))
		if imgSrc == "" || strings.HasPrefix(imgSrc, "data:") {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if visitedLinks[imgSrc] {
			return
		}
		visitedLinks[imgSrc] = true
		checkLink(imgSrc, currentPage, "image", &deadLinks, infoColor, successColor, errorColor)
	})

	c.OnHTML("video source[src], video[src], iframe[src]", func(e *colly.HTMLElement) {
		videoSrc := e.Request.AbsoluteURL(e.Attr("src"))
		if videoSrc == "" {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if visitedLinks[videoSrc] {
			return
		}
		visitedLinks[videoSrc] = true

		mediaType := "video"
		if strings.Contains(e.Name, "iframe") {
			mediaType = "iframe"
		}
		checkLink(videoSrc, currentPage, mediaType, &deadLinks, infoColor, successColor, errorColor)
	})

	c.OnHTML("link[href], script[src]", func(e *colly.HTMLElement) {
		var resourceSrc string
		var resourceType string
		if e.Name == "link" {
			resourceSrc = e.Request.AbsoluteURL(e.Attr("href"))
			resourceType = "css"
		} else {
			resourceSrc = e.Request.AbsoluteURL(e.Attr("src"))
			resourceType = "script"
		}
		if resourceSrc == "" {
			return
		}
		mu.Lock()
		defer mu.Unlock()

		if visitedLinks[resourceSrc] {
			return
		}
		visitedLinks[resourceSrc] = true
		checkLink(resourceSrc, currentPage, resourceType, &deadLinks, infoColor, successColor, errorColor)
	})

	// Start crawling
	c.Visit(urlStr)
	c.Wait()

	totalTime := time.Since(startTime).Round(time.Second)
	printResults(deadLinks, visitedLinks, visitedPages, totalTime, titleColor, errorColor)
}
