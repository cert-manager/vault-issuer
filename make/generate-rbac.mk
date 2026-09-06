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

.PHONY: generate-rbac
## Generate the controller's RBAC rules from the +kubebuilder:rbac markers.
##
## The result is written to the chart's files/ directory rather than
## templates/, because controller-gen emits a plain ClusterRole with a fixed
## name. templates/rbac.yaml reads the rules out of it and wraps them in a
## resource named after the release.
##
## @category Generate/ Verify
generate-rbac: | $(NEEDS_CONTROLLER-GEN)
	$(CONTROLLER-GEN) rbac:roleName=vault-issuer \
		paths=./internal/... \
		output:rbac:artifacts:config=$(helm_chart_source_dir)/files

shared_generate_targets += generate-rbac
