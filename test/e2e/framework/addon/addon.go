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

// Package addon exposes the components the suites run against.
//
// Unlike cert-manager's addons, nothing here installs anything: the make
// targets create the cluster, cert-manager, Vault and the controller before the
// tests start. These values only describe what is already running.
package addon

import (
	"context"

	"github.com/cert-manager/vault-issuer/test/e2e/framework/addon/base"
	"github.com/cert-manager/vault-issuer/test/e2e/framework/addon/vault"
)

// The suites reference these as package-level values, matching cert-manager's
// e2e tests. They are populated once by Setup, from the test suite's
// SynchronizedBeforeSuite, and are read-only thereafter.
var (
	// Base holds the cluster connection.
	Base = &base.Base{}

	// Vault describes the Vault server installed by `make e2e-setup-vault`.
	Vault = &vault.Addon{}

	// VaultEnforceMtls describes the Vault server installed by
	// `make e2e-setup-vault-mtls`, which requires client certificates.
	VaultEnforceMtls = &vault.Addon{Prefix: "MTLS_"}
)

// Setup prepares every addon. Call it once per test process, before any spec
// runs.
func Setup(ctx context.Context) error {
	if err := Base.Setup(); err != nil {
		return err
	}

	details := Base.Details()

	for _, addon := range []*vault.Addon{Vault, VaultEnforceMtls} {
		if err := addon.Setup(ctx, details.KubeClient, details.KubeConfig); err != nil {
			return err
		}
	}

	return nil
}

// Deprovision releases whatever Setup opened. The installed components outlive
// the test process.
func Deprovision(ctx context.Context) error {
	var firstErr error

	for _, addon := range []*vault.Addon{Vault, VaultEnforceMtls} {
		if err := addon.Deprovision(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
