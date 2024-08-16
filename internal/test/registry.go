/*
Copyright 2022 The Flux authors

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

package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/fluxcd/image-reflector-controller/internal/policy"
)

// pre-populated db of tags, so it's not necessary to upload images to
// get results from remote.List.
var convenientTags = map[string][]string{
	"convenient": {
		"tag1", "tag2",
	},
}

// sets up a local registry for testing scanning.
func NewRegistryServer() *httptest.Server {
	logOpt := registry.Logger(log.New(io.Discard, "", log.LstdFlags))
	h := registry.New(logOpt, registry.WithReferrersSupport(true))
	regHandler := &TagListHandler{
		RegistryHandler: h,
		Imagetags:       convenientTags,
	}
	srv := httptest.NewServer(regHandler)
	return srv
}

func NewAuthenticatedRegistryServer(username, pass string) *httptest.Server {
	logOpt := registry.Logger(log.New(io.Discard, "", log.LstdFlags))
	tagListHandler := &TagListHandler{
		RegistryHandler: registry.New(logOpt, registry.WithReferrersSupport(true)),
		Imagetags:       convenientTags,
	}
	regHandler := &AuthHandler{
		registryHandler: tagListHandler,
		allowedUser:     username,
		allowedPass:     pass,
	}
	srv := httptest.NewServer(regHandler)
	return srv
}

// RegistryName gets registry part of an image from the registry server.
func RegistryName(srv *httptest.Server) string {
	if strings.HasPrefix(srv.URL, "https://") {
		return strings.TrimPrefix(srv.URL, "https://")
	}

	// assume HTTP protocol
	return strings.TrimPrefix(srv.URL, "http://")
}

// LoadImages uploads images to the local registry, and returns the
// image repo name.
// [github.com/google/go-containerregistry@v0.18.0/pkg/registry/compatibility_test.go](https://github.com/google/go-containerregistry/blob/v0.18.0/pkg/registry/compatibility_test.go)
// has an example of loading a test registry with a random image.
func LoadImages(srv *httptest.Server, imageName string, tags []policy.Tag, options ...remote.Option) (string, error) {
	imgRepo := RegistryName(srv) + "/" + imageName

	for _, tag := range tags {
		tagRef, err := name.NewTag(imgRepo + ":" + tag.Name)
		if err != nil {
			return imgRepo, fmt.Errorf("unable to create tag: %w", err)
		}
		img, err := random.Image(512, 1)
		if err != nil {
			return imgRepo, fmt.Errorf("unable to make random image: %w", err)
		}
		img, err = mutate.ConfigFile(img, &v1.ConfigFile{Created: v1.Time{Time: tag.Created}})
		if err != nil {
			return imgRepo, fmt.Errorf("unable to add ConfigFile to random image: %w", err)
		}
		if err := remote.Write(tagRef, img, options...); err != nil {
			return imgRepo, fmt.Errorf("error writing image: %w", err)
		}
	}

	return imgRepo, nil
}

// the go-containerregistry test registry implementation does not
// serve /myimage/tags/list. Until it does, I'm adding this handler.
// NB:
// - assumes repo name is a single element
// - assumes no overwriting tags

type TagListHandler struct {
	RegistryHandler http.Handler
	Imagetags       map[string][]string
}

type TagListResult struct {
	Name    string    `json:"name"`
	Tags    []string  `json:"tags"`
	Created time.Time `json:"createdAt,omitempty"`
}

// ServeHTTP implements [http.Handler].
func (h *TagListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// a tag list request has a path like: /v2/<repo>/tags/list
	if withoutTagsList := strings.TrimSuffix(r.URL.Path, "/tags/list"); r.Method == http.MethodGet && withoutTagsList != r.URL.Path {
		repo := strings.TrimPrefix(withoutTagsList, "/v2/")

		if tags, ok := h.Imagetags[repo]; ok {
			w.Header().Set("Content-Type", `application/json`)
			w.WriteHeader(http.StatusOK)
			result := TagListResult{
				Name: repo,
				Tags: tags,
			}
			enc := json.NewEncoder(w)
			if err := enc.Encode(&result); err != nil {
				http.Error(w, fmt.Errorf("unable to encode: %w", err).Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
		return
	}

	// record the fact of a PUT to a tag; the path looks like: /v2/<repo>/manifests/<tag>
	h.RegistryHandler.ServeHTTP(w, r)

	if r.Method == http.MethodPut {
		pathElements := strings.Split(r.URL.Path, "/")
		if len(pathElements) == 5 && pathElements[1] == "v2" && pathElements[3] == "manifests" {
			repo, tag := pathElements[2], pathElements[4]
			h.Imagetags[repo] = append(h.Imagetags[repo], tag)

			_, err := name.ParseReference(repo+":"+tag, name.WithDefaultRegistry(""))
			if err != nil {
				err = fmt.Errorf("error parsing tag: %w", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

// there's no authentication in go-containerregistry/pkg/registry;
// this wrapper adds basic auth to a registry handler. NB: the
// important thing is to be able to test that the credentials get from
// the secret to the registry API library; it's assumed that the
// registry API library does e.g., OAuth2 correctly.  See
// https://tools.ietf.org/html/rfc7617 regarding basic authentication.

type AuthHandler struct {
	allowedUser, allowedPass string
	registryHandler          http.Handler
}

// ServeHTTP serves a request which needs authentication.
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Add("WWW-Authenticate", `Basic realm="Registry"`)
		w.WriteHeader(401)
		return
	}
	if !strings.HasPrefix(authHeader, "Basic ") {
		w.WriteHeader(403)
		w.Write([]byte(`Authorization header does not being with "Basic "`))
		return
	}
	namePass, err := base64.StdEncoding.DecodeString(authHeader[6:])
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte(`Authorization header doesn't appear to be base64-encoded`))
		return
	}
	namePassSlice := strings.SplitN(string(namePass), ":", 2)
	if len(namePassSlice) != 2 {
		w.WriteHeader(403)
		w.Write([]byte(`Authorization header doesn't appear to be colon-separated value `))
		w.Write(namePass)
		return
	}
	if namePassSlice[0] != h.allowedUser || namePassSlice[1] != h.allowedPass {
		w.WriteHeader(403)
		w.Write([]byte(`Authorization failed: wrong username or password`))
		return
	}
	h.registryHandler.ServeHTTP(w, r)
}
