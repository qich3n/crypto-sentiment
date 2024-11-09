package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Tweet struct {
	Text string `json:"text"`
}

type RedditPost struct {
	Title    string `json:"title"`
	SelfText string `json:"selftext"`
}

// Global variable to store X Bearer Token
var xBearerToken string

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize X Bearer Token
	var err error
	xBearerToken, err = getXBearerToken()
	if err != nil {
		log.Printf("Warning: Failed to initialize X Bearer Token: %v", err)
	}

	// Create necessary directories if they don't exist
	createRequiredDirectories()
}

func createRequiredDirectories() {
	dirs := []string{
		"frontend/templates",
		"frontend/static/js",
		"frontend/static/css",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("Warning: Failed to create directory %s: %v", dir, err)
		}
	}

	// Create index.html if it doesn't exist
	indexPath := filepath.Join("frontend", "templates", "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		file, err := os.Create(indexPath)
		if err != nil {
			log.Printf("Warning: Failed to create index.html: %v", err)
			return
		}
		defer file.Close()

		// Write basic HTML template
		_, err = file.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Crypto Sentiment Analyzer</title>
    <script src="https://unpkg.com/react@17/umd/react.development.js"></script>
    <script src="https://unpkg.com/react-dom@17/umd/react-dom.development.js"></script>
    <script src="https://unpkg.com/babel-standalone@6/babel.min.js"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div id="root"></div>
    <script type="text/babel" src="/static/js/main.js"></script>
</body>
</html>`)
		if err != nil {
			log.Printf("Warning: Failed to write to index.html: %v", err)
		}
	}
}

func main() {
	// Check if we're running in test mode
	if len(os.Args) > 1 && os.Args[1] == "test-apis" {
		testAPIs()
		return
	}

	// Normal server startup
	r := gin.Default()

	// Basic CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Serve static files
	r.Static("/static", "./frontend/static")

	// Load HTML templates
	r.LoadHTMLGlob("frontend/templates/*")

	// Serve index.html for root route
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/health", healthCheck)
		api.GET("/sentiment/:symbol", getSentiment)
		api.GET("/trending", getTrending)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}

func getXBearerToken() (string, error) {
	apiKey := os.Getenv("TWITTER_API_KEY")
	apiSecret := os.Getenv("TWITTER_API_SECRET")

	// Encode credentials
	credentials := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", apiKey, apiSecret)))

	// Create request for bearer token
	req, err := http.NewRequest(
		"POST",
		"https://api.twitter.com/oauth2/token",
		strings.NewReader("grant_type=client_credentials"),
	)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Basic "+credentials)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func testAPIs() {
	fmt.Println("Testing API Connections...")

	// Test Reddit API
	fmt.Println("\n1. Testing Reddit API:")
	testRedditAPI()

	// Test X (Twitter) API
	fmt.Println("\n2. Testing X (Twitter) API:")
	testXAPI()
}

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

func testXAPI() {
	if xBearerToken == "" {
		var err error
		xBearerToken, err = getXBearerToken()
		if err != nil {
			log.Fatal("Failed to get X Bearer Token:", err)
		}
	}

	req, err := http.NewRequest(
		"GET",
		"https://api.twitter.com/2/tweets/search/recent?query=bitcoin",
		nil,
	)
	if err != nil {
		log.Fatal("X API request creation failed:", err)
	}

	req.Header.Add("Authorization", "Bearer "+xBearerToken)

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
		fmt.Println("Bearer Token is valid and working!")
	} else {
		fmt.Println("❌ X (Twitter) API Connection Failed:", result)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
	})
}

func getSentiment(c *gin.Context) {
	symbol := c.Param("symbol")

	// Get Twitter sentiment
	tweets, err := getTwitterPosts(symbol)
	if err != nil {
		log.Printf("Twitter error: %v", err)
	}

	// Get Reddit sentiment
	redditPosts, err := getRedditPosts(symbol)
	if err != nil {
		log.Printf("Reddit error: %v", err)
	}

	// Calculate Twitter sentiment
	var twitterScore float64
	var tweetCount int
	for _, tweet := range tweets {
		score := calculateSentiment(tweet.Text)
		twitterScore += score
		tweetCount++
	}
	if tweetCount > 0 {
		twitterScore = twitterScore / float64(tweetCount)
	}

	// Calculate Reddit sentiment
	var redditScore float64
	var redditCount int
	for _, post := range redditPosts {
		score := calculateSentiment(post.Title + " " + post.SelfText)
		redditScore += score
		redditCount++
	}
	if redditCount > 0 {
		redditScore = redditScore / float64(redditCount)
	}

	// Calculate overall sentiment
	totalPosts := tweetCount + redditCount
	var overallScore float64
	if totalPosts > 0 {
		overallScore = (twitterScore*float64(tweetCount) + redditScore*float64(redditCount)) / float64(totalPosts)
	}

	c.JSON(200, gin.H{
		"symbol":        symbol,
		"overall_score": overallScore,
		"reddit_score":  redditScore,
		"twitter_score": twitterScore,
		"reddit_posts":  redditCount,
		"tweets":        tweetCount,
		"timestamp":     time.Now(),
	})
}

func getTwitterPosts(symbol string) ([]Tweet, error) {
	if xBearerToken == "" {
		var err error
		xBearerToken, err = getXBearerToken()
		if err != nil {
			return nil, err
		}
	}

	query := url.QueryEscape(fmt.Sprintf("%s crypto -is:retweet lang:en", symbol))
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.twitter.com/2/tweets/search/recent?query=%s&max_results=100", query),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+xBearerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []Tweet `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func getRedditPosts(symbol string) ([]RedditPost, error) {
	clientID := os.Getenv("REDDIT_CLIENT_ID")
	clientSecret := os.Getenv("REDDIT_CLIENT_SECRET")

	// Get Reddit access token
	authString := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	tokenReq, err := http.NewRequest(
		"POST",
		"https://www.reddit.com/api/v1/access_token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}

	tokenReq.Header.Add("Authorization", "Basic "+authString)
	tokenReq.Header.Add("User-Agent", "CryptoSentimentBot/1.0")
	tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return nil, err
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		return nil, err
	}

	// Use token to get posts
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://oauth.reddit.com/r/cryptocurrency/search.json?q=%s&limit=100&sort=new",
			url.QueryEscape(symbol)),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+tokenResult.AccessToken)
	req.Header.Add("User-Agent", "CryptoSentimentBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Children []struct {
				Data RedditPost `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var posts []RedditPost
	for _, child := range result.Data.Children {
		posts = append(posts, child.Data)
	}

	return posts, nil
}

func calculateSentiment(text string) float64 {
	text = strings.ToLower(text)

	positiveWords := []string{
		"bullish", "moon", "buy", "green", "up", "good", "great",
		"positive", "profit", "gains", "win", "winning", "success",
	}

	negativeWords := []string{
		"bearish", "dump", "sell", "red", "down", "bad", "crash",
		"negative", "loss", "bear", "fail", "failing", "scam",
	}

	var positiveCount, negativeCount float64

	words := strings.Fields(text)
	for _, word := range words {
		for _, pos := range positiveWords {
			if strings.Contains(word, pos) {
				positiveCount++
				break
			}
		}
		for _, neg := range negativeWords {
			if strings.Contains(word, neg) {
				negativeCount++
				break
			}
		}
	}

	if positiveCount == 0 && negativeCount == 0 {
		return 0
	}

	total := positiveCount + negativeCount
	return (positiveCount - negativeCount) / total
}

func getTrending(c *gin.Context) {
	symbols := []string{"BTC", "ETH", "BNB", "XRP", "DOGE"}
	var trending []gin.H

	for _, symbol := range symbols {
		tweets, _ := getTwitterPosts(symbol)
		redditPosts, _ := getRedditPosts(symbol)

		var totalScore float64
		var count int

		// Calculate sentiment from tweets
		for _, tweet := range tweets {
			score := calculateSentiment(tweet.Text)
			totalScore += score
			count++
		}

		// Calculate sentiment from Reddit posts
		for _, post := range redditPosts {
			score := calculateSentiment(post.Title + " " + post.SelfText)
			totalScore += score
			count++
		}

		if count > 0 {
			trending = append(trending, gin.H{
				"symbol": symbol,
				"score":  totalScore / float64(count),
				"posts":  count,
			})
		}
	}

	c.JSON(200, gin.H{
		"trending":  trending,
		"timestamp": time.Now(),
	})
}
