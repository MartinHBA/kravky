package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

func main() {
	http.HandleFunc("/fetch-data", fetchDataHandler)

	// Start the HTTP server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Handler for the /fetch-data endpoint
func fetchDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Only POST is allowed.", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Received request to fetch data...")
	err := fetchData()
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Failed to fetch data"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Message: "Data fetched successfully"})
}

// Function to fetch data
func fetchData() error {
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
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
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

	// Convert response body to UTF-8
	log.Println("Converting response body to UTF-8...")
	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		log.Fatalf("Failed to create charset reader: %v", err)
	}
	utf8Body, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("Failed to read UTF-8 response body: %v", err)
	}

	// Parse the HTML response
	log.Println("Parsing the HTML response...")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(utf8Body)))
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	// Extract key-value pairs with timestamp
	log.Println("Extracting key-value pairs with timestamp...")
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	doc.Find("table.form_tab tr").Each(func(i int, s *goquery.Selection) {
		label := s.Find("label").Text()
		value := s.Find("td.text_CehzSumm_Count").Text()
		if label != "" && value != "" {
			log.Printf("%s | Extracted: %s -> %s\n", timestamp, label, value)
		}
	})

	log.Println("Data fetching completed successfully.")
	return nil
}

// Response structure for API responses
type Response struct {
	Message string `json:"message"`
}
