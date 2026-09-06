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

package vault

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Environment variables exported by the `make e2e-setup-vault` target. They are
// the contract between the make targets, which install and configure Vault, and
// the tests, which only consume it.
const (
	EnvURL            = "E2E_VAULT_URL"
	EnvNamespace      = "E2E_VAULT_NAMESPACE"
	EnvPod            = "E2E_VAULT_POD"
	EnvCAFile         = "E2E_VAULT_CA_FILE"
	EnvClientCertFile = "E2E_VAULT_CLIENT_CERT_FILE"
	EnvClientKeyFile  = "E2E_VAULT_CLIENT_KEY_FILE"
	EnvEnforceMtls    = "E2E_VAULT_ENFORCE_MTLS"
)

// Details describes the Vault server the tests run against.
//
// It replaces the identically named struct from cert-manager's Vault addon.
// The field names are the ones setup.go reads; do not rename them without
// re-checking that file, which is kept byte-identical to upstream.
type Details struct {
	// URL is the address of Vault from inside the cluster. It is what goes into
	// an Issuer's spec.vault.server.
	URL string

	// ProxyURL is the address of Vault from the test process, i.e. through the
	// port-forward started by Addon.Setup.
	ProxyURL string

	// VaultCA is the CA that signed Vault's serving certificate.
	VaultCA []byte

	// VaultClientCertificate and VaultClientPrivateKey are presented to Vault
	// when it enforces mTLS.
	VaultClientCertificate []byte
	VaultClientPrivateKey  []byte

	// EnforceMtls reports whether the server requires client certificates.
	EnforceMtls bool
}

// Addon owns the connection to a Vault server that was installed outside the
// test process. It replaces cert-manager's Helm-installing addon: Setup only
// reads what make exported and opens a port-forward.
type Addon struct {
	// Prefix is prepended to the environment variable names, so that a second
	// Vault installation (for example the mTLS-enforcing one) can be described
	// by its own set of variables.
	Prefix string

	details   Details
	stopProxy func(context.Context) error
}

// Details returns the description of the Vault server. It must not be called
// before Setup.
func (a *Addon) Details() *Details {
	return &a.details
}

// Setup reads the Vault details from the environment and starts a port-forward
// to the Vault pod. The returned Details are complete, including ProxyURL.
func (a *Addon) Setup(ctx context.Context, clientset kubernetes.Interface, kubeConfig *rest.Config) error {
	details, err := a.detailsFromEnv()
	if err != nil {
		return err
	}

	proxyURL, stop, err := StartProxy(
		ctx,
		clientset,
		kubeConfig,
		a.env(EnvNamespace),
		a.env(EnvPod),
	)
	if err != nil {
		return fmt.Errorf("starting the Vault port-forward: %w", err)
	}

	details.ProxyURL = proxyURL

	a.details = details
	a.stopProxy = stop

	return nil
}

// Deprovision stops the port-forward. The Vault server itself outlives the test
// run; `make e2e-setup-vault` owns its lifecycle.
func (a *Addon) Deprovision(ctx context.Context) error {
	if a.stopProxy == nil {
		return nil
	}

	return a.stopProxy(ctx)
}

func (a *Addon) env(name string) string {
	return os.Getenv(a.Prefix + name)
}

func (a *Addon) detailsFromEnv() (Details, error) {
	url := a.env(EnvURL)
	if url == "" {
		return Details{}, fmt.Errorf("%s is not set; run `make e2e-setup-vault` first", a.Prefix+EnvURL)
	}

	caFile := a.env(EnvCAFile)
	ca, err := os.ReadFile(caFile)
	if err != nil {
		return Details{}, fmt.Errorf("reading the Vault CA from %q: %w", caFile, err)
	}

	details := Details{
		URL:     url,
		VaultCA: ca,
	}

	enforceMtls, err := strconv.ParseBool(orDefault(a.env(EnvEnforceMtls), "false"))
	if err != nil {
		return Details{}, fmt.Errorf("parsing %s: %w", a.Prefix+EnvEnforceMtls, err)
	}

	if !enforceMtls {
		return details, nil
	}

	certFile, keyFile := a.env(EnvClientCertFile), a.env(EnvClientKeyFile)

	cert, err := os.ReadFile(certFile)
	if err != nil {
		return Details{}, fmt.Errorf("reading the Vault client certificate from %q: %w", certFile, err)
	}

	key, err := os.ReadFile(keyFile)
	if err != nil {
		return Details{}, fmt.Errorf("reading the Vault client key from %q: %w", keyFile, err)
	}

	details.EnforceMtls = true
	details.VaultClientCertificate = cert
	details.VaultClientPrivateKey = key

	return details, nil
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
