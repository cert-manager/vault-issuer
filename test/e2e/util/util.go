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

// Package util holds the wait and object-construction helpers the suites share.
// The signatures match cert-manager's e2e util package.
package util

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

const (
	pollInterval = 500 * time.Millisecond
	pollTimeout  = time.Minute
)

// WaitForIssuerCondition waits for the named Issuer to report a condition whose
// type and status match the supplied one.
func WaitForIssuerCondition(ctx context.Context, client clientset.IssuerInterface, name string, condition cmapi.IssuerCondition) error {
	logf, done := log.LogBackoff()
	defer done()

	pollErr := wait.PollUntilContextTimeout(ctx, pollInterval, pollTimeout, true, func(ctx context.Context) (bool, error) {
		logf("Waiting for issuer %v condition %#v", name, condition)

		issuer, err := client.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Issuer %q: %w", name, err)
		}

		return apiutil.IssuerHasCondition(issuer, condition), nil
	})

	if pollErr == nil {
		return nil
	}

	issuer, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return pollErr
	}

	return withConditionDetail(pollErr, issuer.GetStatus().Conditions, condition.Type)
}

// WaitForClusterIssuerCondition waits for the named ClusterIssuer to report a
// condition whose type and status match the supplied one.
func WaitForClusterIssuerCondition(ctx context.Context, client clientset.ClusterIssuerInterface, name string, condition cmapi.IssuerCondition) error {
	logf, done := log.LogBackoff()
	defer done()

	pollErr := wait.PollUntilContextTimeout(ctx, pollInterval, pollTimeout, true, func(ctx context.Context) (bool, error) {
		logf("Waiting for clusterissuer %v condition %#v", name, condition)

		issuer, err := client.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting ClusterIssuer %q: %w", name, err)
		}

		return apiutil.IssuerHasCondition(issuer, condition), nil
	})

	if pollErr == nil {
		return nil
	}

	issuer, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return pollErr
	}

	return withConditionDetail(pollErr, issuer.GetStatus().Conditions, condition.Type)
}

// withConditionDetail appends the last observed condition to a timeout error,
// which is otherwise indistinguishable between "never reconciled" and
// "reconciled and failed".
func withConditionDetail(pollErr error, conditions []cmapi.IssuerCondition, conditionType cmapi.IssuerConditionType) error {
	for _, cond := range conditions {
		if cond.Type == conditionType {
			return fmt.Errorf(
				"%w: last status: %q, reason: %q, message: %q",
				pollErr, cond.Status, cond.Reason, cond.Message,
			)
		}
	}

	return pollErr
}

// NewCertManagerVaultCertificate returns a Certificate that requests
// test.domain.com from the named issuer.
func NewCertManagerVaultCertificate(name, secretName, issuerName string, issuerKind string, duration *metav1.Duration, renewBefore *metav1.Duration) *cmapi.Certificate {
	return &cmapi.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: cmapi.CertificateSpec{
			CommonName:  "test.domain.com",
			SecretName:  secretName,
			Duration:    duration,
			RenewBefore: renewBefore,
			IssuerRef: cmmeta.IssuerReference{
				Name: issuerName,
				Kind: issuerKind,
			},
		},
	}
}
