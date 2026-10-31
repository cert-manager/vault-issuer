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

// Package base provides the clients that every other addon and suite builds on.
package base

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cert-manager/vault-issuer/test/e2e/framework/config"
)

// Details holds the cluster connection shared by the whole suite.
type Details struct {
	// Config is the suite-wide configuration.
	Config *config.Config

	// KubeConfig is the loaded Kubernetes client configuration.
	KubeConfig *rest.Config

	// KubeClient is a clientset built from KubeConfig.
	KubeClient kubernetes.Interface
}

// Base builds the cluster connection from the ambient kubeconfig.
type Base struct {
	details Details
}

// Details returns the cluster connection. It must not be called before Setup.
func (b *Base) Details() *Details {
	return &b.details
}

// Setup loads the kubeconfig and builds the clients.
func (b *Base) Setup() error {
	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return fmt.Errorf("loading kubeconfig, is KUBECONFIG set?: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("building the Kubernetes clientset: %w", err)
	}

	b.details = Details{
		Config:     config.Default(),
		KubeConfig: kubeConfig,
		KubeClient: kubeClient,
	}

	return nil
}
