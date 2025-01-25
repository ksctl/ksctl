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
	"errors"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"gotest.tools/v3/assert"
	"os"
	"testing"
	"time"
)

var (
	log logger.Logger = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestWaiterRun_SuccessOnFirstAttempt(t *testing.T) {
	ctx := context.Background()

	executeFunc := func() error {
		return nil
	}

	isSuccessful := func() bool {
		return true
	}

	errorFunc := func(err error) (error, bool) {
		return nil, false
	}

	successFunc := func() error {
		return nil
	}

	backOff := NewWaiter(1*time.Second, 1, 3)

	err := backOff.Run(ctx, log, executeFunc, isSuccessful, errorFunc, successFunc, "Waiting message")
	assert.Assert(t, err == nil)
}

func TestWaiterRun_RetryOnFailure(t *testing.T) {
	ctx := context.Background()

	callCount := 0
	executeFunc := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("execute error")
		}
		return nil
	}

	isSuccessful := func() bool {
		return callCount == 3
	}

	errorFunc := func(err error) (error, bool) {
		return nil, false
	}

	successFunc := func() error {
		return nil
	}

	backOff := NewWaiter(1*time.Second, 1, 3)

	err := backOff.Run(ctx, log, executeFunc, isSuccessful, errorFunc, successFunc, "Waiting message")
	assert.Assert(t, err == nil)

	assert.Equal(t, 3, callCount)
}

func TestWaiterRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	executeFunc := func() error {
		return errors.New("execute error")
	}

	isSuccessful := func() bool {
		return false
	}

	errorFunc := func(err error) (error, bool) {
		return nil, false
	}

	successFunc := func() error {
		return nil
	}

	backOff := NewWaiter(1*time.Second, 1, 3)

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	err := backOff.Run(ctx, log, executeFunc, isSuccessful, errorFunc, successFunc, "Waiting message")
	assert.Assert(t, err != nil && ksctlErrors.IsContextCancelled(err))

	assert.Equal(t, context.Canceled, ctx.Err())
}

func TestWaiterRun_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()

	executeFunc := func() error {
		return errors.New("execute error")
	}

	isSuccessful := func() bool {
		return false
	}

	errorFunc := func(err error) (error, bool) {
		return nil, false
	}

	successFunc := func() error {
		return nil
	}

	backOff := NewWaiter(1*time.Second, 1, 3)

	err := backOff.Run(ctx, log, executeFunc, isSuccessful, errorFunc, successFunc, "Waiting message")
	assert.Assert(t, err != nil && ksctlErrors.IsTimeout(err))
}
