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

func TestTCPDialCheck(t *testing.T) {
	if err := TCPDialCheck("heptio.com:80", 5*time.Second)(); err != nil {
		t.Errorf("expected TCP check to heptio.com:80 to succeed, got %v", err)
	}

	if err := TCPDialCheck("heptio.com:25327", 5*time.Second)(); err == nil {
		t.Errorf("expected TCP check to heptio.com:25327 fail, but got nil error")
	}
}

func TestHTTPGetCheck(t *testing.T) {
	if err := HTTPGetCheck("https://heptio.com", 5*time.Second)(); err != nil {
		t.Errorf("expected HTTP GET check to http://heptio.com to succeed, got %v", err)
	}

	if err := HTTPGetCheck("http://heptio.com", 5*time.Second)(); err == nil {
		t.Errorf("expected HTTP GET check to fail on a redirect (http->https) but got nil error")
	}

	if err := HTTPGetCheck("https://heptio.com/nonexistent", 5*time.Second)(); err == nil {
		t.Errorf("expected HTTP GET check to http://heptio.com/nonexistent to fail, but got nil error")
	}
}

func TestDNSResolveCheck(t *testing.T) {
	if err := DNSResolveCheck("heptio.com", 5*time.Second)(); err != nil {
		t.Errorf("expected DNS lookup for heptio.com to succeed, got %v", err)
	}

	if err := DNSResolveCheck("nonexistent.heptio.com", 5*time.Second)(); err == nil {
		t.Errorf("expected DNS lookup for nonexistent.heptio.com to fail, but got nil error")
	}
}

func TestGoroutineCountCheck(t *testing.T) {
	if err := GoroutineCountCheck(1000)(); err != nil {
		t.Errorf("expected goroutine count check to succeed, but got %v", err)
	}

	if err := GoroutineCountCheck(0)(); err == nil {
		t.Errorf("expected goroutine count check to fail, but got nil error")
	}
}
