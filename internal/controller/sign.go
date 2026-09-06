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
	"context"
	"errors"
	"fmt"
	"net/http"

	apiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmerrors "github.com/cert-manager/cert-manager/pkg/util/errors"
	v1alpha1 "github.com/cert-manager/issuer-lib/api/v1alpha1"
	"github.com/cert-manager/issuer-lib/controllers/signer"
	"github.com/cert-manager/issuer-lib/intree"
	vaultapi "github.com/hashicorp/vault/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (s *Signer) Sign(ctx context.Context, cr signer.CertificateRequestObject, issuerObject v1alpha1.Issuer) (signer.PEMBundle, error) {
	issuerObj, ok := issuerObject.(intree.CMGenericIssuer)
	if !ok {
		return signer.PEMBundle{}, signer.PermanentError{Err: fmt.Errorf("issuer does not implement CMGenericIssuer")}
	}

	genericIssuer, ok := issuerObj.Unwrap().(cmapi.GenericIssuer)
	if !ok {
		return signer.PEMBundle{}, fmt.Errorf("issuer %T does not implement cmapi.GenericIssuer", issuerObj.Unwrap())
	}

	resourceNamespace := s.Options.ResourceNamespace(issuerObj)

	client, err := s.vaultClientBuilder(ctx, resourceNamespace, s.createTokenFn, s.Reader, genericIssuer, s.Options.CanUseAmbientCredentials(issuerObj))
	if err != nil {
		if cmerrors.IsInvalidData(err) {
			return signer.PEMBundle{}, signer.IssuerError{
				Err: signer.PermanentError{Err: fmt.Errorf("%s: %w", messageVaultClientInitFailed, err)},
			}
		}

		return signer.PEMBundle{}, signer.IssuerError{
			Err: fmt.Errorf("%s: %w", messageVaultClientInitFailed, err),
		}
	}

	certDetails, err := cr.GetCertificateDetails()
	if err != nil {
		return signer.PEMBundle{}, signer.PermanentError{
			Err: fmt.Errorf("failed to read certificate details from the request: %w", err),
		}
	}

	certDuration := apiutil.DefaultCertDuration(&metav1.Duration{Duration: certDetails.Duration})
	certPem, caPem, err := client.Sign(certDetails.CSR, certDuration)
	if err != nil {
		return signer.PEMBundle{}, classifySigningError(err)
	}

	log.FromContext(ctx).V(1).Info("certificate issued", "issuer", issuerObj.GetName())

	return signer.PEMBundle{
		CAPEM:    caPem,
		ChainPEM: certPem,
	}, nil
}

// classifySigningError decides whether a failure returned by Vault's sign
// endpoint is worth retrying.
//
// Vault rejects a CSR with a 4xx when the PKI role forbids it — a wrong common
// name, a TTL beyond max_ttl, a disallowed SAN. Retrying cannot change that, so
// the request fails permanently and cert-manager creates a new one. Everything
// else (5xx, a sealed server, a dropped connection) is treated as transient.
func classifySigningError(err error) error {
	var respErr *vaultapi.ResponseError
	if errors.As(err, &respErr) &&
		respErr.StatusCode >= http.StatusBadRequest &&
		respErr.StatusCode < http.StatusInternalServerError {
		return signer.PermanentError{
			Err: fmt.Errorf("vault refused to sign the certificate: %w", err),
		}
	}

	return fmt.Errorf("failed to sign certificate: %w", err)
}
