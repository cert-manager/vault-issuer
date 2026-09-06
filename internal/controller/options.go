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
	"errors"
	"time"

	"github.com/cert-manager/issuer-lib/api/v1alpha1"
	"github.com/cert-manager/issuer-lib/intree"
)

// Defaults match cert-manager's in-tree controller so that an operator
// migrating from the in-tree Vault issuer gets identical behaviour without
// passing any flags.
const (
	defaultClusterResourceNamespace        = "kube-system"
	defaultIssuerAmbientCredentials        = false
	defaultClusterIssuerAmbientCredentials = true
	defaultMaxRetryDuration                = 2 * time.Minute
)

// Options configures issuance behaviour that is not expressible in the Issuer
// API itself.
type Options struct {
	// ClusterResourceNamespace is the namespace Secrets are read from when the
	// Issuer is cluster-scoped and therefore has no namespace of its own.
	ClusterResourceNamespace string

	// IssuerAmbientCredentials allows namespaced Issuers to authenticate using
	// credentials drawn from the controller's environment (instance metadata,
	// IRSA, local files) rather than from the Issuer spec.
	IssuerAmbientCredentials bool

	// ClusterIssuerAmbientCredentials is the same, for ClusterIssuers.
	ClusterIssuerAmbientCredentials bool

	// MaxRetryDuration bounds how long a CertificateRequest is retried, measured
	// from its creationTimestamp. Once exceeded, the request is failed
	// permanently and cert-manager must create a new one.
	MaxRetryDuration time.Duration
}

// NewDefaultOptions returns Options populated with the in-tree defaults.
func NewDefaultOptions() Options {
	return Options{
		ClusterResourceNamespace:        defaultClusterResourceNamespace,
		IssuerAmbientCredentials:        defaultIssuerAmbientCredentials,
		ClusterIssuerAmbientCredentials: defaultClusterIssuerAmbientCredentials,
		MaxRetryDuration:                defaultMaxRetryDuration,
	}
}

// Validate returns an error if the options cannot be used.
func (o Options) Validate() error {
	if o.ClusterResourceNamespace == "" {
		return errors.New("--cluster-resource-namespace must be set; it could not be inferred from the environment")
	}

	if o.MaxRetryDuration <= 0 {
		return errors.New("--max-retry-duration must be a positive duration")
	}

	return nil
}

// ResourceNamespace returns t referenced by the
// given Issuer are looked up.
func (o Options) ResourceNamespace(issuerObject v1alpha1.Issuer) string {
	if ns := issuerObject.GetNamespace(); ns != "" {
		return ns
	}

	return o.ClusterResourceNamespace
}

// CanUseAmbientCredentials rer is permitted to
// authenticate using credentials drawn from the controller's environment.
func (o Options) CanUseAmbientCredentials(issuerObject v1alpha1.Issuer) bool {
	switch issuerObject.(type) {
	case *intree.CMClusterIssuer:
		return o.ClusterIssuerAmbientCredentials
	case *intree.CMIssuer:
		return o.IssuerAmbientCredentials
	}

	return false
}
