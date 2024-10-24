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

// Fungsi untuk mengambil HTML dari website
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

// Fungsi untuk mengambil CSS dari link <link> atau <style> dalam HTML
func getCSS(doc *goquery.Document, baseURL string) string {
	var cssContent string

	// Mengambil konten CSS dari elemen <link rel="stylesheet">
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

	// Mengambil konten CSS dari elemen <style>
	doc.Find("style").Each(func(i int, s *goquery.Selection) {
		cssContent += s.Text() + "\n"
	})

	return cssContent
}

// Fungsi untuk mengambil JavaScript dari elemen <script>
func getJS(doc *goquery.Document, baseURL string) string {
	var jsContent string

	// Mengambil konten JavaScript dari elemen <script>
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

// Fungsi untuk mendapatkan semua link di halaman
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

// Fungsi untuk menyimpan konten ke file
func saveToFile(filename, content string) {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		log.Fatalf("Error saat menyimpan file %s: %v", filename, err)
	}
	fmt.Printf("Konten berhasil disimpan ke file %s\n", filename)
}

// Fungsi untuk mengubah path relatif menjadi absolut
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

// Fungsi utama untuk scraping semua halaman
func scrapePage(pageURL, folderName string) {
	// Ambil HTML dari halaman
	html := getHTML(pageURL)

	// Buat folder untuk menyimpan file
	if err := os.MkdirAll(folderName, os.ModePerm); err != nil {
		log.Fatalf("Error membuat folder %s: %v", folderName, err)
	}

	// Simpan HTML ke file
	saveToFile(filepath.Join(folderName, "index.html"), html)

	// Parsing HTML menggunakan goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}

	// Ambil CSS dan JavaScript dari halaman
	cssContent := getCSS(doc, pageURL)
	jsContent := getJS(doc, pageURL)

	// Simpan CSS dan JavaScript ke file
	saveToFile(filepath.Join(folderName, "styles.css"), cssContent)
	saveToFile(filepath.Join(folderName, "scripts.js"), jsContent)

	// Ambil semua link di halaman
	links := getAllLinks(doc, pageURL)
	fmt.Printf("Ditemukan %d link pada halaman %s\n", len(links), pageURL)

	// Rekursif untuk setiap halaman yang ditemukan
	for _, link := range links {
		// Gunakan nama folder yang unik berdasarkan URL
		linkFolder := folderName + "/" + strings.ReplaceAll(strings.TrimPrefix(link, pageURL), "/", "_")
		scrapePage(link, linkFolder)
	}
}

func main() {
	// URL website yang ingin diambil kontennya
	url := "https://templatekit.jegtheme.com/riderhood"

	// Mulai scraping halaman utama
	scrapePage(url, "output")
}
