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
)

func TestAsync(t *testing.T) {
	async := Async(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}, 1*time.Millisecond)

	// expect the first call to return ErrNoData since it takes 50ms to return the first time
	if err := async(); err != ErrNoData {
		t.Errorf("expected ErrNoData, but got %v", err)
		t.FailNow()
	}

	// wait for the first run to finish
	time.Sleep(100 * time.Millisecond)

	// make sure the next call returns nil ~immediately
	start := time.Now()
	if err := async(); err != nil {
		t.Errorf("unexpected error %v", err)
		t.FailNow()
	}
	latency := time.Since(start)
	if latency > (1 * time.Millisecond) {
		t.Errorf("unexpected async() to return almost immediately, but it took %v", latency)
		t.FailNow()
	}
}
