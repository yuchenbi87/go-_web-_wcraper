package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Struct to hold scraped data
type PageData struct {
	URL   string   `json:"url"`
	Title string   `json:"title"`
	Text  string   `json:"text"`
	Tags  []string `json:"tags"`
}

func main() {
	start := time.Now()
	// Initialize a new collector
	c := colly.NewCollector(
		// Set options to limit scraping to Wikipedia
		colly.AllowedDomains("en.wikipedia.org"),
	)

	// Create a slice to store collected data
	var pages []PageData

	// On every request, collect text content without matching specific tags
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Save Wikipedia page HTML to wikipages directory
		page := strings.Split(e.Request.URL.Path, "/")[2]
		pageDir := "wikipages"
		filename := fmt.Sprintf("%s.html", page)
		if err := os.MkdirAll(pageDir, os.ModePerm); err != nil {
			log.Println("Could not create directory", err)
			return
		}
		filePath := filepath.Join(pageDir, filename)
		if err := os.WriteFile(filePath, e.Response.Body, 0644); err != nil {
			log.Println("Could not save file", err)
			return
		}
		fmt.Printf("Saved file %s\n", filename)

		// Extract text for the item for document corpus
		item := PageData{
			URL:   e.Request.URL.String(),
			Title: e.DOM.Find("h1").Text(),
			Text:  e.DOM.Find("#mw-content-text").Text(),
		}

		// Generate tags from URL segments
		tagsList := []string{
			strings.Split(e.Request.URL.Hostname(), ".")[0],
			strings.Split(e.Request.URL.Path, "/")[1],
		}
		moreTags := strings.Split(page, "_")
		for _, tag := range moreTags {
			tag = regexp.MustCompile(`[^a-zA-Z]`).ReplaceAllString(tag, "")
			if tag != "" {
				tagsList = append(tagsList, strings.ToLower(tag))
			}
		}
		item.Tags = tagsList

		// Add item to pages slice
		pages = append(pages, item)
	})

	// Handle errors in scraping
	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	// Set up a URL list for the crawl
	urls := []string{
		"https://en.wikipedia.org/wiki/Robotics",
		"https://en.wikipedia.org/wiki/Robot",
		"https://en.wikipedia.org/wiki/Reinforcement_learning",
		"https://en.wikipedia.org/wiki/Robot_Operating_System",
		"https://en.wikipedia.org/wiki/Intelligent_agent",
		"https://en.wikipedia.org/wiki/Software_agent",
		"https://en.wikipedia.org/wiki/Robotic_process_automation",
		"https://en.wikipedia.org/wiki/Chatbot",
		"https://en.wikipedia.org/wiki/Applications_of_artificial_intelligence",
		"https://en.wikipedia.org/wiki/Android_(robot)",
	}

	// Visit each URL
	for _, url := range urls {
		fmt.Println("Visiting:", url)
		c.Visit(url)
	}

	// Save the results to a JSON lines file
	file, err := os.Create("scraped_data.jl")
	if err != nil {
		log.Fatal("Could not create file", err)
	}
	defer file.Close()

	// Write JSON lines to the file
	for _, page := range pages {
		data, err := json.Marshal(page)
		if err != nil {
			log.Println("Error marshaling data: ", err)
			continue
		}
		file.Write(data)
		file.WriteString("\n")
	}

	fmt.Println("Scraping completed, data saved to scraped_data.jl")
	elapsed := time.Since(start)
	fmt.Printf("Running time is: %s", elapsed)
}
