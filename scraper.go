package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Function to retrieve HTML from website
func getHTML(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error saat mengambil HTML: %v", err)
	}
	defer resp.Body.Close()

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error membaca response body: %v", err)
	}

	return string(htmlData)
}

// Function to retrieve CSS from a <link> or <style> link in HTML
func getCSS(doc *goquery.Document, baseURL string) string {
	var cssContent string

	// Retrieve CSS content from the <link rel=“stylesheet”> element
	doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			absURL := resolveURL(baseURL, href)
			resp, err := http.Get(absURL)
			if err == nil {
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				cssContent += string(body) + "\n"
			}
		}
	})

	// Retrieve CSS content from the <style> element
	doc.Find("style").Each(func(i int, s *goquery.Selection) {
		cssContent += s.Text() + "\n"
	})

	return cssContent
}

// Function to retrieve JavaScript from <script> element
func getJS(doc *goquery.Document, baseURL string) string {
	var jsContent string

	// Fetch the JavaScript content of the <script> element
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			absURL := resolveURL(baseURL, src)
			resp, err := http.Get(absURL)
			if err == nil {
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				jsContent += string(body) + "\n"
			}
		} else {
			jsContent += s.Text() + "\n"
		}
	})

	return jsContent
}

// Function to get all links in the page
func getAllLinks(doc *goquery.Document, baseURL string) []string {
	var links []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			absURL := resolveURL(baseURL, href)
			if strings.HasPrefix(absURL, baseURL) {
				links = append(links, absURL)
			}
		}
	})
	return links
}

// Function to save content to file
func saveToFile(filename, content string) {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		log.Fatalf("Error saat menyimpan file %s: %v", filename, err)
	}
	fmt.Printf("Konten berhasil disimpan ke file %s\n", filename)
}

// Function to convert relative path to absolute
func resolveURL(baseURL, href string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Error parsing base URL: %v", err)
	}

	ref, err := url.Parse(href)
	if err != nil {
		log.Fatalf("Error parsing href: %v", err)
	}

	return base.ResolveReference(ref).String()
}

// Main function for scraping all pages
func scrapePage(pageURL, folderName string) {
	// Retrieve HTML from the page
	html := getHTML(pageURL)

	// Create a folder to save the file
	if err := os.MkdirAll(folderName, os.ModePerm); err != nil {
		log.Fatalf("Error membuat folder %s: %v", folderName, err)
	}

	// Save HTML to file
	saveToFile(filepath.Join(folderName, "index.html"), html)

	// Parsing HTML using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}

	// Fetch CSS and JavaScript from the page
	cssContent := getCSS(doc, pageURL)
	jsContent := getJS(doc, pageURL)

	// Save CSS and JavaScript to a file
	saveToFile(filepath.Join(folderName, "styles.css"), cssContent)
	saveToFile(filepath.Join(folderName, "scripts.js"), jsContent)

	// Fetch all links in the page
	links := getAllLinks(doc, pageURL)
	fmt.Printf("Ditemukan %d link pada halaman %s\n", len(links), pageURL)

	// Recursive for each page found
	for _, link := range links {
		// Use a unique folder name based on the URL
		linkFolder := folderName + "/" + strings.ReplaceAll(strings.TrimPrefix(link, pageURL), "/", "_")
		scrapePage(link, linkFolder)
	}
}

func main() {
	// URL of the website you want to retrieve content from
	url := ""

	// Start scraping the main page
	scrapePage(url, "output")
}
