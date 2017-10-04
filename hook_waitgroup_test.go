// Copyright © 2017 Heptio
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

package healthcheck

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitHook(t *testing.T) {
	ctx, wg, hook := WaitHook()

	// expect the context to be open (not done)
	select {
	case <-ctx.Done():
		assert.FailNow(t, "expected context not to be done yet")
	default:
	}

	// expect this not to block
	wg.Wait()

	// add a unit of in-flight work
	wg.Add(1)

	// call hook, which should block
	result := make(chan error)
	go func() {
		result <- hook()
	}()

	// expect the context to be closed
	select {
	case <-ctx.Done():
	case <-time.After(time.Millisecond):
		assert.FailNow(t, "expected context to be done")
	}

	// expect the hook to still be blocked
	select {
	case <-result:
		assert.FailNow(t, "expected hook to be blocked while work is in flight")
	default:
	}

	// finish the in-flight work
	wg.Done()

	// expect the hook to have returned nil now that the WorkGroup is done with in-flight work
	select {
	case result := <-result:
		assert.NoErrorf(t, result, "expected hook to have returned nil")
	case <-time.After(time.Millisecond):
		assert.FailNow(t, "expected hook to have returned a result")
	}
}
