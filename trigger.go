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

import "sync"

type basicTrigger struct {
	state error
	mutex sync.RWMutex
}

// NewTrigger creates a basic Trigger
func NewTrigger() Trigger {
	return &basicTrigger{}
}

func (t *basicTrigger) Trip(err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.state = err
}

func (t *basicTrigger) Reset() {
	t.Trip(nil)
}

func (t *basicTrigger) Check() Check {
	return func() error {
		t.mutex.RLock()
		defer t.mutex.RUnlock()
		return t.state
	}
}
