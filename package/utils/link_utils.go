package utils

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MdSadiqMd/Scrape404/package/types"
	"github.com/fatih/color"
)

func CheckLink(link, currentPage, linkType string, deadLinks *[]types.DeadLink, infoColor, successColor, errorColor *color.Color) {
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

func ParseURL(rawURL string) (*url.URL, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}
	return url.Parse(rawURL)
}

func SameHost(link, baseURL string) bool {
	linkURL, err := ParseURL(link)
	if err != nil {
		return false
	}

	baseURLParsed, err := ParseURL(baseURL)
	if err != nil {
		return false
	}

	return linkURL.Hostname() == baseURLParsed.Hostname()
}
