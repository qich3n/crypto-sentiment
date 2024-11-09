package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Tweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type TwitterResponse struct {
	Data []struct {
		ID        string    `json:"id"`
		Text      string    `json:"text"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"data"`
	Meta struct {
		ResultCount  int    `json:"result_count"`
		NextToken    string `json:"next_token"`
		RefreshedURL string `json:"refreshed_url"`
	} `json:"meta"`
}

type TwitterService struct {
	apiKey      string
	apiSecret   string
	bearerToken string
	httpClient  *http.Client
}

func NewTwitterService() (*TwitterService, error) {
	ts := &TwitterService{
		apiKey:     os.Getenv("TWITTER_API_KEY"),
		apiSecret:  os.Getenv("TWITTER_API_SECRET"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Get bearer token during initialization
	token, err := ts.getBearerToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get bearer token: %v", err)
	}

	ts.bearerToken = token
	return ts, nil
}

func (ts *TwitterService) getBearerToken() (string, error) {
	credentials := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", ts.apiKey, ts.apiSecret)))

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(
		"POST",
		"https://api.twitter.com/oauth2/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return "", fmt.Errorf("twitter API error: %v", errorResponse)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

// FetchTweets retrieves tweets for a given cryptocurrency symbol
func (ts *TwitterService) FetchTweets(symbol string) ([]Tweet, error) {
	// Create query parameters
	query := url.QueryEscape(fmt.Sprintf("%s crypto -is:retweet lang:en", symbol))
	requestURL := fmt.Sprintf(
		"https://api.twitter.com/2/tweets/search/recent?query=%s&max_results=100&tweet.fields=created_at",
		query,
	)

	// Create request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.bearerToken))

	// Make request
	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return nil, fmt.Errorf("twitter API error: %v", errorResponse)
	}

	// Parse response
	var twitterResp TwitterResponse
	if err := json.NewDecoder(resp.Body).Decode(&twitterResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Convert response to Tweet slice
	tweets := make([]Tweet, len(twitterResp.Data))
	for i, data := range twitterResp.Data {
		tweets[i] = Tweet{
			ID:        data.ID,
			Text:      data.Text,
			CreatedAt: data.CreatedAt,
		}
	}

	return tweets, nil
}

func (ts *TwitterService) TestConnection() error {
	// Test the connection with a simple search request
	req, err := http.NewRequest(
		"GET",
		"https://api.twitter.com/2/tweets/search/recent?query=bitcoin",
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ts.bearerToken))

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return fmt.Errorf("API test failed: %v", errorResponse)
	}

	return nil
}
