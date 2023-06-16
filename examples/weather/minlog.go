package main

import (
	"context"
	"fmt"
	"strings"
)

// implement a (sub)minimal logger to give an idea of what's logged

type minLog struct{}

func (ml *minLog) Info(ctx context.Context, msg string, kv ...any) {

	// non-string and blank values are logged as "*"

	strs := []string{}
	for _, korv := range kv {
		str, ok := korv.(string)
		if !ok || str == "" {
			str = "*"
		}
		strs = append(strs, str)
	}

	fmt.Printf("msg > %s  %s\n", msg, strings.Join(strs, "|"))
}

func (ml *minLog) Error(ctx context.Context, msg string, err error, kv ...any) {

	// Todo: include err
	ml.Info(ctx, msg, kv)
}

func (ml *minLog) WithFields(ctx context.Context, kv ...any) context.Context {

	// Todo: stash kv
	return ctx
}
