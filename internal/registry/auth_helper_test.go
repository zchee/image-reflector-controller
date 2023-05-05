/*
Copyright 2023 The Flux authors

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

package registry_test

import (
	"context"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/oci/auth/login"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	imagev1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	"github.com/fluxcd/image-reflector-controller/internal/registry"
	"github.com/fluxcd/image-reflector-controller/internal/secret"
	"github.com/fluxcd/image-reflector-controller/internal/test"
)

func TestDefaultHelperAuthOptions(t *testing.T) {
	testImg := "example.com/foo/bar"

	testSecretName := "test-secret"
	testTLSSecretName := "test-tls-secret"
	testServiceAccountName := "test-service-account"
	testNamespace := "test-ns"

	dockerconfigjson := []byte(`
{
	"auths": {
		"example.com": {
			"username": "user",
			"password": "pass"
		}
	}
}`)

	testSecret := &corev1.Secret{}
	testSecret.Name = testSecretName
	testSecret.Namespace = testNamespace
	testSecret.Type = corev1.SecretTypeDockerConfigJson
	testSecret.Data = map[string][]byte{".dockerconfigjson": dockerconfigjson}
	g := NewWithT(t)

	// Create a test TLS server to get valid cert data. The server is never
	// started or used below.
	_, rootCertPEM, clientCertPEM, clientKeyPEM, _, err := test.CreateTLSServer()
	g.Expect(err).To(Not(HaveOccurred()))

	testTLSSecret := &corev1.Secret{}
	testTLSSecret.Name = testTLSSecretName
	testTLSSecret.Namespace = testNamespace
	testTLSSecret.Type = corev1.SecretTypeTLS
	testTLSSecret.Data = map[string][]byte{
		secret.CACert:     rootCertPEM,
		secret.ClientCert: clientCertPEM,
		secret.ClientKey:  clientKeyPEM,
	}

	// Secret with docker config and TLS secrets.
	testSecretWithTLS := testSecret.DeepCopy()
	testSecretWithTLS.Data = map[string][]byte{
		".dockerconfigjson": dockerconfigjson,
		secret.CACert:       rootCertPEM,
		secret.ClientCert:   clientCertPEM,
		secret.ClientKey:    clientKeyPEM,
	}

	// ServiceAccount without image pull secret.
	testServiceAccount := &corev1.ServiceAccount{}
	testServiceAccount.Name = testServiceAccountName
	testServiceAccount.Namespace = testNamespace

	// ServiceAccount with image pull secret.
	testServiceAccountWithSecret := testServiceAccount.DeepCopy()
	testServiceAccountWithSecret.ImagePullSecrets = []corev1.LocalObjectReference{{Name: testSecretName}}

	tests := []struct {
		name       string
		repo       imagev1.ImageRepository
		k8sObjs    []client.Object
		expectErr  bool
		expectOpts int
	}{
		{
			name:       "fails with empty image reference",
			repo:       imagev1.ImageRepository{},
			expectErr:  true,
			expectOpts: 0,
		},
		{
			name: "succeeds with no auth options",
			repo: imagev1.ImageRepository{
				Spec: imagev1.ImageRepositorySpec{
					Image: testImg,
				},
			},
			expectErr:  false,
			expectOpts: 0,
		},
		{
			name: "succeeds with ECR image ref",
			repo: imagev1.ImageRepository{
				Spec: imagev1.ImageRepositorySpec{
					Image: "123456789000.dkr.ecr.us-east-2.amazonaws.com/test",
				},
			},
			expectErr:  false,
			expectOpts: 0,
		},
		{
			name: "fails with contextual login but no auth credentials",
			repo: imagev1.ImageRepository{
				Spec: imagev1.ImageRepositorySpec{
					Image:    "123456789000.dkr.ecr.us-east-2.amazonaws.com/test",
					Provider: "aws",
				},
			},
			expectErr:  true,
			expectOpts: 0,
		},
		{
			name: "fails with missing secret",
			repo: imagev1.ImageRepository{
				Spec: imagev1.ImageRepositorySpec{
					SecretRef: &meta.LocalObjectReference{
						Name: "does-not-exist",
					},
				},
			},
			expectErr:  true,
			expectOpts: 0,
		},
		{
			name: "fails with wrong secret type",
			repo: imagev1.ImageRepository{
				Spec: imagev1.ImageRepositorySpec{
					SecretRef: &meta.LocalObjectReference{
						Name: "registry-auth",
					},
				},
			},
			k8sObjs: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "registry-auth",
					},
					Type: corev1.SecretTypeOpaque,
				},
			},
			expectErr:  true,
			expectOpts: 0,
		},
		{
			name: "fails with empty secret data",
			repo: imagev1.ImageRepository{
				Spec: imagev1.ImageRepositorySpec{
					SecretRef: &meta.LocalObjectReference{
						Name: "registry-auth",
					},
				},
			},
			k8sObjs: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "registry-auth",
					},
					Type: corev1.SecretTypeDockerConfigJson,
				},
			},
			expectErr:  true,
			expectOpts: 0,
		},
		{
			name: "adds authenticator from secret",
			repo: imagev1.ImageRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace},
				Spec: imagev1.ImageRepositorySpec{
					Image: testImg,
					SecretRef: &meta.LocalObjectReference{
						Name: testSecretName,
					},
				},
			},
			k8sObjs:    []client.Object{testSecret},
			expectErr:  false,
			expectOpts: 1,
		},
		{
			name: "fails with non-existing cert secret ref",
			repo: imagev1.ImageRepository{
				Spec: imagev1.ImageRepositorySpec{
					Image: testImg,
					CertSecretRef: &meta.LocalObjectReference{
						Name: "non-existing-secret",
					},
				},
			},
			expectErr:  true,
			expectOpts: 0,
		},
		{
			name: "sets transport from cert secret ref",
			repo: imagev1.ImageRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace},
				Spec: imagev1.ImageRepositorySpec{
					Image: testImg,
					CertSecretRef: &meta.LocalObjectReference{
						Name: testTLSSecretName,
					},
				},
			},
			k8sObjs:    []client.Object{testTLSSecret},
			expectErr:  false,
			expectOpts: 1,
		},
		{
			name: "sets transport and auth from secret ref and cert secret ref",
			repo: imagev1.ImageRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace},
				Spec: imagev1.ImageRepositorySpec{
					Image: testImg,
					SecretRef: &meta.LocalObjectReference{
						Name: testSecretName,
					},
					CertSecretRef: &meta.LocalObjectReference{
						Name: testTLSSecretName,
					},
				},
			},
			k8sObjs:    []client.Object{testSecret, testTLSSecret},
			expectErr:  false,
			expectOpts: 2,
		},
		{
			name: "sets transport and auth from same secret and cert ref",
			repo: imagev1.ImageRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace},
				Spec: imagev1.ImageRepositorySpec{
					Image: testImg,
					SecretRef: &meta.LocalObjectReference{
						Name: testSecretName,
					},
					CertSecretRef: &meta.LocalObjectReference{
						Name: testSecretName,
					},
				},
			},
			k8sObjs:    []client.Object{testSecretWithTLS},
			expectErr:  false,
			expectOpts: 2,
		},
		{
			name: "sets auth option from SA with pull secret",
			repo: imagev1.ImageRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace},
				Spec: imagev1.ImageRepositorySpec{
					Image:              testImg,
					ServiceAccountName: testServiceAccountName,
				},
			},
			k8sObjs:    []client.Object{testSecret, testServiceAccountWithSecret},
			expectErr:  false,
			expectOpts: 1,
		},
		{
			name: "fails with SA an non-existing pull secret",
			repo: imagev1.ImageRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace},
				Spec: imagev1.ImageRepositorySpec{
					Image:              testImg,
					ServiceAccountName: testServiceAccountName,
				},
			},
			k8sObjs:    []client.Object{testServiceAccountWithSecret},
			expectErr:  true,
			expectOpts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			k8sClient := fake.NewClientBuilder().
				WithObjects(tt.k8sObjs...).
				Build()
			h := registry.NewDefaultHelper(k8sClient, login.ProviderOptions{})

			opts, err := h.GetAuthOptions(context.Background(), tt.repo)
			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
			g.Expect(opts).To(HaveLen(tt.expectOpts))
		})
	}
}

func TestParseImageReference(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		wantRef string
	}{
		{
			name:    "simple valid url",
			url:     "example.com/foo/bar",
			wantRef: "example.com/foo/bar",
		},
		{
			name:    "with scheme prefix",
			url:     "https://example.com/foo/bar",
			wantErr: true,
		},
		{
			name:    "with tag",
			url:     "example.com/foo/bar:baz",
			wantErr: true,
		},
		{
			name:    "with host port",
			url:     "example.com:9999/foo/bar",
			wantErr: false,
			wantRef: "example.com:9999/foo/bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			ref, err := registry.ParseImageReference(tt.url)
			g.Expect(err != nil).To(Equal(tt.wantErr))
			if err == nil {
				g.Expect(ref.String()).To(Equal(tt.wantRef))
			}
		})
	}
}
