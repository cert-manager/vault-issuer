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
	"fmt"

	"github.com/cert-manager/issuer-lib/api/v1alpha1"
	"github.com/cert-manager/issuer-lib/controllers"
	"github.com/cert-manager/issuer-lib/intree"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	vaultinternal "github.com/cert-manager/vault-issuer/internal/vault"
)

// FieldOwner is the field manager used for all server-side apply patches made
// by this controller. It must stay stable across releases: changing it makes
// the apiserver treat previously owned fields as owned by a stranger.
const FieldOwner = "vault.cert-manager.io"

// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests,verbs=get;list;watch
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests/status,verbs=patch
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers;clusterissuers,verbs=get;list;watch
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers/status;clusterissuers/status,verbs=patch

// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/status,verbs=patch
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,verbs=sign,resourceNames=issuers.cert-manager.io/*;clusterissuers.cert-manager.io/*

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// issuer-lib records events through k8s.io/client-go/tools/events, which writes
// to events.k8s.io/v1 rather than the legacy core API.
// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=create;patch

// Signer implements the issuer-lib Check and Sign functions for Vault-backed
// cert-manager Issuers and ClusterIssuers.
type Signer struct {
	Options
	Reader             client.Reader
	createTokenFn      func(ns string) vaultinternal.CreateToken
	vaultClientBuilder vaultinternal.ClientBuilder
}

// New returns a Signer that talks to a real Vault server.
func New(opts Options) *Signer {
	return &Signer{
		Options:            opts,
		vaultClientBuilder: vaultinternal.New,
	}
}

// SetupWithManager registers the Issuer, CertificateRequest and Kubernetes CSR
// controllers with the manager.
func (s *Signer) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		return fmt.Errorf("building kubernetes clientset: %w", err)
	}

	s.createTokenFn = func(ns string) vaultinternal.CreateToken {
		return clientset.CoreV1().ServiceAccounts(ns).CreateToken
	}

	s.Reader = mgr.GetAPIReader()

	return (&controllers.CombinedController{
		IssuerTypes:        intree.Issuers,
		ClusterIssuerTypes: intree.ClusterIssuers,

		FieldOwner:       FieldOwner,
		MaxRetryDuration: s.MaxRetryDuration,

		Check: s.Check,
		Sign:  s.Sign,

		// cert-manager's Issuer CRD is shared with every other in-tree issuer
		// type, so anything without a .spec.vault belongs to someone else.
		IgnoreIssuer: s.IgnoreIssuer,

		// Preserves the in-tree behaviour of populating the CertificateRequest's
		// status.ca (and therefore the Secret's ca.crt).
		//nolint:staticcheck // Deliberate: the in-tree Vault issuer populated
		// status.ca, so leaving this off would silently empty ca.crt in every
		// Secret on the next renewal.
		SetCAOnCertificateRequest: true,

		EventRecorder: mgr.GetEventRecorder(FieldOwner),

		// issuer-lib only watches the Issuer itself, so a successful Check
		// latches the Issuer at Ready until its .spec changes. Watching the
		// Secrets it references makes credential and CA-bundle rotation
		// re-run Check.
		PreSetupWithManager: s.preSetupWithManager,
	}).SetupWithManager(ctx, mgr)
}

// IgnoreIssuer implements signer.IgnoreIssuer. Returning true leaves the Issuer
// untouched — no conditions written, no reconcile.
func (s *Signer) IgnoreIssuer(ctx context.Context, issuerObject v1alpha1.Issuer) (bool, error) {
	iss, ok := issuerObject.(intree.CMGenericIssuer)
	if !ok {
		return false, fmt.Errorf("issuer %T does not implement intree.CMGenericIssuissuerObject", issuerObject)
	}

	return iss.IssuerSpec().Vault == nil, nil
}
