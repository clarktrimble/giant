package svc

import (
	"fmt"
	"strings"
	"time"
)

//
// Inspired by: https://romangaranin.net/posts/2021-02-19-json-time-and-golang
//

type CivilTime time.Time

func (ct *CivilTime) Fmt() string {

	return time.Time(*ct).Format("01-02T15")
}

func (ct *CivilTime) UnmarshalJSON(data []byte) (err error) {

	value := strings.Trim(string(data), `"`)

	tm, err := time.Parse("2006-01-02T15:04", value)
	if err != nil {
		return
	}

	*ct = CivilTime(tm)

	return
}

func (ct CivilTime) MarshalJSON() (data []byte, err error) {

	str := time.Time(ct).Format("2006-01-02T15:04")
	data = fmt.Appendf(nil, `"%s"`, str)

	return
}
