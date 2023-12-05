// Package main demonstrates use of a client service layer built with giant
package main

import (
	"context"
	"os"

	"github.com/clarktrimble/giant"

	"github.com/clarktrimble/giant/examples/weather/minlog"
	"github.com/clarktrimble/giant/examples/weather/svc"
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

	lgr := &minlog.MinLog{}
	client := cfg.Client.NewWithTrippers(lgr)

	ctx := context.Background()
	weatherSvc := &svc.Svc{Client: client}

	hourly, err := weatherSvc.GetHourly(ctx, lat, lon)
	if err != nil {
		lgr.Error(ctx, "failed to get forcast data", err)
		os.Exit(1)
	}

	hourly.Print()
}
