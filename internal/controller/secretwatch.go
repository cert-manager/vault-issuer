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

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// preSetupWithManager runs for every controller issuer-lib builds. Only the two
// issuer controllers get the Secret watch; the CertificateRequest and
// Kubernetes CSR controllers are driven by issuer-lib's own event source.
func (s *Signer) preSetupWithManager(_ context.Context, gvk schema.GroupVersionKind, mgr ctrl.Manager, b *builder.Builder) error {
	switch gvk.Kind {
	case cmapi.IssuerKind, cmapi.ClusterIssuerKind:
	default:
		return nil
	}

	// Metadata-only: the mapping needs nothing but the Secret's name and
	// namespace, and a full Secret informer would cache the data of every
	// Secret in the cluster. The Signer itself keeps reading Secret contents
	// through mgr.GetAPIReader(), which bypasses the cache entirely.
	secret := &metav1.PartialObjectMetadata{}
	secret.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

	b.WatchesMetadata(secret, handler.EnqueueRequestsFromMapFunc(
		s.issuersForSecret(mgr.GetClient(), gvk.Kind == cmapi.ClusterIssuerKind),
	))

	return nil
}

// issuersForSecret maps a Secret to every Issuer (or ClusterIssuer) whose
// .spec.vault references it, so that rotating a CA bundle or an auth
// credential re-runs Check instead of leaving the Issuer latched at Ready.
func (s *Signer) issuersForSecret(c client.Client, clusterScoped bool) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		name, namespace := obj.GetName(), obj.GetNamespace()

		if clusterScoped {
			// ClusterIssuers only ever read Secrets from one namespace.
			if namespace != s.Options.ClusterResourceNamespace {
				return nil
			}

			list := &cmapi.ClusterIssuerList{}
			if err := c.List(ctx, list); err != nil {
				log.FromContext(ctx).Error(err, "listing ClusterIssuers for a Secret", "secret", namespace+"/"+name)
				return nil
			}

			var requests []reconcile.Request
			for i := range list.Items {
				if vaultIssuerUsesSecret(list.Items[i].Spec.Vault, name) {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{Name: list.Items[i].Name},
					})
				}
			}

			return requests
		}

		list := &cmapi.IssuerList{}
		if err := c.List(ctx, list, client.InNamespace(namespace)); err != nil {
			log.FromContext(ctx).Error(err, "listing Issuers for a Secret", "secret", namespace+"/"+name)
			return nil
		}

		var requests []reconcile.Request
		for i := range list.Items {
			if vaultIssuerUsesSecret(list.Items[i].Spec.Vault, name) {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{Name: list.Items[i].Name, Namespace: namespace},
				})
			}
		}

		return requests
	}
}

// vaultIssuerUsesSecret reports whether vault references a Secret called name.
// Every Secret reference reachable from .spec.vault is checked, so any of them
// changing re-triggers Check.
func vaultIssuerUsesSecret(vault *cmapi.VaultIssuer, name string) bool {
	if vault == nil {
		return false
	}

	if vault.CABundleSecretRef != nil && vault.CABundleSecretRef.Name == name {
		return true
	}

	if vault.ClientCertSecretRef != nil && vault.ClientCertSecretRef.Name == name {
		return true
	}

	if vault.ClientKeySecretRef != nil && vault.ClientKeySecretRef.Name == name {
		return true
	}

	auth := vault.Auth

	if auth.TokenSecretRef != nil && auth.TokenSecretRef.Name == name {
		return true
	}

	if auth.AppRole != nil && auth.AppRole.SecretRef.Name == name {
		return true
	}

	if auth.ClientCertificate != nil && auth.ClientCertificate.SecretName == name {
		return true
	}

	if auth.Kubernetes != nil && auth.Kubernetes.SecretRef.Name == name {
		return true
	}

	return false
}
