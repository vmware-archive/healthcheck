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
	"errors"
	"sync"
	"time"
)

// ErrNoData is returned if the first call of an Async() wrapped Check has not
// yet returned.
var ErrNoData = errors.New("no data yet")

// Async converts a Check into an asynchronous check that runs in a background
// goroutine at a fixed interval. Note: the spawned goroutine cannot currently
// be stopped.
func Async(check Check, interval time.Duration) Check {
	// define some variables that we'll close over
	result := ErrNoData
	var mutex sync.Mutex

	// spawn a background goroutine to call the check every interval
	go func() {
		// call once right away (time.Tick() doesn't always tick immediately)
		err := check()
		mutex.Lock()
		result = err
		mutex.Unlock()

		// loop forever on time.Tick
		for _ = range time.Tick(interval) {
			err := check()
			mutex.Lock()
			result = err
			mutex.Unlock()
		}
	}()

	// return a Check function that closes over our result and mutex
	return func() error {
		mutex.Lock()
		defer mutex.Unlock()
		return result
	}
}
