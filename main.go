package main

import (
	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"

	"github.com/gentoomaniac/gocli"
	"github.com/gentoomaniac/logging"
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

	APIKey string   `help:"API key for coinmarketcap.com" short:"k" required:""`
	Coin   []string `help:"coins to fetch data for" short:"c" required:""`

	Version gocli.VersionFlag `short:"V" help:"Display version."`
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

	c, err := NewCoinmarketcap(cli.APIKey, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to set up coinmarketcap client")
		ctx.Exit(1)
	}

	coins, err := c.GetQuotesLatest(cli.Coin)
	if err != nil {
		log.Error().Err(err).Msg("getting data failed")
		ctx.Exit(1)
	}

	for _, coin := range coins.Data {
		log.Info().Str("name", coin.Name).Float64("price", coin.Quote["USD"].Price).Time("lastUpdate", coin.Quote["USD"].LastUpdated).Msg("")
	}

	ctx.Exit(0)
}
