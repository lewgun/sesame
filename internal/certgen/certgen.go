// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package certgen contains the code that handles the `certgen` subcommand
// for the main `sesame` binary.
package certgen

import (
	"context"
	"fmt"
	"path"

	"github.com/projectsesame/sesame/internal/dag"
	"github.com/projectsesame/sesame/pkg/certs"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// CACertificateKey is the dictionary key for the CA certificate bundle.
	CACertificateKey = "cacert.pem"
	// SesameCertificateKey is the dictionary key for the Sesame certificate.
	SesameCertificateKey = "sesamecert.pem"
	// SesamePrivateKeyKey is the dictionary key for the Sesame private key.
	SesamePrivateKeyKey = "sesamekey.pem"
	// EnvoyCertificateKey is the dictionary key for the Envoy certificate.
	EnvoyCertificateKey = "envoycert.pem"
	// EnvoyPrivateKeyKey is the dictionary key for the Envoy private key.
	EnvoyPrivateKeyKey = "envoykey.pem"
)

// OverwritePolicy specifies whether an output should be overwritten.
type OverwritePolicy int

const (
	// NoOverwrite specifies outputs must not be overwritten.
	NoOverwrite OverwritePolicy = 0
	// Overwrite specifies outputs may be overwritten.
	Overwrite OverwritePolicy = 1
)

func newSecret(secretType corev1.SecretType, name string, namespace string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		Type: secretType,
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "sesame",
			},
		},
		Data: data,
	}
}

// WritePEM writes a certificate out to its filename in outputDir.
func writePEM(outputDir, filename string, data []byte, force OverwritePolicy) error {
	filepath := path.Join(outputDir, filename)
	f, err := createFile(filepath, force == Overwrite)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return checkFile(filepath, err)
}

// WriteCertsPEM writes out all the certs in certdata to
// individual PEM files in outputDir
func WriteCertsPEM(outputDir string, certdata *certs.Certificates, force OverwritePolicy) error {
	err := writePEM(outputDir, "cacert.pem", certdata.CACertificate, force)
	if err != nil {
		return err
	}

	err = writePEM(outputDir, "Sesamecert.pem", certdata.SesameCertificate, force)
	if err != nil {
		return err
	}

	err = writePEM(outputDir, "Sesamekey.pem", certdata.SesamePrivateKey, force)
	if err != nil {
		return err
	}

	err = writePEM(outputDir, "envoycert.pem", certdata.EnvoyCertificate, force)
	if err != nil {
		return err
	}

	return writePEM(outputDir, "envoykey.pem", certdata.EnvoyPrivateKey, force)
}

// WriteSecretsYAML writes all the keypairs out to Kubernetes Secrets in YAML form
// in outputDir.
func WriteSecretsYAML(outputDir string, secrets []*corev1.Secret, force OverwritePolicy) error {
	for _, s := range secrets {
		filename := path.Join(outputDir, s.Name+".yaml")
		f, err := createFile(filename, force == Overwrite)
		if err != nil {
			return err
		}
		if err := checkFile(filename, writeSecret(f, s)); err != nil {
			return err
		}
	}

	return nil
}

// WriteSecretsKube writes all the keypairs out to Kubernetes Secrets in the
// compact format which is compatible with Secrets generated by cert-manager.
func WriteSecretsKube(client *kubernetes.Clientset, secrets []*corev1.Secret, force OverwritePolicy) error {
	for _, s := range secrets {
		if _, err := client.CoreV1().Secrets(s.Namespace).Create(context.TODO(), s, metav1.CreateOptions{}); err != nil {
			if k8serrors.IsAlreadyExists(err) && force == NoOverwrite {
				fmt.Printf("secret/%s already exists\n", s.Name)
				return nil
			}

			if _, err := client.CoreV1().Secrets(s.Namespace).Update(context.TODO(), s, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}

		fmt.Printf("secret/%s updated\n", s.Name)
	}

	return nil
}

// AsSecrets transforms the given Certificates struct into a slice of
// Secrets in in compact Secret format, which is compatible with
// both cert-manager and Sesame.
func AsSecrets(namespace string, certdata *certs.Certificates) []*corev1.Secret {
	return []*corev1.Secret{
		newSecret(corev1.SecretTypeTLS,
			"Sesamecert", namespace,
			map[string][]byte{
				dag.CACertificateKey:    certdata.CACertificate,
				corev1.TLSCertKey:       certdata.SesameCertificate,
				corev1.TLSPrivateKeyKey: certdata.SesamePrivateKey,
			}),
		newSecret(corev1.SecretTypeTLS,
			"envoycert", namespace,
			map[string][]byte{
				dag.CACertificateKey:    certdata.CACertificate,
				corev1.TLSCertKey:       certdata.EnvoyCertificate,
				corev1.TLSPrivateKeyKey: certdata.EnvoyPrivateKey,
			}),
	}
}

// AsLegacySecrets transforms the given Certificates struct into a slice of
// Secrets that is compatible with certgen from sesame 1.4 and earlier.
// The difference is that the CA cert is in a separate secret, rather
// than duplicated inline in each TLS secrets.
func AsLegacySecrets(namespace string, certdata *certs.Certificates) []*corev1.Secret {
	return []*corev1.Secret{
		newSecret(corev1.SecretTypeTLS,
			"Sesamecert", namespace,
			map[string][]byte{
				corev1.TLSCertKey:       certdata.SesameCertificate,
				corev1.TLSPrivateKeyKey: certdata.SesamePrivateKey,
			}),
		newSecret(corev1.SecretTypeTLS,
			"envoycert", namespace,
			map[string][]byte{
				corev1.TLSCertKey:       certdata.EnvoyCertificate,
				corev1.TLSPrivateKeyKey: certdata.EnvoyPrivateKey,
			}),
		newSecret(corev1.SecretTypeOpaque,
			"cacert", namespace,
			map[string][]byte{
				"cacert.pem": certdata.CACertificate,
			}),
	}
}
