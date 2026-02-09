// Package main demonstrates use of oauth2rt with giant
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/clarktrimble/giant"
	"github.com/clarktrimble/giant/logrt"
	"github.com/clarktrimble/giant/oauth2rt"
	"github.com/clarktrimble/giant/statusrt"

	"github.com/clarktrimble/giant/examples/oauth2/minlog"
)

func main() {

	cfg := &giant.Config{
		BaseUri:    requireEnv("OAUTH_BASE_URI"),
		SkipVerify: true,
	}

	lgr := &minlog.MinLog{}

	client := cfg.New()
	client.Use(&oauth2rt.OAuth2Rt{
		BaseUri:      cfg.BaseUri,
		TokenPath:    envOr("OAUTH_TOKEN_PATH", "/api/oauth"),
		ClientID:     requireEnv("OAUTH_CLIENT_ID"),
		ClientSecret: requireEnv("OAUTH_CLIENT_SECRET"),
		Logger:       lgr,
	})
	client.Use(&statusrt.StatusRt{})
	client.Use(logrt.New(lgr, []string{}, false))

	ctx := context.Background()
	path := envOr("OAUTH_TEST_PATH", "/api/session")

	data, err := client.SendJson(ctx, "GET", path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", data)
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Fprintf(os.Stderr, "missing required env: %s\n", key)
		os.Exit(1)
	}
	return val
}

func envOr(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
