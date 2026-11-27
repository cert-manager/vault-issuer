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
	"testing"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestIgnoreCertificateRequest(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, cmapi.AddToScheme(scheme))

	issuerGVK := cmapi.SchemeGroupVersion.WithKind(cmapi.IssuerKind)
	clusterIssuerGVK := cmapi.SchemeGroupVersion.WithKind(cmapi.ClusterIssuerKind)

	vaultSpec := cmapi.IssuerSpec{IssuerConfig: cmapi.IssuerConfig{
		Vault: &cmapi.VaultIssuer{Server: "https://vault.example", Path: "pki/sign/example"},
	}}
	selfSignedSpec := cmapi.IssuerSpec{IssuerConfig: cmapi.IssuerConfig{
		SelfSigned: &cmapi.SelfSignedIssuer{},
	}}
	caSpec := cmapi.IssuerSpec{IssuerConfig: cmapi.IssuerConfig{
		CA: &cmapi.CAIssuer{SecretName: "ca"},
	}}

	objects := []client.Object{
		&cmapi.Issuer{ObjectMeta: metav1.ObjectMeta{Name: "vault", Namespace: "team-a"}, Spec: vaultSpec},
		&cmapi.Issuer{ObjectMeta: metav1.ObjectMeta{Name: "selfsigned", Namespace: "team-a"}, Spec: selfSignedSpec},
		&cmapi.ClusterIssuer{ObjectMeta: metav1.ObjectMeta{Name: "vault"}, Spec: vaultSpec},
		&cmapi.ClusterIssuer{ObjectMeta: metav1.ObjectMeta{Name: "ca"}, Spec: caSpec},
	}

	apiErr := errors.New("apiserver unavailable")

	tests := map[string]struct {
		gvk        schema.GroupVersionKind
		issuerName types.NamespacedName
		getErr     error

		wantIgnore bool
		wantErr    error
	}{
		"Issuer backed by Vault is ours": {
			gvk:        issuerGVK,
			issuerName: types.NamespacedName{Namespace: "team-a", Name: "vault"},
			wantIgnore: false,
		},
		"Issuer backed by another in-tree type is ignored": {
			gvk:        issuerGVK,
			issuerName: types.NamespacedName{Namespace: "team-a", Name: "selfsigned"},
			wantIgnore: true,
		},
		"Issuer in a different namespace with the same name is ignored": {
			gvk:        issuerGVK,
			issuerName: types.NamespacedName{Namespace: "team-b", Name: "vault"},
			wantIgnore: true,
		},
		"ClusterIssuer backed by Vault is ours": {
			gvk:        clusterIssuerGVK,
			issuerName: types.NamespacedName{Name: "vault"},
			wantIgnore: false,
		},
		// issuer-lib derives issuerName from the CertificateRequest, which for a
		// ClusterIssuer reference still carries the request's namespace. The
		// lookup must drop it or every ClusterIssuer-backed request is ignored.
		"ClusterIssuer lookup ignores the request namespace": {
			gvk:        clusterIssuerGVK,
			issuerName: types.NamespacedName{Namespace: "team-a", Name: "vault"},
			wantIgnore: false,
		},
		"ClusterIssuer backed by another in-tree type is ignored": {
			gvk:        clusterIssuerGVK,
			issuerName: types.NamespacedName{Name: "ca"},
			wantIgnore: true,
		},
		"missing Issuer is ignored rather than claimed": {
			gvk:        issuerGVK,
			issuerName: types.NamespacedName{Namespace: "team-a", Name: "does-not-exist"},
			wantIgnore: true,
		},
		"missing ClusterIssuer is ignored rather than claimed": {
			gvk:        clusterIssuerGVK,
			issuerName: types.NamespacedName{Name: "does-not-exist"},
			wantIgnore: true,
		},
		"unexpected issuer kind is an error": {
			gvk:        cmapi.SchemeGroupVersion.WithKind("Certificate"),
			issuerName: types.NamespacedName{Namespace: "team-a", Name: "vault"},
			wantIgnore: false,
			wantErr:    errors.New(`unexpected issuer kind "Certificate" for CertificateRequest`),
		},
		"transient API error is surfaced, not treated as ignore": {
			gvk:        issuerGVK,
			issuerName: types.NamespacedName{Namespace: "team-a", Name: "vault"},
			getErr:     apiErr,
			wantIgnore: false,
			wantErr:    apiErr,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...)
			if tc.getErr != nil {
				builder = builder.WithInterceptorFuncs(interceptor.Funcs{
					Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
						return tc.getErr
					},
				})
			}

			s := &Signer{client: builder.Build()}

			ignore, err := s.IgnoreCertificateRequest(t.Context(), nil, tc.gvk, tc.issuerName)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantIgnore, ignore)
		})
	}
}
