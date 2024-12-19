package waiter

import (
	"context"
	"github.com/ksctl/ksctl/pkg/logger"
	"time"
)

type Waiter struct {
	waitT0     time.Duration
	factor     int
	maxRetries int
}

func NewWaiter(initialWaitTime time.Duration, multiplicativeFactor int, maxNoOfRetries int) *Waiter {
	return &Waiter{
		waitT0:     initialWaitTime,
		factor:     multiplicativeFactor,
		maxRetries: maxNoOfRetries,
	}
}

func (b *Waiter) Run(
	ctx context.Context,
	log logger.Logger,
	executeFunc func() error,
	isSuccessful func() bool,
	errorFunc func(err error) (errW error, escalateErr bool),
	successFunc func() error,
	messageForWaiting string,
) error {
	waitTime := b.waitT0
	var storePrevErr error

	for i := 0; i <= b.maxRetries; i++ {

		err := executeFunc()
		if err == nil {
			if isSuccessful() {
				return successFunc()
			} else {
				log.Warn(ctx, messageForWaiting, "Attempt", i+1)
			}
		} else {
			storePrevErr = err
			log.Warn(ctx, "Failure", "Attempt", i+1, "failed", err)
			if errorFunc != nil {
				if _e, ok := errorFunc(err); ok {
					return _e
				}
			}
		}

		select {
		case <-ctx.Done():
			log.Print(ctx, "Operation cancelled during backoff")
			if ctx.Err() != nil {
				return ksctlErrors.ErrContextCancelled.Wrap(
					log.NewError(ctx, "backoff termination", "Reason", ctx.Err()),
				)
			}
			return nil
		case <-time.After(waitTime):
			waitTime *= time.Duration(b.factor)
		}
	}

	if storePrevErr != nil {
		return ksctlErrors.ErrTimeOut.Wrap(
			log.NewError(ctx, "Max backoff retries reached", "Reason", storePrevErr),
		)
	}

	return ksctlErrors.ErrTimeOut.Wrap(
		log.NewError(ctx, "Max backoff retries reached"),
	)
}
