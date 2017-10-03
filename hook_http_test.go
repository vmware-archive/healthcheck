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
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const workTime = 500 * time.Millisecond
const quietTime = 1 * time.Second
const acceptableDelta = 10 * time.Millisecond

func slowHandler() http.Handler {
	// create an http.Handler where GET / takes 1s to response
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(workTime)
		w.WriteHeader(http.StatusOK)
	})
	return mux

}

func TestHTTPDrainHookEmpty(t *testing.T) {

	_, hook := HTTPDrainHook(slowHandler(), quietTime)

	// make sure hook() waits for the quietTime to drain
	start := time.Now()
	hook()
	end := time.Now()
	assert.WithinDuration(t, start.Add(quietTime), end, acceptableDelta,
		"expected hook() to return after %s but it took %s",
		quietTime.String(),
		end.Sub(start).String())
}

func TestHTTPDrainHook(t *testing.T) {
	handler, hook := HTTPDrainHook(slowHandler(), quietTime)
	// kick off a few in-flight requests
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			assert.HTTPSuccess(t, handler.ServeHTTP, "GET", "/", nil)
			wg.Done()
		}()
	}

	// make sure hook() takes about workTime+quietTime to respond (to finish all
	// in-flight requests and wait for the quiet time to expire)
	start := time.Now()
	hook()
	end := time.Now()
	assert.WithinDuration(t, start.Add(workTime+quietTime), end, acceptableDelta,
		"expected hook() to return after %s but it took %s",
		quietTime.String(),
		end.Sub(start).String())

	// wait for all the requests to finish so we can make sure they all succeeded
	// this should return ~immediately
	wg.Wait()
}
