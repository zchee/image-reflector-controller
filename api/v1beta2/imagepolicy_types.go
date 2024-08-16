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

package v1beta2

import (
	"github.com/fluxcd/pkg/apis/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ImagePolicyKind = "ImagePolicy"

// Deprecated: Use ImageFinalizer.
const ImagePolicyFinalizer = "finalizers.fluxcd.io"

// ImagePolicySpec defines the parameters for calculating the
// ImagePolicy.
type ImagePolicySpec struct {
	// ImageRepositoryRef points at the object specifying the image
	// being scanned
	// +required
	ImageRepositoryRef meta.NamespacedObjectReference `json:"imageRepositoryRef"`
	// Policy gives the particulars of the policy to be followed in
	// selecting the most recent image
	// +required
	Policy ImagePolicyChoice `json:"policy"`
	// FilterTags enables filtering for only a subset of tags based on a set of
	// rules. If no rules are provided, all the tags from the repository will be
	// ordered and compared.
	// +optional
	FilterTags *TagFilter `json:"filterTags,omitempty"`
}

// ImagePolicyChoice is a union of all the types of policy that can be
// supplied.
type ImagePolicyChoice struct {
	// SemVer gives a semantic version range to check against the tags
	// available.
	// +optional
	SemVer *SemVerPolicy `json:"semver,omitempty"`
	// Alphabetical set of rules to use for alphabetical ordering of the tags.
	// +optional
	Alphabetical *AlphabeticalPolicy `json:"alphabetical,omitempty"`
	// Numerical set of rules to use for numerical ordering of the tags.
	// +optional
	Numerical *NumericalPolicy `json:"numerical,omitempty"`
	// Newest set of rules to use for the creation date ordering of the tags.
	// +optional
	Newest *NewestPolicy `json:"newest,omitempty"`
}

// SemVerPolicy specifies a semantic version policy.
type SemVerPolicy struct {
	// Range gives a semver range for the image tag; the highest
	// version within the range that's a tag yields the latest image.
	// +required
	Range string `json:"range"`
}

// AlphabeticalPolicy specifies a alphabetical ordering policy.
type AlphabeticalPolicy struct {
	// Order specifies the sorting order of the tags. Given the letters of the
	// alphabet as tags, ascending order would select Z, and descending order
	// would select A.
	// +kubebuilder:default:="asc"
	// +kubebuilder:validation:Enum=asc;desc
	// +optional
	Order string `json:"order,omitempty"`
}

// NumericalPolicy specifies a numerical ordering policy.
type NumericalPolicy struct {
	// Order specifies the sorting order of the tags. Given the integer values
	// from 0 to 9 as tags, ascending order would select 9, and descending order
	// would select 0.
	// +kubebuilder:default:="asc"
	// +kubebuilder:validation:Enum=asc;desc
	// +optional
	Order string `json:"order,omitempty"`
}

// NewestPolicy specifies a creation date ordering policy.
type NewestPolicy struct {
	// Order specifies the sorting order of the tags. Based on the image build
	// creation time, ascending order would select oldest built tags, and descending order
	// would select the newly built tags.
	// +kubebuilder:default:="desc"
	// +kubebuilder:validation:Enum=asc;desc
	// +optional
	Order string `json:"order,omitempty"`
}

// TagFilter enables filtering tags based on a set of defined rules
type TagFilter struct {
	// Pattern specifies a regular expression pattern used to filter for image
	// tags.
	// +optional
	Pattern string `json:"pattern"`
	// Extract allows a capture group to be extracted from the specified regular
	// expression pattern, useful before tag evaluation.
	// +optional
	Extract string `json:"extract"`
}

// ImagePolicyStatus defines the observed state of ImagePolicy
type ImagePolicyStatus struct {
	// LatestImage gives the first in the list of images scanned by
	// the image repository, when filtered and ordered according to
	// the policy.
	LatestImage string `json:"latestImage,omitempty"`
	// ObservedPreviousImage is the observed previous LatestImage. It is used
	// to keep track of the previous and current images.
	// +optional
	ObservedPreviousImage string `json:"observedPreviousImage,omitempty"`
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// GetConditions returns the status conditions of the object.
func (p ImagePolicy) GetConditions() []metav1.Condition {
	return p.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (p *ImagePolicy) SetConditions(conditions []metav1.Condition) {
	p.Status.Conditions = conditions
}

// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="LatestImage",type=string,JSONPath=`.status.latestImage`

// ImagePolicy is the Schema for the imagepolicies API
type ImagePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ImagePolicySpec `json:"spec,omitempty"`
	// +kubebuilder:default={"observedGeneration":-1}
	Status ImagePolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ImagePolicyList contains a list of ImagePolicy
type ImagePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImagePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImagePolicy{}, &ImagePolicyList{})
}
