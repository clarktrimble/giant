// Package svc demonstrates an api client service layer built with giant
package svc

import (
	"context"
	"fmt"
	"strings"
)

type Client interface {
	SendObject(ctx context.Context, method, path string, snd, rcv any) (err error)
}

type Hourly struct {
	Time     []CivilTime `json:"time"`
	Temp     []float64   `json:"temperature_2m"`
	Humidity []int       `json:"relativehumidity_2m"`
	Wind     []float64   `json:"windspeed_10m"`
}

type Svc struct {
	Client Client
}

func (svc *Svc) GetHourly(ctx context.Context, lat, lon float64) (hourly Hourly, err error) {

	var fc forecast
	err = svc.Client.SendObject(ctx, "GET", path(lat, lon), nil, &fc)
	if err != nil {
		return
	}

	hourly = fc.Hourly
	return
}

func (hourly *Hourly) Print() {

	for i, tm := range hourly.Time {
		fmt.Printf("%s   %.1f   %d   %.1f\n", tm.Fmt(), hourly.Temp[i], hourly.Humidity[i], hourly.Wind[i])
	}
}

// unexported

var (
	pathSpec   = "/v1/forecast?latitude=%.2f&longitude=%.2f&hourly=%s&forecast_days=%d"
	hourlyVars = []string{
		"temperature_2m",
		"relativehumidity_2m",
		"windspeed_10m",
	}
	days = 1
)

type forecast struct {
	Hourly Hourly `json:"hourly"`
}

func path(lat, lon float64) string {
	vars := strings.Join(hourlyVars, ",")
	return fmt.Sprintf(pathSpec, lat, lon, vars, days)
}
