package transaction

import (
	"context"
	"errors"
)

var ErrTransactionPanic = errors.New("transaction panic")

type Runner interface {
	Run(ctx context.Context, callback func(ctx context.Context) error) error
}
