# Giant: Add OAuth2 Support to NewWithTrippers

## Context

In `reauth-acp/main.go`, the OAuth2 client setup is verbose:

```go
client := cfg.Client.New()
client.Use(&oauth2rt.OAuth2Rt{
    BaseUri:      cfg.Oauth.Uri,
    TokenPath:    cfg.Oauth.Path,
    ClientID:     cfg.Oauth.ClientId,
    ClientSecret: string(cfg.Oauth.ClientSecret),
    Logger:       lgr,
})
client.Use(&statusrt.StatusRt{})
client.Use(logrt.New(lgr, []string{}, false))
```

This could be simplified to:

```go
client := cfg.Client.NewWithTrippers(lgr)
```

...if giant's `NewWithTrippers` supported OAuth2 config the same way it supports BasicAuth.

## Proposed Changes

### 1. Add OAuth2Config type to giant.go

```go
// OAuth2Config represents OAuth2 client credentials configuration.
type OAuth2Config struct {
    BaseUri      string `json:"base_uri" desc:"OAuth2 token endpoint base URI" required:"true"`
    TokenPath    string `json:"token_path" desc:"path to token endpoint" default:"/oauth/token"`
    ClientID     string `json:"client_id" desc:"OAuth2 client ID" required:"true"`
    ClientSecret Redact `json:"client_secret" desc:"OAuth2 client secret" required:"true"`
}
```

### 2. Add OAuth2 field to Config

```go
type Config struct {
    // ... existing fields ...

    // OAuth2 is for OAuth2 client credentials in NewWithTrippers.
    OAuth2 *OAuth2Config `json:"oauth2,omitempty" desc:"OAuth2 client credentials config"`
}
```

### 3. Extend NewWithTrippers

Add OAuth2Rt when configured (note: order matters - OAuth2 should wrap before StatusRt/LogRt):

```go
func (cfg *Config) NewWithTrippers(lgr logger) (giant *Giant) {

    giant = cfg.New()

    // OAuth2 goes first (innermost) so auth header is set before logging/status
    if cfg.OAuth2 != nil && cfg.OAuth2.ClientID != "" {
        giant.Use(&oauth2rt.OAuth2Rt{
            BaseUri:      cfg.OAuth2.BaseUri,
            TokenPath:    cfg.OAuth2.TokenPath,
            ClientID:     cfg.OAuth2.ClientID,
            ClientSecret: string(cfg.OAuth2.ClientSecret),
            Logger:       lgr,
        })
    }

    giant.Use(&statusrt.StatusRt{})
    giant.Use(logrt.New(lgr, cfg.RedactHeaders, cfg.SkipBody))

    if cfg.User != "" && cfg.Pass != "" {
        basicRt := basicrt.New(cfg.User, string(cfg.Pass))
        giant.Use(basicRt)
    }

    return
}
```

## Notes

- OAuth2 and BasicAuth are mutually exclusive in practice, but no need to enforce that
- The oauth2rt package already exists in giant, just needs config integration
- Consider adding "Authorization" to default RedactHeaders when OAuth2 is configured
