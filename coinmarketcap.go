package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type URLValues map[string]string
type HTTPHeaders map[string]string

var (
	APIDefaultURL, _ = url.Parse("https://pro-api.coinmarketcap.com/v1")
)

func NewCoinmarketcap(key string, baseURL *url.URL) (*Coinmarketcap, error) {
	if key == "" {
		return nil, fmt.Errorf("api key cannot be empty")
	}

	c := &Coinmarketcap{
		Key:     key,
		BaseURL: APIDefaultURL,
	}

	if baseURL != nil {
		c.BaseURL = baseURL
	}

	return c, nil
}

type Coinmarketcap struct {
	Key     string
	BaseURL *url.URL
}

func (c Coinmarketcap) sendRequest(endpoint string, urlValues URLValues, headers HTTPHeaders) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", c.BaseURL.String()+"/"+endpoint, nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	for k, v := range urlValues {
		q.Add(k, v)
	}

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", c.Key)
	for k, v := range headers {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

type Quote struct {
	Price                 float64   `json:"price"`
	Volume24H             float64   `json:"volume_24h"`
	VolumeChange24H       float64   `json:"volume_change_24h"`
	PercentChange1H       float64   `json:"percent_change_1h"`
	PercentChange24H      float64   `json:"percent_change_24h"`
	PercentChange7D       float64   `json:"percent_change_7d"`
	PercentChange30D      float64   `json:"percent_change_30d"`
	MarketCap             float64   `json:"market_cap"`
	MarketCapDominance    float64   `json:"market_cap_dominance"`
	FullyDilutedMarketCap float64   `json:"fully_diluted_market_cap"`
	LastUpdated           time.Time `json:"last_updated"`
}

type Coin struct {
	ID                int              `json:"id"`
	Name              string           `json:"name"`
	Symbol            string           `json:"symbol"`
	Slug              string           `json:"slug"`
	IsActive          int              `json:"is_active"`
	IsFiat            int              `json:"is_fiat"`
	CirculatingSupply float64          `json:"circulating_supply"`
	TotalSupply       float64          `json:"total_supply"`
	MaxSupply         float64          `json:"max_supply"`
	DateAdded         time.Time        `json:"date_added"`
	NumMarketPairs    int              `json:"num_market_pairs"`
	CmcRank           int              `json:"cmc_rank"`
	LastUpdated       time.Time        `json:"last_updated"`
	Tags              []string         `json:"tags"`
	Platform          interface{}      `json:"platform"`
	Quote             map[string]Quote `json:"quote"`
}

// https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyQuotesLatest
type QuotesLatest struct {
	Data   map[int]Coin `json:"data"`
	Status struct {
		Timestamp    time.Time `json:"timestamp"`
		ErrorCode    int       `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Elapsed      int       `json:"elapsed"`
		CreditCount  int       `json:"credit_count"`
	} `json:"status"`
}

func (c Coinmarketcap) GetQuotesLatest(slugs []string) (*QuotesLatest, error) {
	urlValues := make(URLValues)
	urlValues["slug"] = strings.Join(slugs, ",")
	data, err := c.sendRequest("cryptocurrency/quotes/latest", urlValues, nil)
	if err != nil {
		return nil, err
	}

	var quotes QuotesLatest
	err = json.Unmarshal([]byte(data), &quotes)
	if err != nil {
		return nil, err
	}

	return &quotes, nil
}

// https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyListingsLatest
type ListingsLatest struct {
	Status struct {
		Timestamp    time.Time   `json:"timestamp"`
		ErrorCode    int         `json:"error_code"`
		ErrorMessage interface{} `json:"error_message"`
		Elapsed      int         `json:"elapsed"`
		CreditCount  int         `json:"credit_count"`
		Notice       interface{} `json:"notice"`
		TotalCount   int         `json:"total_count"`
	} `json:"status"`
	Data []struct {
		ID                int         `json:"id"`
		Name              string      `json:"name"`
		Symbol            string      `json:"symbol"`
		Slug              string      `json:"slug"`
		NumMarketPairs    int         `json:"num_market_pairs"`
		DateAdded         time.Time   `json:"date_added"`
		Tags              []string    `json:"tags"`
		MaxSupply         float64     `json:"max_supply"`
		CirculatingSupply float64     `json:"circulating_supply"`
		TotalSupply       float64     `json:"total_supply"`
		Platform          interface{} `json:"platform"`
		CmcRank           int         `json:"cmc_rank"`
		LastUpdated       time.Time   `json:"last_updated"`
		Quote             struct {
			Usd struct {
				Price                 float64   `json:"price"`
				Volume24H             float64   `json:"volume_24h"`
				VolumeChange24H       float64   `json:"volume_change_24h"`
				PercentChange1H       float64   `json:"percent_change_1h"`
				PercentChange24H      float64   `json:"percent_change_24h"`
				PercentChange7D       float64   `json:"percent_change_7d"`
				PercentChange30D      float64   `json:"percent_change_30d"`
				PercentChange60D      float64   `json:"percent_change_60d"`
				PercentChange90D      float64   `json:"percent_change_90d"`
				MarketCap             float64   `json:"market_cap"`
				MarketCapDominance    float64   `json:"market_cap_dominance"`
				FullyDilutedMarketCap float64   `json:"fully_diluted_market_cap"`
				LastUpdated           time.Time `json:"last_updated"`
			} `json:"USD"`
		} `json:"quote"`
	} `json:"data"`
}
