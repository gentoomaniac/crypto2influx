package main

import (
	"encoding/json"
	"strings"
	"time"
)

type Config struct {
	Coinmarketcap CoinmarketcapCfg `json:"coinmarketcap"`
	Influxcloud   InfluxcloudCfg   `json:"influxcloud"`
	Coins         []Coin           `json:"coins"`
}

type CoinmarketcapCfg struct {
	Token string `json:"token"`
}

type InfluxcloudCfg struct {
	Token      string `json:"token"`
	OrgName    string `json:"org"`
	BucketName string `json:"bucket"`
}

type Coin struct {
	Slug        string       `json:"slug"`
	Investments []Investment `json:"investments"`
}

type Investment struct {
	BuyPrice float64 `json:"buy_price"`
	Amount   float64 `json:"amount"`
	Date     Date    `json:"date"`
	Platform string  `json:"platform"`
}

// First create a type alias
type Date time.Time

// Implement Marshaler and Unmarshaler interface
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d))
}
func (d Date) Format(s string) string {
	t := time.Time(d)
	return t.Format(s)
}
