package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func NewConfigFromFile(f *os.File) (*Config, error) {
	raw, _ := ioutil.ReadAll(bufio.NewReader(f))
	config := &Config{}
	err := json.Unmarshal(raw, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

type Config struct {
	Coinmarketcap CoinmarketcapCfg `json:"coinmarketcap"`
	Influxcloud   InfluxcloudCfg   `json:"influxcloud"`
	Coins         map[string]Coin  `json:"coins"`
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
	Slug        string                `json:"slug"`
	Investments map[string]Investment `json:"investments"`
}

type Investment struct {
	BuyPrice float64   `json:"buy_price"`
	Amount   float64   `json:"amount"`
	Date     time.Time `json:"date"`
	Platform string    `json:"platform"`
}

// First create a type alias
type Date time.Time

// Implement Marshaler and Unmarshaler interface
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(time.RFC3339, s)
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
