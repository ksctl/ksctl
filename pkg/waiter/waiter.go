// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package waiter

import (
	"context"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
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
				return ksctlErrors.WrapError(
					ksctlErrors.ErrContextCancelled,
					log.NewError(ctx, "backoff termination", "Reason", ctx.Err()),
				)
			}
			return nil
		case <-time.After(waitTime):
			waitTime *= time.Duration(b.factor)
		}
	}

	if storePrevErr != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrTimeOut,
			log.NewError(ctx, "Max backoff retries reached", "Reason", storePrevErr),
		)
	}

	return ksctlErrors.WrapError(
		ksctlErrors.ErrTimeOut,
		log.NewError(ctx, "Max backoff retries reached"),
	)
}
