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

package helper

import (
	"context"
	"crypto"
	"fmt"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/cert-manager/vault-issuer/test/e2e/framework/helper/featureset"
	"github.com/cert-manager/vault-issuer/test/e2e/framework/helper/validation"
	"github.com/cert-manager/vault-issuer/test/e2e/framework/helper/validation/certificaterequests"
	"github.com/cert-manager/vault-issuer/test/e2e/framework/helper/validation/certificates"
)

// ValidateCertificate fetches the Secret the Certificate issued into and runs
// every validation function against the pair.
//
// With no validations given, the set that holds for any issuer is used.
func (h *Helper) ValidateCertificate(certificate *cmapi.Certificate, validations ...certificates.ValidationFunc) error {
	if len(validations) == 0 {
		validations = validation.CertificateSetForUnsupportedFeatureSet(featureset.NewFeatureSet())
	}

	secret, err := h.KubeClient.CoreV1().Secrets(certificate.Namespace).Get(
		context.TODO(),
		certificate.Spec.SecretName,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}

	var errs []error

	for _, fn := range validations {
		if err := fn(certificate, secret); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		errs = append(errs, fmt.Errorf(
			"while validating Certificate %s/%s against Secret %s",
			certificate.Namespace, certificate.Name, certificate.Spec.SecretName,
		))
	}

	return kerrors.NewAggregate(errs)
}

// ValidateCertificateRequest fetches the named CertificateRequest and runs
// every validation function against it.
//
// With no validations given, the set that holds for any issuer is used.
func (h *Helper) ValidateCertificateRequest(name types.NamespacedName, key crypto.Signer, validations ...certificaterequests.ValidationFunc) error {
	if len(validations) == 0 {
		validations = validation.CertificateRequestSetForUnsupportedFeatureSet(featureset.NewFeatureSet())
	}

	cr, err := h.CMClient.CertmanagerV1().CertificateRequests(name.Namespace).Get(
		context.TODO(),
		name.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return err
	}

	for _, fn := range validations {
		if err := fn(cr, key); err != nil {
			return fmt.Errorf("while validating CertificateRequest %s: %w", name, err)
		}
	}

	return nil
}
