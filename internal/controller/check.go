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

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/cert-manager/issuer-lib/api/v1alpha1"
	"github.com/cert-manager/issuer-lib/controllers/signer"
	"github.com/cert-manager/issuer-lib/intree"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	messageVaultClientInitFailed             = "Failed to initialize Vault client"
	messageVaultInitializedAndUnsealedFailed = "Failed to verify Vault is initialized and unsealed"
	messageVaultConfigRequired               = "Vault config cannot be empty"
	messageServerAndPathRequired             = "Vault server and path are required fields"
	messageAuthFieldsRequired                = "Vault tokenSecretRef, appRole, clientCertificate, kubernetes, or aws is required"
	messageMultipleAuthFieldsSet             = "Multiple auth methods cannot be set on the same Vault issuer"

	messageKubeAuthRoleRequired      = "Vault Kubernetes auth requires a role to be set"
	messageKubeAuthEitherRequired    = "Vault Kubernetes auth requires either secretRef.name or serviceAccountRef.name to be set"
	messageKubeAuthSingleRequired    = "Vault Kubernetes auth cannot be used with both secretRef.name and serviceAccountRef.name"
	messageTokenAuthNameRequired     = "Vault Token auth requires tokenSecretRef.name"
	messageAppRoleAuthFieldsRequired = "Vault AppRole auth requires both roleId and tokenSecretRef.name"
	messageAppRoleAuthKeyRequired    = "Vault AppRole auth requires secretRef.key"
)

// Setup creates a new Vault client and attempts to authenticate with the Vault instance and sets the issuer's conditions to reflect the success of the setup.
func (s *Signer) Check(ctx context.Context, issuerObject v1alpha1.Issuer) error {
	iss, ok := issuerObject.(intree.CMGenericIssuer)
	if !ok {
		return signer.PermanentError{
			Err: fmt.Errorf("issuer %T does not implement intree.CMGenericIssuer", issuerObject),
		}
	}

	// Configuration errors cannot be resolved without an edit to .spec, so they
	// are permanent: issuer-lib will not re-check until the generation changes.
	if err := validateVaultSpec(iss.IssuerSpec().Vault); err != nil {
		return signer.PermanentError{Err: err}
	}

	genericIssuer, ok := iss.Unwrap().(cmapi.GenericIssuer)
	if !ok {
		return signer.PermanentError{
			Err: fmt.Errorf("issuer %T does not implement cmapi.GenericIssuer", iss.Unwrap()),
		}
	}

	// Everything below depends on state outside the Issuer (Secrets, the Vault
	// server), so failures stay retryable: a plain error requeues with backoff.
	client, err := s.vaultClientBuilder(
		ctx,
		s.Options.ResourceNamespace(issuerObject),
		s.createTokenFn,
		s.Reader,
		genericIssuer,
		s.Options.CanUseAmbientCredentials(issuerObject),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", messageVaultClientInitFailed, err)
	}

	if err := client.IsVaultInitializedAndUnsealed(); err != nil {
		return fmt.Errorf("%s: %w", messageVaultInitializedAndUnsealedFailed, err)
	}

	log.FromContext(ctx).V(1).Info("Vault verified")

	return nil
}

func validateVaultSpec(vault *cmapi.VaultIssuer) error {
	if vault == nil {
		return errors.New(messageVaultConfigRequired)
	}

	// check if Vault server info is specified.
	if vault.Server == "" || vault.Path == "" {
		return errors.New(messageServerAndPathRequired)
	}

	tokenAuth := vault.Auth.TokenSecretRef
	appRoleAuth := vault.Auth.AppRole
	clientCertificateAuth := vault.Auth.ClientCertificate
	kubeAuth := vault.Auth.Kubernetes
	awsAuth := vault.Auth.AWS

	// check if at least one auth method is specified.
	if tokenAuth == nil && appRoleAuth == nil && clientCertificateAuth == nil && kubeAuth == nil && awsAuth == nil {
		return errors.New(messageAuthFieldsRequired)
	}

	// count how many auth methods are set
	authCount := 0
	if tokenAuth != nil {
		authCount++
	}
	if appRoleAuth != nil {
		authCount++
	}
	if clientCertificateAuth != nil {
		authCount++
	}
	if kubeAuth != nil {
		authCount++
	}
	if awsAuth != nil {
		authCount++
	}

	// check only one auth method is set
	if authCount > 1 {
		return errors.New(messageMultipleAuthFieldsSet)
	}

	// check if all mandatory Vault Token fields are set.
	if tokenAuth != nil && len(tokenAuth.Name) == 0 {
		return errors.New(messageTokenAuthNameRequired)
	}

	// check if all mandatory Vault appRole fields are set.
	if appRoleAuth != nil && (len(appRoleAuth.RoleId) == 0 || len(appRoleAuth.SecretRef.Name) == 0) {
		return errors.New(messageAppRoleAuthFieldsRequired)
	}
	if appRoleAuth != nil && len(appRoleAuth.SecretRef.Key) == 0 {
		return errors.New(messageAppRoleAuthKeyRequired)
	}

	// When using the Kubernetes auth, giving a role is mandatory.
	if kubeAuth != nil && len(kubeAuth.Role) == 0 {
		return errors.New(messageKubeAuthRoleRequired)
	}

	// When using the Kubernetes auth, you must either set secretRef or
	// serviceAccountRef.
	if kubeAuth != nil && (kubeAuth.SecretRef.Name == "" && kubeAuth.ServiceAccountRef == nil) {
		return errors.New(messageKubeAuthEitherRequired)
	}

	// When using the Kubernetes auth, you can't use secretRef and
	// serviceAccountRef simultaneously.
	if kubeAuth != nil && (kubeAuth.SecretRef.Name != "" && kubeAuth.ServiceAccountRef != nil) {
		return errors.New(messageKubeAuthSingleRequired)
	}

	return nil
}
