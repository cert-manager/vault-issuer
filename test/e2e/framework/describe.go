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

package framework

import (
	"github.com/onsi/ginkgo/v2"
)

// CertManagerDescribe is a wrapper for ginkgo.Describe that prefixes the spec
// text, so `--ginkgo.focus` can select the whole suite.
func CertManagerDescribe(text string, body func()) bool {
	return ginkgo.Describe("[cert-manager] "+text, body)
}

// ConformanceDescribe is a wrapper for ginkgo.Describe used by the conformance
// suites, which are selected separately from the rest.
func ConformanceDescribe(text string, body func()) bool {
	return ginkgo.Describe("[Conformance] "+text, body)
}
