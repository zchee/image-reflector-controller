package registry

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/pkg/oci"
	"github.com/fluxcd/pkg/oci/auth/login"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	imagev1 "github.com/fluxcd/image-reflector-controller/api/v1beta3"
	"github.com/fluxcd/image-reflector-controller/internal/secret"
)

// GetAuthOptions returns authentication options required to scan a repository.
func (h DefaultHelper) GetAuthOptions(ctx context.Context, obj imagev1.ImageRepository) ([]remote.Option, error) {
	timeout := obj.GetTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Configure authentication strategy to access the registry.
	var options []remote.Option
	var authSecret corev1.Secret
	var auth authn.Authenticator
	var authErr error

	ref, err := ParseImageReference(obj.Spec.Image)
	if err != nil {
		return nil, fmt.Errorf("failed parsing image reference: %w", err)
	}

	if obj.Spec.SecretRef != nil {
		if err := h.k8sClient.Get(ctx, types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.Spec.SecretRef.Name,
		}, &authSecret); err != nil {
			return nil, err
		}
		auth, authErr = secret.AuthFromSecret(authSecret, ref)
	} else {
		// Build login provider options and use it to attempt registry login.
		opts := login.ProviderOptions{}
		switch obj.GetProvider() {
		case "aws":
			opts.AwsAutoLogin = true
		case "azure":
			opts.AzureAutoLogin = true
		case "gcp":
			opts.GcpAutoLogin = true
		default:
			opts = h.DeprecatedLoginOpts
		}
		auth, authErr = login.NewManager().Login(ctx, obj.Spec.Image, ref, opts)
	}
	if authErr != nil {
		// If it's not unconfigured provider error, abort reconciliation.
		// Continue reconciliation if it's unconfigured providers for scanning
		// public repositories.
		if !errors.Is(authErr, oci.ErrUnconfiguredProvider) {
			return nil, authErr
		}
	}
	if auth != nil {
		options = append(options, remote.WithAuth(auth))
	}

	// Load any provided certificate.
	if obj.Spec.CertSecretRef != nil {
		var certSecret corev1.Secret
		if obj.Spec.SecretRef != nil && obj.Spec.SecretRef.Name == obj.Spec.CertSecretRef.Name {
			certSecret = authSecret
		} else {
			if err := h.k8sClient.Get(ctx, types.NamespacedName{
				Namespace: obj.GetNamespace(),
				Name:      obj.Spec.CertSecretRef.Name,
			}, &certSecret); err != nil {
				return nil, err
			}
		}

		tr, err := secret.TransportFromKubeTLSSecret(&certSecret)
		if err != nil {
			return nil, err
		}
		if tr.TLSClientConfig == nil {
			tr, err = secret.TransportFromSecret(&certSecret)
			if err != nil {
				return nil, err
			}
			if tr.TLSClientConfig != nil {
				ctrl.LoggerFrom(ctx).
					Info("warning: specifying TLS auth data via `certFile`/`keyFile`/`caFile` is deprecated, please use `tls.crt`/`tls.key`/`ca.crt` instead")
			}
		}
		options = append(options, remote.WithTransport(tr))
	}

	if obj.Spec.ServiceAccountName != "" {
		serviceAccount := corev1.ServiceAccount{}
		// Lookup service account
		if err := h.k8sClient.Get(ctx, types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.Spec.ServiceAccountName,
		}, &serviceAccount); err != nil {
			return nil, err
		}

		if len(serviceAccount.ImagePullSecrets) > 0 {
			imagePullSecrets := make([]corev1.Secret, len(serviceAccount.ImagePullSecrets))
			for i, ips := range serviceAccount.ImagePullSecrets {
				var saAuthSecret corev1.Secret
				if err := h.k8sClient.Get(ctx, types.NamespacedName{
					Namespace: obj.GetNamespace(),
					Name:      ips.Name,
				}, &saAuthSecret); err != nil {
					return nil, err
				}
				imagePullSecrets[i] = saAuthSecret
			}
			keychain, err := k8schain.NewFromPullSecrets(ctx, imagePullSecrets)
			if err != nil {
				return nil, err
			}
			options = append(options, remote.WithAuthFromKeychain(keychain))
		}
	}

	return options, nil
}
