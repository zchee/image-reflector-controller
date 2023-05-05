package registry

import (
	"context"
	"fmt"
	"strings"

	imagev1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	"github.com/fluxcd/pkg/oci/auth/login"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Helper interface {
	GetAuthOptions(ctx context.Context, obj imagev1.ImageRepository) ([]remote.Option, error)
}

type DefaultHelper struct {
	k8sClient           client.Client
	DeprecatedLoginOpts login.ProviderOptions
}

var _ Helper = DefaultHelper{}

func NewDefaultHelper(c client.Client, deprecatedLoginOpts login.ProviderOptions) DefaultHelper {
	return DefaultHelper{
		k8sClient:           c,
		DeprecatedLoginOpts: deprecatedLoginOpts,
	}
}

// ParseImageReference parses the given URL into a container registry repository
// reference.
func ParseImageReference(url string) (name.Reference, error) {
	if s := strings.Split(url, "://"); len(s) > 1 {
		return nil, fmt.Errorf(".spec.image value should not start with URL scheme; remove '%s://'", s[0])
	}

	ref, err := name.ParseReference(url)
	if err != nil {
		return nil, err
	}

	imageName := strings.TrimPrefix(url, ref.Context().RegistryStr())
	if s := strings.Split(imageName, ":"); len(s) > 1 {
		return nil, fmt.Errorf(".spec.image value should not contain a tag; remove ':%s'", s[1])
	}

	return ref, nil
}
