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

package e2e

import (
	"context"
	"testing"

	"github.com/cert-manager/vault-issuer/test/e2e/framework/addon"
	// Importing the suites registers their specs.
	_ "github.com/cert-manager/vault-issuer/test/e2e/suite/vault"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestE2E runs the end-to-end suite against the cluster in KUBECONFIG. The
// cluster is expected to already have cert-manager, Vault and the vault-issuer
// controller installed; `make e2e-setup` does that.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vault-issuer e2e suite")
}

// BeforeSuite runs once in every parallel process, because each one opens its
// own port-forward to Vault.
var _ = BeforeSuite(func(ctx context.Context) {
	Expect(addon.Setup(ctx)).To(Succeed(), "failed to connect to the test cluster and Vault")
})

var _ = AfterSuite(func(ctx context.Context) {
	Expect(addon.Deprovision(ctx)).To(Succeed(), "failed to tear down the Vault port-forward")
})
