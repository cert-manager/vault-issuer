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

// Package helper provides the wait and validation operations the suites need.
//
// It mirrors the method set of cert-manager's e2e helper, trimmed to what the
// Vault suites call. The kubectl-describe diagnostics that upstream emits on
// failure are not carried over; failures report the object's conditions
// instead.
package helper

import (
	cmclient "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"

	"github.com/cert-manager/vault-issuer/test/e2e/framework/config"
)

// Helper provides methods for common operations needed during tests.
type Helper struct {
	cfg *config.Config

	KubeClient kubernetes.Interface
	CMClient   cmclient.Interface
}

// NewHelper returns a Helper. The clients must be set by the caller before use;
// the framework does this in its BeforeEach.
func NewHelper(cfg *config.Config) *Helper {
	return &Helper{
		cfg: cfg,
	}
}
