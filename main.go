package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"

	"github.com/gentoomaniac/crypto2influx/pkg/coinmarketcap"
	"github.com/gentoomaniac/gocli"
	"github.com/gentoomaniac/logging"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	version = "unknown"
	commit  = "unknown"
	binName = "unknown"
	builtBy = "unknown"
	date    = "unknown"
)

var cli struct {
	logging.LoggingConfig

	Coin               []string `help:"coins to fetch data for and its values: <slug>:<buy_price>:<amount>" short:"c" required:""`
	CoinmarketcapToken string   `help:"API key for coinmarketcap.com" required:""`
	InfluxToken        string   `help:"API token for influxcloud" required:""`
	InfluxURL          *url.URL `help:"url of influxdb" default:"https://eu-central-1-1.aws.cloud2.influxdata.com"`
	InfluxOrg          string   `help:"influxdb org name" required:""`
	InfluxBucket       string   `help:"influxdb bucket name" required:""`

	WriteOld bool `help:"write old line format data" default:"true" negatable:""`

	Version gocli.VersionFlag `short:"V" help:"Display version."`
}

func NewCoinData(parameter string) *CoinData {
	c := &CoinData{}

	fields := strings.Split(parameter, ":")
	if len(fields) != 3 {
		log.Error().Str("parameter", parameter).Msg("incorrect number of fields")
		return nil
	}

	var err error
	c.Slug = fields[0]
	c.BuyPrice, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		log.Error().Str("buyPrice", fields[1]).Msg("field is not a float")
		return nil
	}

	c.AmountOwned, err = strconv.ParseFloat(fields[2], 64)
	if err != nil {
		log.Error().Str("amountOwned", fields[2]).Msg("field is not a float")
		return nil
	}

	return c
}

type CoinData struct {
	Slug        string
	BuyPrice    float64
	AmountOwned float64
}

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError(), kong.Vars{
		"version": version,
		"commit":  commit,
		"binName": binName,
		"builtBy": builtBy,
		"date":    date,
	})
	logging.Setup(&cli.LoggingConfig)

	coinData := make(map[string]*CoinData)
	var slugs []string
	for _, coin := range cli.Coin {
		cd := NewCoinData(coin)
		coinData[cd.Slug] = cd
		slugs = append(slugs, cd.Slug)
	}

	c, err := coinmarketcap.NewCoinmarketcap(cli.CoinmarketcapToken, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to set up coinmarketcap client")
		ctx.Exit(1)
	}

	coins, err := c.GetQuotesLatest(slugs)
	if err != nil {
		log.Error().Err(err).Msg("getting data failed")
		ctx.Exit(1)
	}

	client := influxdb2.NewClient(cli.InfluxURL.String(), cli.InfluxToken)
	defer client.Close()
	writeAPI := client.WriteAPI(cli.InfluxOrg, cli.InfluxBucket)

	for _, coin := range coins.Data {
		log.Info().Str("name", coin.Name).Float64("price", coin.Quote["USD"].Price).Time("lastUpdate", coin.Quote["USD"].LastUpdated).Msg("sending data for coin")
		coin_data := fmt.Sprintf(
			"coin,symbol=%s,slug=%s,name=%s,is_active=%d,is_fiat=%d circulating_supply=%f,total_supply=%f,max_supply=%f,cmc_rank=%d",
			coin.Symbol,
			coin.Slug,
			strings.ReplaceAll(coin.Name, " ", "\\ "),
			coin.IsActive,
			coin.IsFiat,
			coin.CirculatingSupply,
			coin.TotalSupply,
			coin.MaxSupply,
			coin.CmcRank,
		)
		writeAPI.WriteRecord(coin_data)
		quote_data := fmt.Sprintf(
			"quote,symbol=%s price=%f,volume24h=%f,volumechange24h=%f,change1h=%f,change24h=%f,change7d=%f,change30d=%f,marketcap=%f,fullydillutedmarketcap=%f,last_updat=%d",
			coin.Symbol,
			coin.Quote["USD"].Price,
			coin.Quote["USD"].Volume24H,
			coin.Quote["USD"].VolumeChange24H,
			coin.Quote["USD"].PercentChange1H,
			coin.Quote["USD"].PercentChange24H,
			coin.Quote["USD"].PercentChange7D,
			coin.Quote["USD"].PercentChange30D,
			coin.Quote["USD"].MarketCap,
			coin.Quote["USD"].FullyDilutedMarketCap,
			coin.Quote["USD"].LastUpdated.UnixMilli(),
		)
		writeAPI.WriteRecord(quote_data)

		if cli.WriteOld {
			recordString := fmt.Sprintf(
				"%s,slug=%s,name=%s price=%f,buy=%f,amount=%f,volume24h=%f,volumechange24h=%f,change1h=%f,change24h=%f,change7d=%f,change30d=%f,marketcap=%f",
				coin.Symbol,
				coin.Slug,
				strings.ReplaceAll(coin.Name, " ", "\\ "),
				coin.Quote["USD"].Price,
				coinData[coin.Slug].BuyPrice,
				coinData[coin.Slug].AmountOwned,
				coin.Quote["USD"].Volume24H,
				coin.Quote["USD"].VolumeChange24H,
				coin.Quote["USD"].PercentChange1H,
				coin.Quote["USD"].PercentChange24H,
				coin.Quote["USD"].PercentChange7D,
				coin.Quote["USD"].PercentChange30D,
				coin.Quote["USD"].MarketCap,
			)
			writeAPI.WriteRecord(recordString)
		}
	}
	writeAPI.Flush()

	ctx.Exit(0)
}
