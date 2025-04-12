package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
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

// Response structure for API responses
type Response struct {
	Message string `json:"message"`
}

// DataEntry structure for extracted data
type DataEntry struct {
	Timestamp string `json:"timestamp"`
	Label     string `json:"label"`
	Value     string `json:"value"`
}

// Function to upload data to Azure Table Storage
func uploadToAzureTable(data []DataEntry) error {
	// Get Azure Storage account details from environment variables
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	accountKey := os.Getenv("AZURE_STORAGE_KEY")
	tableName := os.Getenv("AZURE_TABLE_NAME")

	if accountName == "" || accountKey == "" || tableName == "" {
		return fmt.Errorf("Azure Storage account details or table name are not set in environment variables")
	}

	// Create a service client
	cred, err := aztables.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return fmt.Errorf("failed to create Azure credential: %v", err)
	}

	serviceClient, err := aztables.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.table.core.windows.net", accountName), cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create Azure Table service client: %v", err)
	}

	// Get a client for the table
	tableClient := serviceClient.NewClient(tableName)

	// Insert data into the table
	for _, entry := range data {
		entity := map[string]interface{}{
			"PartitionKey": to.Ptr("DataPartition"),
			"RowKey":       to.Ptr(fmt.Sprintf("%s-%s", entry.Timestamp, entry.Label)),
			"Timestamp":    to.Ptr(entry.Timestamp),
			"Label":        to.Ptr(entry.Label),
			"Value":        to.Ptr(entry.Value),
		}

		_, err := tableClient.AddEntity(context.Background(), entity, nil)
		if err != nil {
			log.Printf("Failed to insert entity: %v", err)
		}
	}

	log.Println("Data successfully uploaded to Azure Table Storage.")
	return nil
}

// Updated fetchDataHandler to upload data to Azure Table Storage
func fetchDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Only POST is allowed.", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Received request to fetch data...")
	data, err := fetchData()
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Failed to fetch data"})
		return
	}

	// Upload data to Azure Table Storage
	err = uploadToAzureTable(data)
	if err != nil {
		log.Printf("Error uploading data to Azure Table Storage: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Failed to upload data to Azure Table Storage"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Message: "Data fetched and uploaded successfully"})
}

// Updated fetchData function to return JSON response
func fetchData() ([]DataEntry, error) {
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
	var data []DataEntry
	doc.Find("table.form_tab tr").Each(func(i int, s *goquery.Selection) {
		label := s.Find("label").Text()
		value := s.Find("td.text_CehzSumm_Count").Text()
		if label != "" && value != "" {
			data = append(data, DataEntry{
				Timestamp: timestamp,
				Label:     label,
				Value:     value,
			})
		}
	})

	log.Println("Data fetching completed successfully.")
	return data, nil
}
