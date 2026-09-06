/*
Copyright 2020 The cert-manager Authors.

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

// Package vault contains helpers for driving the Vault server that
// `make e2e-setup-vault` installs into the test cluster.
//
// GenerateCA, generateVaultServingCert and generateVaultClientCert are copied
// verbatim from cert-manager's Vault addon (test/e2e/framework/addon/vault/vault.go).
// The rest of that file installed Vault with Helm at test time; here that is
// the job of the make targets, so it is not carried over.
package vault

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/cert-manager/cert-manager/pkg/util/pki"
)

func generateVaultServingCert(vaultCA []byte, vaultCAPrivateKey []byte, dnsName string) ([]byte, []byte) {
	catls, err := tls.X509KeyPair(vaultCA, vaultCAPrivateKey)
	if err != nil {
		panic(err)
	}

	ca, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		panic(err)
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   dnsName,
			Organization: []string{"cert-manager vault server"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{dnsName},
	}

	privateKey, err := pki.GenerateRSAPrivateKey(2048)
	if err != nil {
		panic(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &privateKey.PublicKey, catls.PrivateKey)
	if err != nil {
		panic(err)
	}

	encodedPrivateKey, err := pki.EncodePKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	return encodePublicKey(certBytes), encodedPrivateKey
}

func generateVaultClientCert(vaultCA []byte, vaultCAPrivateKey []byte) ([]byte, []byte) {
	catls, err := tls.X509KeyPair(vaultCA, vaultCAPrivateKey)
	if err != nil {
		panic(err)
	}

	ca, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		panic(err)
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   "cert-manager vault client",
			Organization: []string{"cert-manager"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	privateKey, err := pki.GenerateRSAPrivateKey(2048)
	if err != nil {
		panic(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &privateKey.PublicKey, catls.PrivateKey)
	if err != nil {
		panic(err)
	}

	encodedPrivateKey, err := pki.EncodePKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	return encodePublicKey(certBytes), encodedPrivateKey
}

func GenerateCA() ([]byte, []byte, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization: []string{"cert-manager test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	privateKey, err := pki.GenerateRSAPrivateKey(2048)
	if err != nil {
		return nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	encodedPrivateKey, err := pki.EncodePKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	return encodePublicKey(caBytes), encodedPrivateKey, nil
}

func encodePublicKey(pub []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: pub})
}
