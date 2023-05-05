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

// AuthHelper facilitates interacting with an OCI registry, in particular authentication against it.
type AuthHelper interface {
	// GetAuthOptions will generate a list of options from the given ImageRepositry. These can then
	// be used to interact with a registry.
	GetAuthOptions(ctx context.Context, obj imagev1.ImageRepository) ([]remote.Option, error)
}

// DefaultAuthHelper implements AuthHelper and delegates calls to the Kubernets API to the given Client.
type DefaultAuthHelper struct {
	k8sClient           client.Client
	DeprecatedLoginOpts login.ProviderOptions
}

var _ AuthHelper = DefaultAuthHelper{}

// NewDefaultHelper creates a new DefaultAuthHelper and passes in the given client and provider options.
func NewDefaultHelper(c client.Client, deprecatedLoginOpts login.ProviderOptions) DefaultAuthHelper {
	return DefaultAuthHelper{
		k8sClient:           c,
		DeprecatedLoginOpts: deprecatedLoginOpts,
	}
}

// ParseImageReference parses the given image name into a container registry repository
// reference.
func ParseImageReference(image string) (name.Reference, error) {
	if s := strings.Split(image, "://"); len(s) > 1 {
		return nil, fmt.Errorf("the image name must not include a URL scheme; remove '%s://'", s[0])
	}

	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, err
	}

	imageName := strings.TrimPrefix(image, ref.Context().RegistryStr())
	if s := strings.Split(imageName, ":"); len(s) > 1 {
		return nil, fmt.Errorf("the image name must not contain a tag; remove ':%s'", s[1])
	}

	return ref, nil
}
