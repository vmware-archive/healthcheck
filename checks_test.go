// Copyright 2017 by the contributors.
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

package healthcheck

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestTCPDialCheck(t *testing.T) {
	ctx := context.Background()

	assert.NoError(t, TCPDialCheck("heptio.com:80", 5*time.Second)(ctx))
	assert.Error(t, TCPDialCheck("heptio.com:25327", 5*time.Second)(ctx))
}

func TestHTTPGetCheck(t *testing.T) {
	ctx := context.Background()
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()

	assert.NoError(t, HTTPGetCheck("https://heptio.com", 5*time.Second)(ctx))
	assert.Error(t, HTTPGetCheck("https://heptio.com", 5*time.Second)(canceledCtx))
	assert.Error(t, HTTPGetCheck("http://heptio.com", 5*time.Second)(ctx), "redirect should fail")
	assert.Error(t, HTTPGetCheck("https://heptio.com/nonexistent", 5*time.Second)(ctx), "404 should fail")
}

func TestDatabasePingCheck(t *testing.T) {
	ctx := context.Background()
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()

	assert.Error(t, DatabasePingCheck(nil, 1*time.Second)(ctx), "nil DB should fail")

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	assert.NoError(t, DatabasePingCheck(db, 1*time.Second)(ctx), "ping should succeed")
	assert.Error(t, DatabasePingCheck(db, 1*time.Second)(canceledCtx), "ping should fail")
}

func TestDNSResolveCheck(t *testing.T) {
	ctx := context.Background()
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()

	assert.NoError(t, DNSResolveCheck("heptio.com", 5*time.Second)(ctx))
	assert.Error(t, DNSResolveCheck("nonexistent.heptio.com", 5*time.Second)(ctx))
	assert.Error(t, DNSResolveCheck("heptio.com", 5*time.Second)(canceledCtx))
}

func TestGoroutineCountCheck(t *testing.T) {
	assert.NoError(t, GoroutineCountCheck(1000)(context.Background()))
	assert.Error(t, GoroutineCountCheck(0)(context.Background()))
}
