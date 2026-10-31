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

// Package config holds the suite-wide settings.
//
// cert-manager's equivalent is driven by a large set of command line flags,
// because its e2e run installs and configures everything it tests. Here the
// make targets do the installing, so this only carries what the tests cannot
// discover for themselves.
package config

import "os"

// EnvClusterResourceNamespace names the namespace the vault-issuer controller
// was started with as --cluster-resource-namespace. Secrets referenced by a
// ClusterIssuer must live there.
const EnvClusterResourceNamespace = "E2E_CLUSTER_RESOURCE_NAMESPACE"

const defaultClusterResourceNamespace = "cert-manager"

// Config is the suite-wide configuration.
type Config struct {
	Addons AddonConfig
}

// AddonConfig groups settings by the component they describe.
type AddonConfig struct {
	CertManager CertManagerConfig
}

// CertManagerConfig describes the cert-manager installation under test.
type CertManagerConfig struct {
	// ClusterResourceNamespace is where the controller looks for Secrets that
	// a ClusterIssuer references.
	ClusterResourceNamespace string
}

// Default returns the configuration built from the environment, falling back to
// the values the make targets install with.
func Default() *Config {
	ns := os.Getenv(EnvClusterResourceNamespace)
	if ns == "" {
		ns = defaultClusterResourceNamespace
	}

	return &Config{
		Addons: AddonConfig{
			CertManager: CertManagerConfig{
				ClusterResourceNamespace: ns,
			},
		},
	}
}
