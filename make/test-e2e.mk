# Copyright 2026 The cert-manager Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

vault_helm_chart_repo := https://helm.releases.hashicorp.com
vault_helm_chart_version := 0.25.0

# Two Vault servers are installed: a default one, and one that requires client
# certificates. cert-manager runs the same pair, as its e2e-vault and
# e2e-vault-mtls fixtures.
# The release names must differ: the chart creates a cluster-scoped
# ClusterRoleBinding named <release>-server-binding, which two releases sharing
# a name cannot both own, whatever namespaces they live in.
vault_release_name := vault
vault_namespace := vault
vault_mtls_release_name := vault-mtls
vault_mtls_namespace := vault-mtls

vault_tls_dir := $(bin_dir)/scratch/vault-tls
vault_mtls_tls_dir := $(bin_dir)/scratch/vault-mtls-tls

# The namespace the controller is deployed into. The chart defaults
# --cluster-resource-namespace to the release namespace, so this is also where
# ClusterIssuer Secrets are read from.
cluster_resource_namespace := cert-manager

# The environment the end-to-end tests read. This is the entire contract between
# the make targets, which install and configure everything, and the Go tests,
# which only consume it. See test/e2e/framework/addon/vault/details.go.
export E2E_CLUSTER_RESOURCE_NAMESPACE := $(cluster_resource_namespace)

export E2E_VAULT_URL := https://$(vault_release_name).$(vault_namespace).svc.cluster.local:8200
export E2E_VAULT_NAMESPACE := $(vault_namespace)
export E2E_VAULT_POD := $(vault_release_name)-0
export E2E_VAULT_CA_FILE := $(CURDIR)/$(vault_tls_dir)/ca.crt
export E2E_VAULT_CLIENT_CERT_FILE := $(CURDIR)/$(vault_tls_dir)/client.crt
export E2E_VAULT_CLIENT_KEY_FILE := $(CURDIR)/$(vault_tls_dir)/client.key
export E2E_VAULT_ENFORCE_MTLS := false

export MTLS_E2E_VAULT_URL := https://$(vault_mtls_release_name).$(vault_mtls_namespace).svc.cluster.local:8200
export MTLS_E2E_VAULT_NAMESPACE := $(vault_mtls_namespace)
export MTLS_E2E_VAULT_POD := $(vault_mtls_release_name)-0
export MTLS_E2E_VAULT_CA_FILE := $(CURDIR)/$(vault_mtls_tls_dir)/ca.crt
export MTLS_E2E_VAULT_CLIENT_CERT_FILE := $(CURDIR)/$(vault_mtls_tls_dir)/client.crt
export MTLS_E2E_VAULT_CLIENT_KEY_FILE := $(CURDIR)/$(vault_mtls_tls_dir)/client.key
export MTLS_E2E_VAULT_ENFORCE_MTLS := true

.PHONY: e2e-setup-cert-manager
## Install cert-manager into the kind cluster, using the preloaded images.
## @category Testing
e2e-setup-cert-manager: | kind-cluster $(NEEDS_HELM) $(NEEDS_KUBECTL)
	$(HELM) upgrade \
		--install \
		--create-namespace \
		--wait \
		--version $(quay.io/jetstack/cert-manager-controller.TAG) \
		--namespace cert-manager \
		--repo https://charts.jetstack.io \
		--set crds.enabled=true \
		--set image.repository=$(quay.io/jetstack/cert-manager-controller.REPO) \
		--set image.tag=$(quay.io/jetstack/cert-manager-controller.TAG) \
		--set image.pullPolicy=Never \
		--set cainjector.image.repository=$(quay.io/jetstack/cert-manager-cainjector.REPO) \
		--set cainjector.image.tag=$(quay.io/jetstack/cert-manager-cainjector.TAG) \
		--set cainjector.image.pullPolicy=Never \
		--set webhook.image.repository=$(quay.io/jetstack/cert-manager-webhook.REPO) \
		--set webhook.image.tag=$(quay.io/jetstack/cert-manager-webhook.TAG) \
		--set webhook.image.pullPolicy=Never \
		--set startupapicheck.image.repository=$(quay.io/jetstack/cert-manager-startupapicheck.REPO) \
		--set startupapicheck.image.tag=$(quay.io/jetstack/cert-manager-startupapicheck.TAG) \
		--set startupapicheck.image.pullPolicy=Never \
		--set "extraArgs={--controllers=*\,-certificaterequests-issuer-vault\,-issuers\,-clusterissuers}" \
		cert-manager cert-manager >/dev/null

# Generate the CA, serving certificate and client certificate for a Vault
# installation. $1 is the output directory, $2 the DNS name of the server.
define generate-vault-tls
	cd $(CURDIR)/test/e2e && GOWORK=off $(GO) run ./hack/gen-vault-certs \
		--out-dir $(CURDIR)/$(1) \
		--dns-name $(2)
endef

