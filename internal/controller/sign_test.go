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
	"errors"
	"net/http"
	"testing"

	"github.com/cert-manager/issuer-lib/controllers/signer"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestClassifySigningError(t *testing.T) {
	tests := map[string]struct {
		err           error
		wantPermanent bool
	}{
		"403 from a PKI role that forbids the CSR is permanent": {
			err:           &vaultapi.ResponseError{StatusCode: http.StatusForbidden},
			wantPermanent: true,
		},
		"400 for a malformed CSR is permanent": {
			err:           &vaultapi.ResponseError{StatusCode: http.StatusBadRequest},
			wantPermanent: true,
		},
		"412 from replication lag is retried once the Vault client gives up": {
			err:           &vaultapi.ResponseError{StatusCode: http.StatusPreconditionFailed},
			wantPermanent: false,
		},
		"429 rate limit is retried": {
			err:           &vaultapi.ResponseError{StatusCode: http.StatusTooManyRequests},
			wantPermanent: false,
		},
		"503 from a sealed Vault is retried": {
			err:           &vaultapi.ResponseError{StatusCode: http.StatusServiceUnavailable},
			wantPermanent: false,
		},
		"a dropped connection is retried": {
			err:           errors.New("dial tcp: connection refused"),
			wantPermanent: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := classifySigningError(tc.err)

			var permanent signer.PermanentError
			assert.Equal(t, tc.wantPermanent, errors.As(got, &permanent))
			assert.ErrorIs(t, got, tc.err, "original error must stay in the chain")
		})
	}
}
