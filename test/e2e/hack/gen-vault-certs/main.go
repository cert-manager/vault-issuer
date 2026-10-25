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

// Command gen-vault-certs writes the TLS material for a Vault server used by
// the end-to-end tests. It is run by `make e2e-setup-vault`.
package main

import (
	"flag"
	"fmt"
	"os"

	vaultaddon "github.com/cert-manager/vault-issuer/test/e2e/framework/addon/vault"
)

func main() {
	var dir, dnsName string

	flag.StringVar(&dir, "out-dir", "", "Directory to write ca.crt, server.crt, server.key, client.crt and client.key into.")
	flag.StringVar(&dnsName, "dns-name", "", "DNS name to issue the Vault serving certificate for.")
	flag.Parse()

	if dir == "" || dnsName == "" {
		fmt.Fprintln(os.Stderr, "both --out-dir and --dns-name are required")
		os.Exit(2)
	}

	if err := vaultaddon.WriteTLSFiles(dir, dnsName); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write the Vault TLS files: %v\n", err)
		os.Exit(1)
	}
}
