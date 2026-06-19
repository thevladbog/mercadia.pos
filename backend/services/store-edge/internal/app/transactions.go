package app

import "context"

type TransactionRunner interface {
	Run(ctx context.Context, fn func(context.Context) error) error
}

func RunTransaction(ctx context.Context, runner TransactionRunner, fn func(context.Context) error) error {
	if runner == nil {
		return fn(ctx)
	}
	return runner.Run(ctx, fn)
}
