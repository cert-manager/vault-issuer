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
	"fmt"
	"os"
	"path/filepath"
)

// File names written by WriteTLSFiles. The server and CA names match the keys
// the Vault Helm chart expects in the secret it mounts at /vault/tls.
const (
	FileCA         = "ca.crt"
	FileCAKey      = "ca.key"
	FileServerCert = "server.crt"
	FileServerKey  = "server.key"
	FileClientCert = "client.crt"
	FileClientKey  = "client.key"
)

// WriteTLSFiles generates the CA, the Vault serving certificate for dnsName,
// and a client certificate, and writes them into dir.
//
// cert-manager's e2e framework does this in-process, as part of installing
// Vault with Helm. Here the install belongs to the make targets, so the files
// are written to disk: make builds the Secret from them, and the tests read the
// CA and client certificate back through the E2E_VAULT_*_FILE variables.
func WriteTLSFiles(dir, dnsName string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	ca, caKey, err := GenerateCA()
	if err != nil {
		return fmt.Errorf("generating the CA: %w", err)
	}

	serverCert, serverKey := generateVaultServingCert(ca, caKey, dnsName)
	clientCert, clientKey := generateVaultClientCert(ca, caKey)

	for name, contents := range map[string][]byte{
		FileCA:         ca,
		FileCAKey:      caKey,
		FileServerCert: serverCert,
		FileServerKey:  serverKey,
		FileClientCert: clientCert,
		FileClientKey:  clientKey,
	} {
		// The private keys are test fixtures for a throwaway cluster, but there
		// is no reason to make them world-readable.
		if err := os.WriteFile(filepath.Join(dir, name), contents, 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}

	return nil
}
