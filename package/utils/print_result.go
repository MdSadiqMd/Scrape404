package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MdSadiqMd/Scrape404/package/types"
	"github.com/fatih/color"
)

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func PrintResults(deadLinks []types.DeadLink, visitedLinks map[string]bool, visitedPages int, duration time.Duration, titleColor, errorColor *color.Color) {
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
