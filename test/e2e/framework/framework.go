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

// The API deliberately mirrors cert-manager's e2e framework, so suites ported
// from that repository need only their import paths changed. What it does not
// mirror is the addon lifecycle: cert-manager, Vault and the vault-issuer
// controller are installed by the make targets before the suite starts.
package framework

import (
	"context"
	"fmt"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmclient "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cert-manager/vault-issuer/test/e2e/framework/addon"
	"github.com/cert-manager/vault-issuer/test/e2e/framework/config"
	"github.com/cert-manager/vault-issuer/test/e2e/framework/helper"
	"github.com/cert-manager/vault-issuer/test/e2e/framework/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Framework holds the per-spec state: a namespace created before the spec runs
// and deleted after it, plus clients scoped to the test cluster.
type Framework struct {
	BaseName string

	Config *config.Config

	// KubeClientConfig is the connection the clients below were built from.
	KubeClientConfig *rest.Config

	// Scheme encodes and decodes the Kubernetes objects the tests use.
	Scheme *runtime.Scheme

	KubeClientSet        kubernetes.Interface
	CertManagerClientSet cmclient.Interface

	// CRClient is a controller-runtime client, for the parts of a test that
	// prefer it over the typed clientsets.
	CRClient crclient.Client

	// Namespace is created for this spec and deleted when it finishes.
	Namespace *corev1.Namespace

	helper *helper.Helper
}

// NewDefaultFramework returns a Framework using the suite-wide configuration.
func NewDefaultFramework(baseName string) *Framework {
	return NewFramework(baseName, config.Default())
}

// NewFramework returns a Framework and registers the BeforeEach and AfterEach
// that create and delete its namespace.
func NewFramework(baseName string, cfg *config.Config) *Framework {
	f := &Framework{
		BaseName: baseName,
		Config:   cfg,
		helper:   helper.NewHelper(cfg),
	}

	BeforeEach(f.BeforeEach)
	AfterEach(f.AfterEach)

	return f
}

// NewScheme returns a scheme with every type the suites touch registered.
func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(certificatesv1.AddToScheme(scheme))
	utilruntime.Must(cmapi.AddToScheme(scheme))

	return scheme
}

// BeforeEach builds the clients and creates the spec's namespace.
func (f *Framework) BeforeEach(ctx context.Context) {
	details := addon.Base.Details()

	f.KubeClientConfig = details.KubeConfig
	f.KubeClientSet = details.KubeClient
	f.Scheme = NewScheme()

	var err error

	f.CertManagerClientSet, err = cmclient.NewForConfig(f.KubeClientConfig)
	Expect(err).NotTo(HaveOccurred(), "failed to build the cert-manager clientset")

	f.CRClient, err = crclient.New(f.KubeClientConfig, crclient.Options{Scheme: f.Scheme})
	Expect(err).NotTo(HaveOccurred(), "failed to build the controller-runtime client")

	f.helper.KubeClient = f.KubeClientSet
	f.helper.CMClient = f.CertManagerClientSet

	By("Creating a namespace for this test")
	ns, err := f.CreateKubeNamespace(ctx, f.BaseName)
	Expect(err).NotTo(HaveOccurred(), "failed to create the test namespace")

	f.Namespace = ns
}

// AfterEach deletes the spec's namespace.
func (f *Framework) AfterEach(ctx context.Context) {
	if f.Namespace == nil {
		return
	}

	By("Deleting the test namespace")
	err := f.DeleteKubeNamespace(ctx, f.Namespace.Name)
	Expect(err).NotTo(HaveOccurred(), "failed to delete the test namespace")

	f.Namespace = nil
}

// CreateKubeNamespace creates a namespace whose name starts with baseName.
func (f *Framework) CreateKubeNamespace(ctx context.Context, baseName string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("e2e-tests-%v-", baseName),
		},
	}

	created, err := f.KubeClientSet.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	log.Logf("Created namespace %v", created.Name)

	return created, nil
}

// DeleteKubeNamespace deletes the named namespace, tolerating one that has
// already gone away.
func (f *Framework) DeleteKubeNamespace(ctx context.Context, name string) error {
	err := f.KubeClientSet.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}

	return err
}

// Helper returns the wait and validation helpers, wired to this spec's clients.
func (f *Framework) Helper() *helper.Helper {
	return f.helper
}