# Install a Vault server. $1 is the release name, $2 the namespace, $3 the TLS
# directory, $4 whether the listener requires a client certificate.
define install-vault
	$(KUBECTL) create namespace $(2) --dry-run=client -o yaml | $(KUBECTL) apply -f -

	$(KUBECTL) -n $(2) create secret generic vault-tls \
		--from-file=ca.crt=$(CURDIR)/$(3)/ca.crt \
		--from-file=server.crt=$(CURDIR)/$(3)/server.crt \
		--from-file=server.key=$(CURDIR)/$(3)/server.key \
		--from-file=client.crt=$(CURDIR)/$(3)/client.crt \
		--from-file=client.key=$(CURDIR)/$(3)/client.key \
		--dry-run=client -o yaml | $(KUBECTL) apply -f -

	sed -e 's|{{TLS_REQUIRE_CLIENT_CERT}}|$(4)|g' \
		make/config/vault/values.yaml \
		> $(bin_dir)/scratch/vault-values-$(1).yaml

	# The StatefulSet below uses the OnDelete update strategy, so an updated
	# vault-tls Secret never reaches a running pod, and Vault would keep
	# serving a certificate signed by a CA the tests no longer trust. Restart
	# the pod when the CA it started with is not the one on disk any more.
	if $(KUBECTL) -n $(2) get pod $(1)-0 >/dev/null 2>&1 && \
		! $(KUBECTL) -n $(2) exec $(1)-0 -- cat /vault/tls/ca.crt 2>/dev/null \
			| diff -q - $(CURDIR)/$(3)/ca.crt >/dev/null 2>&1; then \
		$(KUBECTL) -n $(2) delete pod $(1)-0 --wait=false; \
	fi

	# The chart creates the ServiceAccount and the system:auth-delegator
	# binding that Vault's Kubernetes auth method needs to call TokenReview.
	$(HELM) upgrade \
		--install \
		--wait \
		--namespace $(2) \
		--repo $(vault_helm_chart_repo) \
		--version $(vault_helm_chart_version) \
		--values $(bin_dir)/scratch/vault-values-$(1).yaml \
		--set server.image.repository=$(docker.io/hashicorp/vault.REPO) \
		--set server.image.tag=$(docker.io/hashicorp/vault.TAG) \
		$(1) vault >/dev/null

	# The chart's StatefulSet uses the OnDelete update strategy, which
	# `kubectl rollout status` refuses to report on, so wait on the pod instead.
	$(KUBECTL) -n $(2) wait --for=condition=Ready pod \
		--selector app.kubernetes.io/name=vault,app.kubernetes.io/instance=$(1),component=server \
		--timeout=5m
endef

# The TLS files are deliberately not regenerated on every run: `make test-e2e`
# depends on `e2e-setup`, so a .PHONY generate step would mint a new CA on every
# invocation and invalidate the certificate the running Vault server is serving.
$(vault_tls_dir)/ca.crt: | $(NEEDS_GO) $(bin_dir)/scratch
	$(call generate-vault-tls,$(vault_tls_dir),$(vault_release_name).$(vault_namespace).svc.cluster.local)

$(vault_mtls_tls_dir)/ca.crt: | $(NEEDS_GO) $(bin_dir)/scratch
	$(call generate-vault-tls,$(vault_mtls_tls_dir),$(vault_mtls_release_name).$(vault_mtls_namespace).svc.cluster.local)

.PHONY: e2e-setup-vault
## Install a Vault server into the kind cluster, serving TLS.
## @category Testing
e2e-setup-vault: $(vault_tls_dir)/ca.crt | kind-cluster $(NEEDS_HELM) $(NEEDS_KUBECTL) $(bin_dir)/scratch
	$(call install-vault,$(vault_release_name),$(vault_namespace),$(vault_tls_dir),false)

.PHONY: e2e-setup-vault-mtls
## Install a second Vault server that requires client certificates.
## @category Testing
e2e-setup-vault-mtls: $(vault_mtls_tls_dir)/ca.crt | kind-cluster $(NEEDS_HELM) $(NEEDS_KUBECTL) $(bin_dir)/scratch
	$(call install-vault,$(vault_mtls_release_name),$(vault_mtls_namespace),$(vault_mtls_tls_dir),true)

.PHONY: e2e-setup-vault-issuer
## Build, load and deploy the vault-issuer controller into the kind cluster.
## @category Testing
e2e-setup-vault-issuer: oci-load-controller $(helm_chart_archive) | kind-cluster $(NEEDS_HELM) $(NEEDS_KUBECTL)
	# helm_values_mutation_function has already pointed the chart's default
	# image at the one oci-load-controller just loaded into the cluster.
	# On failure, print what the pod did: the Helm error only ever says
	# "Available: 0/1", which is true of a crash loop, a missing image and a
	# failing probe alike.
	$(HELM) upgrade \
		--install \
		--wait \
		--namespace $(cluster_resource_namespace) \
		--set image.pullPolicy=Never \
		$(deploy_name) $(helm_chart_archive) >/dev/null \
	|| ( \
		$(KUBECTL) -n $(cluster_resource_namespace) get pods -l app.kubernetes.io/name=vault-issuer -o wide; \
		$(KUBECTL) -n $(cluster_resource_namespace) describe deployment $(deploy_name); \
		$(KUBECTL) -n $(cluster_resource_namespace) logs -l app.kubernetes.io/name=vault-issuer \
			--all-containers --prefix --tail=200 --ignore-errors; \
		exit 1 \
	)

	$(KUBECTL) -n $(cluster_resource_namespace) rollout status deployment $(deploy_name) --timeout=5m

