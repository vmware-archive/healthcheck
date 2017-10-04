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
	"encoding/json"
	"net/http"
	"sync"
)

// basicHandler is a basic Handler implementation.
type basicHandler struct {
	http.ServeMux
	mutex           sync.RWMutex
	livenessChecks  map[string]Check
	readinessChecks map[string]Check
	shutdownHooks   map[string]Hook
}

// NewHandler creates a new basic Handler
func NewHandler() Handler {
	h := &basicHandler{
		livenessChecks:  make(map[string]Check),
		readinessChecks: make(map[string]Check),
		shutdownHooks:   make(map[string]Hook),
	}
	h.Handle("/live", http.HandlerFunc(h.LiveEndpoint))
	h.Handle("/ready", http.HandlerFunc(h.ReadyEndpoint))
	h.Handle("/shutdown", http.HandlerFunc(h.ShutdownEndpoint))
	return h
}

func (s *basicHandler) LiveEndpoint(w http.ResponseWriter, r *http.Request) {
	s.handle(w, r, s.livenessChecks)
}

func (s *basicHandler) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {
	s.handle(w, r, s.readinessChecks, s.livenessChecks)
}

func (s *basicHandler) ShutdownEndpoint(w http.ResponseWriter, r *http.Request) {
	results, status := s.collectShutdownHooks()

	// TODO: write to the termination log (/dev/termination-log)

	// write out the response code and content type header
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// unless ?full=1, return an empty body. Kubernetes only cares about the
	// HTTP status code, so we won't waste bytes on the full body.
	if r.URL.Query().Get("full") != "1" {
		w.Write([]byte("{}\n"))
		return
	}

	// otherwise, write the JSON body ignoring any encoding errors (which
	// shouldn't really be possible since we're encoding a map[string]string).
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(results)
}

// collectShutdownHooks runs all the shutdown hooks in parallel and collects
// their results into a map[string]string and an HTTP status code
func (s *basicHandler) collectShutdownHooks() (map[string]string, int) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	// make a result channel big enough to buffer all the results
	type hookResult struct {
		name string
		err  error
	}
	results := make(chan hookResult, len(s.shutdownHooks))

	// kick off all the hooks
	for name, hook := range s.shutdownHooks {
		go func(name string, hook Hook) {
			results <- hookResult{name, hook()}
		}(name, hook)
	}

	// read the results into a map
	status := http.StatusOK
	resultsMap := make(map[string]string)

	for i := 0; i < len(s.shutdownHooks); i++ {
		result := <-results
		if result.err == nil {
			resultsMap[result.name] = "OK"
		} else {
			resultsMap[result.name] = result.err.Error()
			status = http.StatusServiceUnavailable
		}
	}

	return resultsMap, status
}

func (s *basicHandler) AddLivenessCheck(name string, check Check) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.livenessChecks[name] = check
}

func (s *basicHandler) AddReadinessCheck(name string, check Check) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.readinessChecks[name] = check
}

func (s *basicHandler) AddShutdownHook(name string, hook Hook) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.shutdownHooks[name] = hook
}

func (s *basicHandler) collectChecks(checks map[string]Check, resultsOut map[string]string, statusOut *int) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for name, check := range checks {
		if err := check(); err != nil {
			*statusOut = http.StatusServiceUnavailable
			resultsOut[name] = err.Error()
		} else {
			resultsOut[name] = "OK"
		}
	}
}

func (s *basicHandler) handle(w http.ResponseWriter, r *http.Request, checks ...map[string]Check) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checkResults := make(map[string]string)
	status := http.StatusOK
	for _, checks := range checks {
		s.collectChecks(checks, checkResults, &status)
	}

	// write out the response code and content type header
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// unless ?full=1, return an empty body. Kubernetes only cares about the
	// HTTP status code, so we won't waste bytes on the full body.
	if r.URL.Query().Get("full") != "1" {
		w.Write([]byte("{}\n"))
		return
	}

	// otherwise, write the JSON body ignoring any encoding errors (which
	// shouldn't really be possible since we're encoding a map[string]string).
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(checkResults)
}
