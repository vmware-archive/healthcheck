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
	"context"
	"sync"
)

// WaitHook creates a shutdown hook linked to a context.Context and a
// sync.WaitGroup. Whenever the shutdown hook is called, it signals that event
// by canceling the Context and waiting until the WaitGroup is done.
func WaitHook() (context.Context, *sync.WaitGroup, Hook) {
	return WaitHookWithContext(context.Background())
}

// WaitHookWithContext creates a shutdown hook linked to a context.Context and a
// sync.WaitGroup. Whenever the shutdown hook is called, it signals that event
// by canceling the Context and waiting until the WaitGroup is done. The returned
// context is a child of the specified parent context.
func WaitHookWithContext(parent context.Context) (context.Context, *sync.WaitGroup, Hook) {
	ctx, cancel := context.WithCancel(parent)
	var wg sync.WaitGroup
	return ctx, &wg, func() error {
		cancel()
		wg.Wait()
		return nil
	}
}
