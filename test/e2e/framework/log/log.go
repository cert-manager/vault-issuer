/*
Copyright 2026 The cert-manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package log writes test progress to Ginkgo's output. It mirrors the small
// part of cert-manager's e2e log package that the Vault suites rely on.
package log

import (
	"fmt"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/util/wait"
)

var Writer = ginkgo.GinkgoWriter

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func logf(level string, format string, args ...any) {
	fmt.Fprintf(Writer, nowStamp()+": "+level+": "+format+"\n", args...)
}

func Logf(format string, args ...any) {
	logf("INFO", format, args...)
}

// LogBackoff returns a logger that backs off exponentially. Callers that poll
// in a tight loop use it so a slow condition does not bury the output in
// thousands of identical lines.
//
// The returned done func must be called when polling finishes.
func LogBackoff() (logf func(format string, args ...any), done func()) {
	backoff := wait.Backoff{
		Duration: 5 * time.Second,
		Factor:   1.2,
		Steps:    10,
		Cap:      1 * time.Minute,
	}

	start := time.Now()
	var msg string
	done = func() {
		Logf(msg + fmt.Sprintf(" (took %v)", time.Since(start).Truncate(time.Second)))
	}

	once := sync.Once{}
	step := time.Now()
	return func(format string, args ...any) {
		msg = fmt.Sprintf(format, args...)
		once.Do(func() {
			Logf(msg)
		})

		if time.Since(step) < backoff.Duration {
			return
		}

		step = time.Now()
		_ = backoff.Step()
		Logf(msg)
	}, done
}
