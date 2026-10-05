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

package controller

import (
	"testing"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
)

func TestVaultIssuerUsesSecret(t *testing.T) {
	ref := cmmeta.SecretKeySelector{LocalObjectReference: cmmeta.LocalObjectReference{Name: "wanted"}}

	// One entry per Secret reference reachable from .spec.vault. If a new
	// reference is added to the API, add it here too, or rotating that Secret
	// will never re-run Check.
	tests := map[string]*cmapi.VaultIssuer{
		"caBundleSecretRef":         {CABundleSecretRef: &ref},
		"clientCertSecretRef":       {ClientCertSecretRef: &ref},
		"clientKeySecretRef":        {ClientKeySecretRef: &ref},
		"auth.tokenSecretRef":       {Auth: cmapi.VaultAuth{TokenSecretRef: &ref}},
		"auth.appRole.secretRef":    {Auth: cmapi.VaultAuth{AppRole: &cmapi.VaultAppRole{SecretRef: ref}}},
		"auth.clientCertificate":    {Auth: cmapi.VaultAuth{ClientCertificate: &cmapi.VaultClientCertificateAuth{SecretName: "wanted"}}},
		"auth.kubernetes.secretRef": {Auth: cmapi.VaultAuth{Kubernetes: &cmapi.VaultKubernetesAuth{SecretRef: ref}}},
	}

	for field, vault := range tests {
		t.Run(field, func(t *testing.T) {
			assert.True(t, vaultIssuerUsesSecret(vault, "wanted"))
			assert.False(t, vaultIssuerUsesSecret(vault, "other"))
		})
	}

	t.Run("nil vault spec", func(t *testing.T) {
		assert.False(t, vaultIssuerUsesSecret(nil, "wanted"))
	})
}
