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
	"time"
)

// httpDrainer wraps an http.Handler to provide a shutdown hook that waits for
// all active requests to drain.
type httpDrainer struct {
	handler          http.Handler
	minimumQuietTime time.Duration

	mutex    sync.Mutex
	inflight int
	quiet    *time.Timer
}

// ServeHTTP wraps the underlying http.Handler to atomically update the number
// of requests in flight
func (d *httpDrainer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	d.mutex.Lock()
	d.inflight++
	// before serving the request, cancel the timer
	// (it can't fire while this request is in flight)
	d.quiet.Stop()
	d.mutex.Unlock()

	defer func() {
		d.mutex.Lock()
		d.inflight--
		if d.inflight == 0 {
			// if we're the last active request, reset the timer to fire after the quiet time
			d.quiet.Reset(d.minimumQuietTime)
		}
		d.mutex.Unlock()
	}()
	d.handler.ServeHTTP(w, r)
}

// hook blocks until there are no active requests
func (d *httpDrainer) hook() error {
	<-d.quiet.C
	return nil
}

// HTTPDrainHook returns a wrapped http.Handler and a shutdown Hook. It will
// pass through all requests to the wrapped handler but will track requests in
// flight. On graceful shutdown, it will block until it has been at least
// minimumQuietTime since the last request completed.
func HTTPDrainHook(handler http.Handler, minimumQuietTime time.Duration) (http.Handler, Hook) {
	d := &httpDrainer{
		handler:          handler,
		minimumQuietTime: minimumQuietTime,
		quiet:            time.NewTimer(minimumQuietTime),
	}
	return d, d.hook
}
