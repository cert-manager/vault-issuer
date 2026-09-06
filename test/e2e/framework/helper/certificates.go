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
	"fmt"
	"time"

	apiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	clientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/cert-manager/vault-issuer/test/e2e/framework/log"
)

const pollInterval = 500 * time.Millisecond

// WaitForCertificateToExist waits for the named Certificate to exist.
func (h *Helper) WaitForCertificateToExist(ctx context.Context, namespace, name string, timeout time.Duration) (*cmapi.Certificate, error) {
	return h.waitForCertificateCondition(
		ctx,
		h.CMClient.CertmanagerV1().Certificates(namespace),
		name,
		func(*cmapi.Certificate) bool { return true },
		timeout,
	)
}

// WaitForCertificateReadyAndDoneIssuing waits for the Certificate to be
// Ready=True and no longer Issuing.
//
// The Ready condition is checked against the generation of the Certificate
// passed in, so a condition left over from an earlier revision does not satisfy
// the wait.
func (h *Helper) WaitForCertificateReadyAndDoneIssuing(ctx context.Context, cert *cmapi.Certificate, timeout time.Duration) (*cmapi.Certificate, error) {
	readyTrue := cmapi.CertificateCondition{
		Type:               cmapi.CertificateConditionReady,
		Status:             cmmeta.ConditionTrue,
		ObservedGeneration: cert.Generation,
	}
	issuingTrue := cmapi.CertificateCondition{
		Type:   cmapi.CertificateConditionIssuing,
		Status: cmmeta.ConditionTrue,
	}

	logf, done := log.LogBackoff()
	defer done()

	return h.waitForCertificateCondition(
		ctx,
		h.CMClient.CertmanagerV1().Certificates(cert.Namespace),
		cert.Name,
		func(certificate *cmapi.Certificate) bool {
			if !apiutil.CertificateHasConditionWithObservedGeneration(certificate, readyTrue) {
				logf(
					"Expected Certificate %v condition %v=%v (generation >= %v) but it has: %v",
					certificate.Name,
					readyTrue.Type,
					readyTrue.Status,
					readyTrue.ObservedGeneration,
					certificate.Status.Conditions,
				)
				return false
			}

			if apiutil.CertificateHasCondition(certificate, issuingTrue) {
				logf(
					"Expected Certificate %v condition %v to be missing but it has: %v",
					certificate.Name,
					issuingTrue.Type,
					certificate.Status.Conditions,
				)
				return false
			}

			if certificate.Status.NextPrivateKeySecretName != nil {
				logf(
					"Expected Certificate %v 'next-private-key-secret-name' to be empty but it has: %v",
					certificate.Name,
					*certificate.Status.NextPrivateKeySecretName,
				)
				return false
			}

			return true
		},
		timeout,
	)
}

// WaitForCertificateNotReadyAndDoneIssuing waits for the Certificate to be
// Ready=False and no longer Issuing.
func (h *Helper) WaitForCertificateNotReadyAndDoneIssuing(ctx context.Context, cert *cmapi.Certificate, timeout time.Duration) (*cmapi.Certificate, error) {
	readyFalse := cmapi.CertificateCondition{
		Type:               cmapi.CertificateConditionReady,
		Status:             cmmeta.ConditionFalse,
		ObservedGeneration: cert.Generation,
	}
	issuingTrue := cmapi.CertificateCondition{
		Type:   cmapi.CertificateConditionIssuing,
		Status: cmmeta.ConditionTrue,
	}

	logf, done := log.LogBackoff()
	defer done()

	return h.waitForCertificateCondition(
		ctx,
		h.CMClient.CertmanagerV1().Certificates(cert.Namespace),
		cert.Name,
		func(certificate *cmapi.Certificate) bool {
			if !apiutil.CertificateHasConditionWithObservedGeneration(certificate, readyFalse) {
				logf(
					"Expected Certificate %v condition %v=%v (generation >= %v) but it has: %v",
					certificate.Name,
					readyFalse.Type,
					readyFalse.Status,
					readyFalse.ObservedGeneration,
					certificate.Status.Conditions,
				)
				return false
			}

			return !apiutil.CertificateHasCondition(certificate, issuingTrue)
		},
		timeout,
	)
}

func (h *Helper) waitForCertificateCondition(
	ctx context.Context,
	client clientset.CertificateInterface,
	name string,
	check func(*cmapi.Certificate) bool,
	timeout time.Duration,
) (*cmapi.Certificate, error) {
	var certificate *cmapi.Certificate

	pollErr := wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		var err error

		certificate, err = client.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			certificate = nil
			return false, fmt.Errorf("error getting Certificate %v: %w", name, err)
		}

		return check(certificate), nil
	})

	if pollErr != nil && certificate != nil {
		// Fold the conditions into the error: without kubectl-describe output,
		// they are the only clue about why the wait timed out.
		pollErr = fmt.Errorf(
			"%w; last seen conditions for Certificate %s/%s: %v",
			pollErr, certificate.Namespace, certificate.Name, certificate.Status.Conditions,
		)
	}

	return certificate, pollErr
}
