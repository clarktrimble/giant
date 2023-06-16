// Package main demonstrates use of a client service layer built with giant
package main

import (
	"context"
	"os"

	"github.com/clarktrimble/giant"
	"github.com/clarktrimble/giant/examples/weather/svc"
	"github.com/clarktrimble/giant/logrt"
	"github.com/clarktrimble/giant/statusrt"
)

var (
	lat     = 58.38
	lon     = 25.73
	baseUri = "https://api.open-meteo.com"
)

type Config struct {
	Client *giant.Config
}

func main() {

	cfg := &Config{
		Client: &giant.Config{
			BaseUri: baseUri,
		},
	}

	ctx := context.Background()
	lgr := &minLog{}

	client := cfg.Client.New()
	client.Use(&statusrt.StatusRt{})
	client.Use(&logrt.LogRt{Logger: lgr})

	weatherSvc := &svc.Svc{Client: client}
	hourly, err := weatherSvc.GetHourly(ctx, lat, lon)
	if err != nil {
		lgr.Error(ctx, "failed to get forcast data", err)
		os.Exit(1)
	}

	hourly.Print()
}
