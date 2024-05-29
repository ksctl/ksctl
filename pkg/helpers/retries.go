package helpers

import (
	"context"
	"time"

	"github.com/ksctl/ksctl/pkg/types"
)

type BackOff struct {
	waitT0     time.Duration
	factor     int
	maxRetries int
}

func NewBackOff(initialWaitTime time.Duration, multiplicativeFactor int, maxNoOfRetries int) *BackOff {
	return &BackOff{
		waitT0:     initialWaitTime,
		factor:     multiplicativeFactor,
		maxRetries: maxNoOfRetries,
	}
}

func (b *BackOff) Run(
	ctx context.Context,
	log types.LoggerFactory,
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
				return log.NewError(ctx, "Operation cancelled", "Reason", ctx.Err())
			}
			return nil
		case <-time.After(waitTime):
			waitTime *= time.Duration(b.factor)
		}
	}

	if storePrevErr != nil {
		return log.NewError(ctx, "Max retries exceeded", "Reason", storePrevErr)
	}

	return log.NewError(ctx, "Max retries exceeded")
}