.PHONY: e2e-setup
## Install everything the end-to-end tests need into the kind cluster.
## @category Testing
e2e-setup: e2e-setup-cert-manager e2e-setup-vault e2e-setup-vault-mtls e2e-setup-vault-issuer

test-e2e-deps: e2e-setup

.PHONY: test-e2e
## Run the end-to-end tests, creating the kind cluster if it does not exist.
##
## To run a subset of the specs:
##
##	make test-e2e E2E_FOCUS='Vault Issuer'
##
## @category Testing
test-e2e: test-e2e-deps | kind-cluster $(NEEDS_GO) $(NEEDS_KUBECTL) $(ARTIFACTS)
	$(eval abs_artifacts := $(abspath $(ARTIFACTS)))
	$(eval E2E_FOCUS ?= )

	# Use `go run` to invoke ginkgo directly from test/e2e/go.mod, so it can
	# generate its own junit output with individual spec entries (not wrapped).
	cd ./test/e2e && \
	GOWORK=off \
	KUBECONFIG=$(absolute_kubeconfig) \
	E2E_CLUSTER_RESOURCE_NAMESPACE='$(E2E_CLUSTER_RESOURCE_NAMESPACE)' \
	E2E_VAULT_URL='$(E2E_VAULT_URL)' \
	E2E_VAULT_NAMESPACE='$(E2E_VAULT_NAMESPACE)' \
	E2E_VAULT_POD='$(E2E_VAULT_POD)' \
	E2E_VAULT_CA_FILE='$(E2E_VAULT_CA_FILE)' \
	E2E_VAULT_CLIENT_CERT_FILE='$(E2E_VAULT_CLIENT_CERT_FILE)' \
	E2E_VAULT_CLIENT_KEY_FILE='$(E2E_VAULT_CLIENT_KEY_FILE)' \
	E2E_VAULT_ENFORCE_MTLS='$(E2E_VAULT_ENFORCE_MTLS)' \
	MTLS_E2E_VAULT_URL='$(MTLS_E2E_VAULT_URL)' \
	MTLS_E2E_VAULT_NAMESPACE='$(MTLS_E2E_VAULT_NAMESPACE)' \
	MTLS_E2E_VAULT_POD='$(MTLS_E2E_VAULT_POD)' \
	MTLS_E2E_VAULT_CA_FILE='$(MTLS_E2E_VAULT_CA_FILE)' \
	MTLS_E2E_VAULT_CLIENT_CERT_FILE='$(MTLS_E2E_VAULT_CLIENT_CERT_FILE)' \
	MTLS_E2E_VAULT_CLIENT_KEY_FILE='$(MTLS_E2E_VAULT_CLIENT_KEY_FILE)' \
	MTLS_E2E_VAULT_ENFORCE_MTLS='$(MTLS_E2E_VAULT_ENFORCE_MTLS)' \
	$(GO) run github.com/onsi/ginkgo/v2/ginkgo \
		--timeout=25m \
		--junit-report=$(abs_artifacts)/junit-go-e2e.xml \
		$(if $(E2E_FOCUS),--focus='$(E2E_FOCUS)',) \
		. \
	|| ( \
		$(MAKE) --no-print-directory e2e-diagnostics; \
		exit 1 \
	)

.PHONY: e2e-diagnostics
## Print the controller logs and the issuance state to stdout, so that a failed
## end-to-end run can be triaged from the CI build log alone.
## @category Testing
e2e-diagnostics: | kind-cluster $(NEEDS_KUBECTL)
	@echo "======== vault-issuer logs ========"
	@$(KUBECTL) -n $(cluster_resource_namespace) logs \
		-l app.kubernetes.io/name=vault-issuer \
		--all-containers --prefix --tail=-1 --ignore-errors || true

	@echo "======== CertificateRequests ========"
	@$(KUBECTL) get certificaterequests --all-namespaces -o yaml || true

	@echo "======== Certificates ========"
	@$(KUBECTL) get certificates --all-namespaces -o yaml || true

	@echo "======== Issuers / ClusterIssuers ========"
	@$(KUBECTL) get issuers --all-namespaces -o yaml || true
	@$(KUBECTL) get clusterissuers -o yaml || true

	@echo "======== Pods ========"
	@$(KUBECTL) get pods --all-namespaces -o wide || true

.PHONY: e2e-logs
## Export the kind cluster's logs into $(ARTIFACTS), for triaging CI failures.
## @category Testing
e2e-logs: | kind-cluster $(NEEDS_KIND) $(ARTIFACTS)
	rm -rf $(ARTIFACTS)/e2e-logs
	mkdir -p $(ARTIFACTS)/e2e-logs
	$(KIND) export logs $(ARTIFACTS)/e2e-logs --name=$(kind_cluster_name)
