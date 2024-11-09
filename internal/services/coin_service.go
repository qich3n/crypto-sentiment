package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CoinService struct {
	httpClient *http.Client
	cache      map[string]*CoinData
	cacheTime  map[string]time.Time
}

type CoinData struct {
	Symbol         string    `json:"symbol"`
	Name           string    `json:"name"`
	CurrentPrice   float64   `json:"current_price"`
	PriceChange24h float64   `json:"price_change_24h"`
	MarketCap      float64   `json:"market_cap"`
	LastUpdated    time.Time `json:"last_updated"`
}

func NewCoinService() *CoinService {
	return &CoinService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cache:      make(map[string]*CoinData),
		cacheTime:  make(map[string]time.Time),
	}
}

func (cs *CoinService) GetCoinData(symbol string) (*CoinData, error) {
	// Check cache (valid for 5 minutes)
	if data, ok := cs.cache[symbol]; ok {
		if time.Since(cs.cacheTime[symbol]) < 5*time.Minute {
			return data, nil
		}
	}

	// CoinGecko API URL
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd&include_24hr_change=true&include_market_cap=true",
		strings.ToLower(symbol))

	resp, err := cs.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]struct {
		Usd          float64 `json:"usd"`
		Usd24hChange float64 `json:"usd_24h_change"`
		UsdMarketCap float64 `json:"usd_market_cap"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Process the first result
	for _, data := range result {
		coinData := &CoinData{
			Symbol:         strings.ToUpper(symbol),
			CurrentPrice:   data.Usd,
			PriceChange24h: data.Usd24hChange,
			MarketCap:      data.UsdMarketCap,
			LastUpdated:    time.Now(),
		}

		// Update cache
		cs.cache[symbol] = coinData
		cs.cacheTime[symbol] = time.Now()

		return coinData, nil
	}

	return nil, fmt.Errorf("no data found for symbol: %s", symbol)
}
