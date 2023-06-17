package main

import (
	"context"
	"fmt"
)

// implement a (sub)minimal logger
// see https://github.com/clarktrimble/sabot for a featureful implementation

type minLog struct{}

func (ml *minLog) Info(ctx context.Context, msg string, kv ...any) {

	fmt.Printf("msg > %s\n", msg)
}

func (ml *minLog) Error(ctx context.Context, msg string, err error, kv ...any) {

	ml.Info(ctx, msg, kv)
}

func (ml *minLog) WithFields(ctx context.Context, kv ...any) context.Context {

	return ctx
}
