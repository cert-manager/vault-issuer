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

include make/generate-rbac.mk
include make/test-unit.mk
include make/test-e2e.mk

.PHONY: release
release:
	$(MAKE) oci-push-controller
	$(MAKE) helm-chart-oci-push
	@echo "RELEASE_OCI_CONTROLLER_IMAGE=$(oci_controller_image_name)" >> "$(GITHUB_OUTPUT)"
	@echo "RELEASE_OCI_CONTROLLER_TAG=$(oci_controller_image_tag)" >> "$(GITHUB_OUTPUT)"
	@echo "RELEASE_HELM_CHART_IMAGE=$(helm_chart_image_name)" >> "$(GITHUB_OUTPUT)"
	@echo "RELEASE_HELM_CHART_VERSION=$(helm_chart_version)" >> "$(GITHUB_OUTPUT)"
	@echo "Release complete!"
