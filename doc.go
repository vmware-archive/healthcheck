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

/*
Package healthcheck helps you implement Kubernetes liveness and readiness checks
for your application. It supports synchronous checks, asynchronous (background)
checks, and edge-triggered checks.

It also includes a small library of generic checks for DNS, TCP, and HTTP
reachability as well as Goroutine usage.

Check status can be reported as a set of Prometheus metrics for easy
cluster-wide monitoring and alerting.
*/
package healthcheck
