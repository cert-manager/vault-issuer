# Copyright 2023 The cert-manager Authors.
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

# `git describe` in the base Makefile falls back to a bare commit hash when no
# v* tag is reachable, which the helm module rejects. That happens in a clone
# made without tags, so give those checkouts a valid development version.
ifeq ($(shell git describe --tags --match='v*' --abbrev=0 2>/dev/null),)
VERSION := v0.0.0-dev.$(shell git rev-parse --short=14 HEAD)
endif

repo_name := github.com/cert-manager/vault-issuer

kind_cluster_name := vault-issuer
kind_cluster_config := $(bin_dir)/scratch/kind_cluster.yaml

build_names := controller

go_controller_main_dir := ./cmd/vault-issuer
go_controller_mod_dir := .
go_controller_ldflags := -X main.Version=$(VERSION)

oci_controller_base_image_flavor := static
oci_controller_image_tag := $(VERSION)
oci_controller_image_name_development := cert-manager.local/vault-issuer

deploy_name := vault-issuer
deploy_namespace := cert-manager

golangci_lint_config := .golangci.yaml

# The Helm chart. There are no CRDs to generate: this issuer serves
# cert-manager's own Issuer and ClusterIssuer resources.
helm_chart_source_dir := deploy/charts/vault-issuer
helm_chart_image_name := quay.io/jetstack/charts/vault-issuer
helm_chart_version := $(VERSION)
helm_dont_include_crds := true
helm_labels_template_name := vault-issuer.labels

# Point the chart's default image at the one that was just built, so that
# `make install` and the end-to-end tests deploy this working tree.
define helm_values_mutation_function
$(YQ) \
	--inplace \
	'.image.repository = "$(oci_controller_image_name_development)" | .image.tag = "$(oci_controller_image_tag)"' \
	$(1)
endef

# The Vault server used by the end-to-end tests. Preloaded into the kind cluster
# so that the Helm install works without pulling from the internet.
#
# These digests are the ones cert-manager pins for its own e2e tests; renovate
# keeps them current once the datasource comment below is in place.
# renovate: datasource=docker packageName=hashicorp/vault
vault_image_version := 1.14.1
images_amd64 += docker.io/hashicorp/vault:$(vault_image_version)@sha256:436d056e8e2a96c7356720069c29229970466f4f686886289dcc94dfa21d3155
images_arm64 += docker.io/hashicorp/vault:$(vault_image_version)@sha256:27dd264f3813c71a66792191db5382f0cf9eeaf1ae91770634911facfcfe4837

$(kind_cluster_config): make/config/kind/cluster.yaml | $(bin_dir)/scratch
	cat $< | \
	sed -e 's|{{KIND_IMAGES}}|$(CURDIR)/$(images_tar_dir)|g' \
	> $@
