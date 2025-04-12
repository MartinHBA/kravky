package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Create an HTTP client with cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to create cookie jar: %v", err)
	}
	client := &http.Client{Jar: jar}

	// Login URL and form data
	loginURL := "https://www.cehz.sk/user/Login.action"
	loginData := "username=web&password=web&doLogIn=Vst%C3%BApte&user_agent=Mozilla/5.0+(Windows+NT+10.0;+Win64;+x64)+AppleWebKit/537.36+(KHTML,+like+Gecko)+Chrome/135.0.0.0+Safari/537.36&g-recaptcha-response=&_sourcePage=SYi4bqarXi_LV7MwLHZnfQZNmX-9l066&__fp=hX-zHhkBJSE%3D"

	// Create the login request
	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(loginData))
	if err != nil {
		log.Fatalf("Failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")

	// Send the login request
	log.Println("Sending login request...")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send login request: %v", err)
	}
	defer resp.Body.Close()

	// Check login response
	log.Printf("Login response status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Login failed with status code: %d", resp.StatusCode)
	}

	// Fetch data from the target URL
	targetURL := "https://www.cehz.sk/summs/CehzSummHD.action"
	log.Printf("Fetching data from URL: %s\n", targetURL)
	resp, err = client.Get(targetURL)
	if err != nil {
		log.Fatalf("Failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	// Read the raw response body
	log.Println("Reading raw response body...")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	log.Printf("Raw response body:\n%s\n", string(body))

	// Parse the HTML response
	log.Println("Parsing the HTML response...")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	// Extract key-value pairs
	log.Println("Extracting key-value pairs...")
	doc.Find("table.form_tab tr").Each(func(i int, s *goquery.Selection) {
		label := s.Find("label").Text()
		value := s.Find("td.text_CehzSumm_Count").Text()
		if label != "" && value != "" {
			log.Printf("Extracted: %s -> %s\n", label, value)
		}
	})

	log.Println("Program finished successfully.")
}
