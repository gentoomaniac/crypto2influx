package main

import (
	"fmt"
	"os"
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

	Coin []string `help:"coins to fetch data for and its values: <slug>:<buy_price>:<amount>" short:"c" required:""`

	ConfigFile *os.File `help:"path to config file" required:""`
	WriteOld   bool     `help:"write old line format data" default:"true" negatable:""`

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

	config, err := NewConfigFromFile(cli.ConfigFile)
	if err != nil {
		log.Error().Err(err).Str("file", cli.ConfigFile.Name()).Msg("failed reading config from file")
		ctx.Exit(1)
	}

	c, err := coinmarketcap.NewCoinmarketcap(config.Coinmarketcap.Token, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to set up coinmarketcap client")
		ctx.Exit(1)
	}

	coins, err := c.GetQuotesLatest(slugs)
	if err != nil {
		log.Error().Err(err).Msg("getting data failed")
		ctx.Exit(1)
	}

	client := influxdb2.NewClient(config.Influxcloud.BaseURL, config.Influxcloud.Token)
	defer client.Close()
	writeAPI := client.WriteAPI(config.Influxcloud.OrgName, config.Influxcloud.BucketName)

	for _, coin := range coins.Data {
		log.Info().Str("name", coin.Name).Float64("price", coin.Quote["USD"].Price).Time("lastUpdate", coin.Quote["USD"].LastUpdated).Msg("sending data for coin")
		coinLineFormat := fmt.Sprintf(
			"coin,symbol=%s,slug=%s,name=%s,is_active=%d,is_fiat=%d circulating_supply=%f,total_supply=%f,max_supply=%f,cmc_rank=%d %d",
			coin.Symbol,
			coin.Slug,
			strings.ReplaceAll(coin.Name, " ", "\\ "),
			coin.IsActive,
			coin.IsFiat,
			coin.CirculatingSupply,
			coin.TotalSupply,
			coin.MaxSupply,
			coin.CmcRank,
			coin.LastUpdated.UnixNano(),
		)
		writeAPI.WriteRecord(coinLineFormat)
		quoteLineFormat := fmt.Sprintf(
			"quote,symbol=%s price=%f,volume24h=%f,volumechange24h=%f,change1h=%f,change24h=%f,change7d=%f,change30d=%f,marketcap=%f,fullydillutedmarketcap=%f %d",
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
			coin.Quote["USD"].LastUpdated.UnixNano(),
		)
		writeAPI.WriteRecord(quoteLineFormat)

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

		for uid, investment := range config.Coins[coin.Symbol].Investments {
			investmentLineFormat := fmt.Sprintf(
				"investments,symbol=%s,platform=%s,uid=%s buy_price=%f,amount=%f,date=%d %d",
				coin.Symbol,
				investment.Platform,
				uid,
				investment.BuyPrice,
				investment.Amount,
				investment.Date.UnixMilli(),
				investment.Date.UnixNano(),
			)
			writeAPI.WriteRecord(investmentLineFormat)
		}
	}
	writeAPI.Flush()

	ctx.Exit(0)
}
