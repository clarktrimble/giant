package minlog

import (
	"context"
	"fmt"
)

type MinLog struct{}

func (ml *MinLog) Info(ctx context.Context, msg string, kv ...any) {
	fmt.Printf("INFO: %s", msg)
	for i := 0; i < len(kv)-1; i += 2 {
		fmt.Printf("  %v=%v", kv[i], kv[i+1])
	}
	fmt.Println()
}

func (ml *MinLog) Debug(ctx context.Context, msg string, kv ...any) {
	fmt.Printf("DEBUG: %s", msg)
	for i := 0; i < len(kv)-1; i += 2 {
		fmt.Printf("  %v=%v", kv[i], kv[i+1])
	}
	fmt.Println()
}

func (ml *MinLog) Trace(ctx context.Context, msg string, kv ...any) {
	fmt.Printf("TRACE: %s", msg)
	for i := 0; i < len(kv)-1; i += 2 {
		fmt.Printf("  %v=%v", kv[i], kv[i+1])
	}
	fmt.Println()
}

func (ml *MinLog) Error(ctx context.Context, msg string, err error, kv ...any) {
	fmt.Printf("ERROR: %s err=%v", msg, err)
	for i := 0; i < len(kv)-1; i += 2 {
		fmt.Printf("  %v=%v", kv[i], kv[i+1])
	}
	fmt.Println()
}

func (ml *MinLog) WithFields(ctx context.Context, kv ...any) context.Context {
	return ctx
}
