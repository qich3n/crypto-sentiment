package models

import (
	"time"
)

type Coin struct {
	Symbol         string    `json:"symbol"`
	Name           string    `json:"name"`
	CurrentPrice   float64   `json:"current_price"`
	PriceChange24h float64   `json:"price_change_24h"`
	PriceChange7d  float64   `json:"price_change_7d"`
	MarketCap      float64   `json:"market_cap"`
	Volume24h      float64   `json:"volume_24h"`
	LastUpdated    time.Time `json:"last_updated"`
}

type CoinPrice struct {
	USD         float64   `json:"usd"`
	LastUpdated time.Time `json:"last_updated"`
}

type MarketData struct {
	CurrentPrice   CoinPrice `json:"current_price"`
	PriceChange24h float64   `json:"price_change_percentage_24h"`
	PriceChange7d  float64   `json:"price_change_percentage_7d"`
	MarketCap      CoinPrice `json:"market_cap"`
	Volume24h      CoinPrice `json:"total_volume"`
}
