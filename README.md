
# Giant

JSON API client for your Golang

## Why?

Why not just use the stdlib client?

 - set client timeouts
 - interpret non-200's statuses as error (tripper)
 - reuse marshal/unmarshal logics
 - set headers
 - log request/response (tripper)

## Useage

I often like to implement a service layer:

    type Client interface {
      SendObject(ctx context.Context, method, path string, snd, rcv any) (err error)
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

and then inject from above:

    client := cfg.Client.New()
    client.Use(&statusrt.StatusRt{})
    client.Use(&logrt.LogRt{Logger: lgr})

    weatherSvc := &svc.Svc{Client: client}
    hourly, err := weatherSvc.GetHourly(ctx, lat, lon)

## License

This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or
distribute this software, either in source code form or as a compiled
binary, for any purpose, commercial or non-commercial, and by any
means.

In jurisdictions that recognize copyright laws, the author or authors
of this software dedicate any and all copyright interest in the
software to the public domain. We make this dedication for the benefit
of the public at large and to the detriment of our heirs and
successors. We intend this dedication to be an overt act of
relinquishment in perpetuity of all present and future rights to this
software under copyright law.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

For more information, please refer to <http://unlicense.org/>
