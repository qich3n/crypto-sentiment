package services

import (
	"math"
	"strings"
	"sync"
	"time"
)

type SentimentAnalyzer struct {
	// Cryptocurrency-specific dictionaries
	positiveWords map[string]float64
	negativeWords map[string]float64
	mutex         sync.RWMutex
}

type SentimentResult struct {
	Score      float64   `json:"score"`
	Confidence float64   `json:"confidence"`
	Keywords   []string  `json:"keywords"`
	Timestamp  time.Time `json:"timestamp"`
}

func NewSentimentAnalyzer() *SentimentAnalyzer {
	sa := &SentimentAnalyzer{
		positiveWords: map[string]float64{
			"bullish":      1.5,
			"moon":         1.2,
			"buy":          1.0,
			"long":         1.0,
			"support":      0.8,
			"up":           0.7,
			"high":         0.7,
			"gains":        1.0,
			"profit":       1.0,
			"breakthrough": 1.2,
			"breakout":     1.2,
			"strong":       0.8,
			"upgrade":      1.0,
			"beat":         0.9,
			"growth":       0.9,
		},
		negativeWords: map[string]float64{
			"bearish":    -1.5,
			"dump":       -1.2,
			"sell":       -1.0,
			"short":      -1.0,
			"resistance": -0.8,
			"down":       -0.7,
			"low":        -0.7,
			"loss":       -1.0,
			"crash":      -1.5,
			"bear":       -1.2,
			"weak":       -0.8,
			"downgrade":  -1.0,
			"miss":       -0.9,
			"decline":    -0.9,
		},
	}
	return sa
}

func (sa *SentimentAnalyzer) AnalyzeText(text string) SentimentResult {
	sa.mutex.RLock()
	defer sa.mutex.RUnlock()

	text = strings.ToLower(text)
	words := strings.Fields(text)

	var score float64
	var matchCount int
	var keywords []string

	// Calculate sentiment score
	for _, word := range words {
		if value, ok := sa.positiveWords[word]; ok {
			score += value
			matchCount++
			keywords = append(keywords, word)
		}
		if value, ok := sa.negativeWords[word]; ok {
			score += value
			matchCount++
			keywords = append(keywords, word)
		}
	}

	// Calculate confidence based on number of matches
	confidence := math.Min(float64(matchCount)/5.0, 1.0) // Max confidence at 5 matches

	// Normalize score to [-1, 1] range
	if matchCount > 0 {
		score = score / float64(matchCount)
	}

	return SentimentResult{
		Score:      math.Max(math.Min(score, 1.0), -1.0),
		Confidence: confidence,
		Keywords:   keywords,
		Timestamp:  time.Now(),
	}
}

// UpdateDictionaries allows updating sentiment dictionaries dynamically
func (sa *SentimentAnalyzer) UpdateDictionaries(positive, negative map[string]float64) {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	if positive != nil {
		sa.positiveWords = positive
	}
	if negative != nil {
		sa.negativeWords = negative
	}
}
