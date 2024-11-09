package models

import "time"

type SentimentData struct {
	ID        int64     `json:"id"`
	Symbol    string    `json:"symbol"`
	Score     float64   `json:"score"`
	Reddit    float64   `json:"reddit_score"`
	Twitter   float64   `json:"twitter_score"`
	Posts     int       `json:"total_posts"`
	Timestamp time.Time `json:"timestamp"`
}

type SocialPost struct {
	Platform  string    `json:"platform"`
	Content   string    `json:"content"`
	Sentiment float64   `json:"sentiment_score"`
	CreatedAt time.Time `json:"created_at"`
}
