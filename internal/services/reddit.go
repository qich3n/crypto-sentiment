package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type RedditService struct {
	clientID     string
	clientSecret string
	accessToken  string
	httpClient   *http.Client
}

type RedditPost struct {
	Title     string    `json:"title"`
	SelfText  string    `json:"selftext"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_utc"`
	Subreddit string    `json:"subreddit"`
}

type RedditResponse struct {
	Data struct {
		Children []struct {
			Data RedditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func NewRedditService() *RedditService {
	return &RedditService{
		clientID:     os.Getenv("REDDIT_CLIENT_ID"),
		clientSecret: os.Getenv("REDDIT_CLIENT_SECRET"),
		httpClient:   &http.Client{},
	}
}

func (rs *RedditService) authenticate() error {
	authURL := "https://www.reddit.com/api/v1/access_token"

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", authURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(rs.clientID, rs.clientSecret)
	req.Header.Add("User-Agent", "CryptoSentimentBot/1.0")

	resp, err := rs.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	rs.accessToken = tokenResp.AccessToken
	return nil
}

func (rs *RedditService) FetchPosts(symbol string) ([]RedditPost, error) {
	if rs.accessToken == "" {
		if err := rs.authenticate(); err != nil {
			return nil, err
		}
	}

	subreddits := []string{"cryptocurrency", fmt.Sprintf("r/%s", symbol)}
	var allPosts []RedditPost

	for _, subreddit := range subreddits {
		url := fmt.Sprintf("https://oauth.reddit.com/r/%s/search.json?q=%s&sort=new&limit=100",
			subreddit, symbol)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("Authorization", "Bearer "+rs.accessToken)
		req.Header.Add("User-Agent", "CryptoSentimentBot/1.0")

		resp, err := rs.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var redditResp RedditResponse
		if err := json.NewDecoder(resp.Body).Decode(&redditResp); err != nil {
			return nil, err
		}

		for _, child := range redditResp.Data.Children {
			allPosts = append(allPosts, child.Data)
		}
	}

	return allPosts, nil
}
