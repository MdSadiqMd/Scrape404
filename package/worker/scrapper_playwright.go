package worker

import (
	"sync"
	"time"

	"github.com/MdSadiqMd/Scrape404/package/types"
	"github.com/MdSadiqMd/Scrape404/package/utils"
	"github.com/fatih/color"
	"github.com/playwright-community/playwright-go"
)

func ScrapeWithPlaywright(urlStr string, maxDepth, delayMs, parallelism, timeoutSec int, userAgent string) {
	titleColor := color.New(color.FgCyan, color.Bold)
	successColor := color.New(color.FgGreen)
	errorColor := color.New(color.FgRed)
	infoColor := color.New(color.FgBlue)

	titleColor.Println("\n=== Dead Link Checker (Playwright Mode) ===")
	infoColor.Printf("Starting scan for: %s\n", urlStr)
	infoColor.Printf("Max depth: %d, Delay: %dms, Parallel workers: %d\n\n", maxDepth, delayMs, parallelism)

	baseURL, err := utils.ParseURL(urlStr)
	if err != nil {
		errorColor.Printf("Error parsing URL: %s\n", err)
		return
	}

	domain := baseURL.Hostname()
	infoColor.Printf("Domain to scan: %s\n", domain)

	err = playwright.Install()
	if err != nil {
		errorColor.Printf("Error installing Playwright: %s\n", err)
		return
	}

	pw, err := playwright.Run()
	if err != nil {
		errorColor.Printf("Error starting Playwright: %s\n", err)
		return
	}
	defer pw.Stop()

	browserOptions := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	}
	browser, err := pw.Chromium.Launch(browserOptions)
	if err != nil {
		errorColor.Printf("Error launching browser: %s\n", err)
		return
	}
	defer browser.Close()

	var mu sync.Mutex
	visitedLinks := make(map[string]bool)
	deadLinks := make([]types.DeadLink, 0)
	visitedPages := 0
	startTime := time.Now()

	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup

	var scrapeURLFn func(string, int)
	scrapeURLFn = func(url string, depth int) {
		defer wg.Done()
		defer func() { <-sem }()

		if depth > maxDepth {
			return
		}

		mu.Lock()
		visitedPages++
		pageNum := visitedPages
		mu.Unlock()

		infoColor.Printf("🔍 [%d] Visiting with Playwright: %s\n", pageNum, url)
		context, err := browser.NewContext(playwright.BrowserNewContextOptions{
			UserAgent: playwright.String(userAgent),
		})
		if err != nil {
			errorColor.Printf("Error creating browser context: %s\n", err)
			return
		}
		defer context.Close()

		page, err := context.NewPage()
		if err != nil {
			errorColor.Printf("Error creating page: %s\n", err)
			return
		}
		defer page.Close()

		page.SetDefaultTimeout(float64(timeoutSec * 1000))
		resp, err := page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		})

		if err != nil {
			mu.Lock()
			errorColor.Printf("⚠️  Error visiting %s: %s\n", url, err)
			mu.Unlock()
			return
		}

		status := resp.Status()
		if status >= 200 && status < 400 {
			mu.Lock()
			successColor.Printf("✓ Page loaded: %s (Status: %d)\n", url, status)
			mu.Unlock()
		} else {
			mu.Lock()
			errorColor.Printf("⚠️ Failed to load %s (Status: %d)\n", url, status)
			mu.Unlock()
			return
		}

		links, err := page.Evaluate(`() => {
			const results = {
				links: [],
				images: [],
				videos: [],
				iframes: [],
				stylesheets: [],
				scripts: []
			};
			
			document.querySelectorAll('a[href]').forEach(a => {
				if (a.href && !a.href.startsWith('javascript:') && !a.href.startsWith('mailto:')) {
					results.links.push(a.href);
				}
			});
			
			document.querySelectorAll('img[src]').forEach(img => {
				if (img.src && !img.src.startsWith('data:')) {
					results.images.push(img.src);
				}
			});
			
			document.querySelectorAll('video[src], video source[src], iframe[src]').forEach(el => {
				if (el.src) {
					if (el.tagName.toLowerCase() === 'iframe') {
						results.iframes.push(el.src);
					} else {
						results.videos.push(el.src);
					}
				}
			});
			
			document.querySelectorAll('link[href][rel="stylesheet"]').forEach(link => {
				if (link.href) {
					results.stylesheets.push(link.href);
				}
			});
			
			document.querySelectorAll('script[src]').forEach(script => {
				if (script.src) {
					results.scripts.push(script.src);
				}
			});
			
			return results;
		}`)

		if err != nil {
			mu.Lock()
			errorColor.Printf("Error extracting links from %s: %s\n", url, err)
			mu.Unlock()
			return
		}
		resourcesObj := links
		resourcesMap := resourcesObj.(map[string]interface{})
		mu.Lock()
		currentPage := url
		if resourcesMap["links"] != nil {
			for _, link := range resourcesMap["links"].([]interface{}) {
				linkStr := link.(string)
				if !visitedLinks[linkStr] {
					visitedLinks[linkStr] = true
					utils.CheckLink(linkStr, currentPage, "link", &deadLinks, infoColor, successColor, errorColor)

					if utils.SameHost(linkStr, urlStr) && depth+1 <= maxDepth {
						wg.Add(1)
						go func(l string, d int) {
							time.Sleep(time.Duration(delayMs) * time.Millisecond)
							sem <- struct{}{}
							scrapeURLFn(l, d)
						}(linkStr, depth+1)
					}
				}
			}
		}

		if resourcesMap["images"] != nil {
			for _, img := range resourcesMap["images"].([]interface{}) {
				imgStr := img.(string)
				if !visitedLinks[imgStr] {
					visitedLinks[imgStr] = true
					utils.CheckLink(imgStr, currentPage, "image", &deadLinks, infoColor, successColor, errorColor)
				}
			}
		}

		if resourcesMap["videos"] != nil {
			for _, video := range resourcesMap["videos"].([]interface{}) {
				videoStr := video.(string)
				if !visitedLinks[videoStr] {
					visitedLinks[videoStr] = true
					utils.CheckLink(videoStr, currentPage, "video", &deadLinks, infoColor, successColor, errorColor)
				}
			}
		}

		if resourcesMap["iframes"] != nil {
			for _, iframe := range resourcesMap["iframes"].([]interface{}) {
				iframeStr := iframe.(string)
				if !visitedLinks[iframeStr] {
					visitedLinks[iframeStr] = true
					utils.CheckLink(iframeStr, currentPage, "iframe", &deadLinks, infoColor, successColor, errorColor)
				}
			}
		}

		if resourcesMap["stylesheets"] != nil {
			for _, css := range resourcesMap["stylesheets"].([]interface{}) {
				cssStr := css.(string)
				if !visitedLinks[cssStr] {
					visitedLinks[cssStr] = true
					utils.CheckLink(cssStr, currentPage, "css", &deadLinks, infoColor, successColor, errorColor)
				}
			}
		}

		if resourcesMap["scripts"] != nil {
			for _, script := range resourcesMap["scripts"].([]interface{}) {
				scriptStr := script.(string)
				if !visitedLinks[scriptStr] {
					visitedLinks[scriptStr] = true
					utils.CheckLink(scriptStr, currentPage, "script", &deadLinks, infoColor, successColor, errorColor)
				}
			}
		}
		mu.Unlock()
	}
	wg.Add(1)
	sem <- struct{}{}
	go scrapeURLFn(urlStr, 0)
	wg.Wait()

	totalTime := time.Since(startTime).Round(time.Second)
	utils.PrintResults(deadLinks, visitedLinks, visitedPages, totalTime, titleColor, errorColor)
}
