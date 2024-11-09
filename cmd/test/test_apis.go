package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// TestAPIs is the exported test function
func TestAPIs() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Testing API Connections...")
	fmt.Println("\n1. Testing Reddit API:")
	testRedditAPI()

	fmt.Println("\n2. Testing X (Twitter) API:")
	testXAPI()
}

// Test Reddit API connection
func testRedditAPI() {
	clientID := os.Getenv("REDDIT_CLIENT_ID")
	clientSecret := os.Getenv("REDDIT_CLIENT_SECRET")

	authString := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(
		"POST",
		"https://www.reddit.com/api/v1/access_token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		log.Fatal("Reddit request creation failed:", err)
	}

	req.Header.Add("Authorization", "Basic "+authString)
	req.Header.Add("User-Agent", "CryptoSentimentBot/1.0")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Reddit API call failed:", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if accessToken, ok := result["access_token"]; ok {
		fmt.Println("✅ Reddit API Connection Successful!")
		fmt.Println("Access Token received:", accessToken)
	} else {
		fmt.Println("❌ Reddit API Connection Failed:", result)
	}
}

// Test X (Twitter) API connection
func testXAPI() {
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")

	req, err := http.NewRequest(
		"GET",
		"https://api.twitter.com/2/tweets/search/recent?query=bitcoin",
		nil,
	)
	if err != nil {
		log.Fatal("X API request creation failed:", err)
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("X API call failed:", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if _, ok := result["data"]; ok {
		fmt.Println("✅ X (Twitter) API Connection Successful!")
	} else {
		fmt.Println("❌ X (Twitter) API Connection Failed:", result)
	}
}
