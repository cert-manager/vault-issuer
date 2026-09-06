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

package app

import (
	"fmt"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/cert-manager/vault-issuer/cmd/vault-issuer/app/options"
	"github.com/cert-manager/vault-issuer/internal/controller"
)

const (
	helpOutput = `vault-issuer is an issuer for HashiCorp Vault that implements the cert-manager Issuer API. 
It is a drop-in replacement for the in-tree Vault issuer controller, and can be used to migrate from the in-tree controller to an out-of-tree controller.`
)

func NewCommand() *cobra.Command {
	opts := options.New()

	cmd := &cobra.Command{
		Use:   "vault-issuer",
		Short: helpOutput,
		Long:  helpOutput,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(); err != nil {
				return err
			}

			log := opts.NewLogger()
			klog.SetLogger(log)
			ctrl.SetLogger(log)

			scheme := runtime.NewScheme()

			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cmapi.AddToScheme(scheme))

			mgr, err := ctrl.NewManager(opts.RestConfig, ctrl.Options{
				Scheme:                        scheme,
				LeaderElection:                opts.LeaderElectionConfig.Enabled,
				LeaderElectionID:              "vault-issuer-leader-election",
				LeaderElectionReleaseOnCancel: true,
				LeaseDuration:                 &opts.LeaderElectionConfig.LeaseDuration,
				RenewDeadline:                 &opts.LeaderElectionConfig.RenewDeadline,
				ReadinessEndpointName:         opts.ReadyzPath,
				HealthProbeBindAddress:        fmt.Sprintf("0.0.0.0:%d", opts.ReadyzPort),
				// We don't need any webhook since this looks at Cert-Manager Issuer
				// which has webhook validation built in.
				// WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
				// 	Port:    opts.Webhook.Port,
				// 	Host:    opts.Webhook.Host,
				// 	CertDir: opts.Webhook.CertDir,
				// 	TLSOpts: tlsOptions,
				// }),
				Metrics: server.Options{
					BindAddress: fmt.Sprintf("0.0.0.0:%d", opts.MetricsPort),
				},
				// Cache: bundle.CacheOpts(opts.Bundle),
			})
			if err != nil {
				return fmt.Errorf("failed to create controller manager: %w", err)
			}

			if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
				return fmt.Errorf("failed to add healthz check: %w", err)
			}

			if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
				return fmt.Errorf("failed to add readyz check: %w", err)
			}

			ctx := ctrl.SetupSignalHandler()
			logf.IntoContext(ctx, log)

			signer := controller.New(opts.IssuerOptions)
			if err := signer.SetupWithManager(ctx, mgr); err != nil {
				return fmt.Errorf("failed to register vault issuer controller: %w", err)
			}

			return mgr.Start(ctx)
		},
	}

	opts = opts.Prepare(cmd)

	return cmd
}
