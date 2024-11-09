package handlers

import (
	"crypto-sentiment/internal/services"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type SentimentHandler struct {
	redditService     *services.RedditService
	twitterService    *services.TwitterService
	sentimentAnalyzer *services.SentimentAnalyzer
}

func NewSentimentHandler() *SentimentHandler {
	return &SentimentHandler{
		redditService:     services.NewRedditService(),
		sentimentAnalyzer: services.NewSentimentAnalyzer(),
	}
}

func (sh *SentimentHandler) GetSentiment(c *gin.Context) {
	symbol := c.Param("symbol")

	var (
		redditPosts           []services.RedditPost
		tweets                []services.Tweet
		redditErr, twitterErr error
		wg                    sync.WaitGroup
	)

	// Fetch data concurrently
	wg.Add(2)

	go func() {
		defer wg.Done()
		redditPosts, redditErr = sh.redditService.FetchPosts(symbol)
	}()

	go func() {
		defer wg.Done()
		tweets, twitterErr = sh.twitterService.FetchTweets(symbol)
	}()

	wg.Wait()

	if redditErr != nil && twitterErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch social media data",
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

	// Analyze Twitter sentiment
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

	// Calculate overall sentiment
	totalPosts := len(redditResults) + len(twitterResults)
	overallScore := 0.0
	if totalPosts > 0 {
		overallScore = (redditScore*float64(len(redditResults)) +
			twitterScore*float64(len(twitterResults))) / float64(totalPosts)
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":        symbol,
		"overall_score": overallScore,
		"reddit_score":  redditScore,
		"twitter_score": twitterScore,
		"reddit_posts":  len(redditResults),
		"tweets":        len(twitterResults),
		"timestamp":     time.Now(),
	})
}
