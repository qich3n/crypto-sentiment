package handlers

import (
	"crypto-sentiment/internal/services"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type SentimentHandler struct {
	redditService     *services.RedditService
	twitterService    *services.TwitterService
	sentimentAnalyzer *services.SentimentAnalyzer
	twitterEnabled    bool
}

func NewSentimentHandler() *SentimentHandler {
	// Initialize base services
	handler := &SentimentHandler{
		redditService:     services.NewRedditService(),
		sentimentAnalyzer: services.NewSentimentAnalyzer(),
	}

	// Check if Twitter credentials are provided
	if apiKey := strings.TrimSpace(os.Getenv("TWITTER_API_KEY")); apiKey != "" {
		var err error
		handler.twitterService, err = services.NewTwitterService()
		if err == nil {
			handler.twitterEnabled = true
		}
	}

	return handler
}

func (sh *SentimentHandler) GetSentiment(c *gin.Context) {
	symbol := c.Param("symbol")

	var (
		redditPosts []services.RedditPost
		tweets      []services.Tweet
		redditErr   error
		wg          sync.WaitGroup
	)

	// Always fetch Reddit data
	wg.Add(1)
	go func() {
		defer wg.Done()
		redditPosts, redditErr = sh.redditService.FetchPosts(symbol)
	}()

	// Only fetch Twitter data if enabled
	var twitterErr error
	if sh.twitterEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tweets, twitterErr = sh.twitterService.FetchTweets(symbol)
		}()
	}

	wg.Wait()

	// Handle Reddit errors
	if redditErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch Reddit data",
		})
		return
	}

	// Analyze Reddit sentiment
	var redditScore float64
	var redditResults []services.SentimentResult
	for _, post := range redditPosts {
		result := sh.sentimentAnalyzer.AnalyzeText(post.Title + " " + post.SelfText)
		redditResults = append(redditResults, result)
		redditScore += result.Score
	}
	if len(redditResults) > 0 {
		redditScore /= float64(len(redditResults))
	}

	// Initialize response
	response := gin.H{
		"symbol":       symbol,
		"reddit_score": redditScore,
		"reddit_posts": len(redditResults),
		"timestamp":    time.Now(),
	}

	// Add Twitter data if enabled and successfully fetched
	if sh.twitterEnabled {
		if twitterErr == nil {
			var twitterScore float64
			var twitterResults []services.SentimentResult
			for _, tweet := range tweets {
				result := sh.sentimentAnalyzer.AnalyzeText(tweet.Text)
				twitterResults = append(twitterResults, result)
				twitterScore += result.Score
			}
			if len(twitterResults) > 0 {
				twitterScore /= float64(len(twitterResults))
			}

			response["twitter_score"] = twitterScore
			response["tweets"] = len(twitterResults)

			// Calculate overall sentiment
			totalPosts := len(redditResults) + len(twitterResults)
			if totalPosts > 0 {
				overallScore := (redditScore*float64(len(redditResults)) +
					twitterScore*float64(len(twitterResults))) / float64(totalPosts)
				response["overall_score"] = overallScore
			}
		} else {
			response["twitter_error"] = "Failed to fetch Twitter data"
		}
	} else {
		// If Twitter is disabled, overall score is just Reddit score
		response["overall_score"] = redditScore
	}

	c.JSON(http.StatusOK, response)
}

// Add a method to check if Twitter is enabled
func (sh *SentimentHandler) IsTwitterEnabled() bool {
	return sh.twitterEnabled
}
