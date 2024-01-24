/*
Copyright 2020, 2021 The Flux authors

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

package policy

import (
	"testing"
)

func TestNewAlphabetical(t *testing.T) {
	cases := []struct {
		label     string
		order     string
		expectErr bool
	}{
		{
			label: "With valid empty order",
			order: "",
		},
		{
			label: "With valid asc order",
			order: AlphabeticalOrderAsc,
		},
		{
			label: "With valid desc order",
			order: AlphabeticalOrderDesc,
		},
		{
			label:     "With invalid order",
			order:     "invalid",
			expectErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			_, err := NewAlphabetical(tt.order)
			if tt.expectErr && err == nil {
				t.Fatalf("expecting error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("returned unexpected error: %s", err)
			}
		})
	}
}

func TestAlphabetical_Latest(t *testing.T) {
	cases := []struct {
		label           string
		order           string
		versions        []Tag
		expectedVersion Tag
		expectErr       bool
	}{
		{
			label:           "With Ubuntu CalVer",
			versions:        []Tag{{Name: "16.04"}, {Name: "16.04.1"}, {Name: "16.10"}, {Name: "20.04"}, {Name: "20.10"}},
			expectedVersion: Tag{Name: "20.10"},
		},
		{
			label:           "With Ubuntu CalVer descending",
			versions:        []Tag{{Name: "16.04"}, {Name: "16.04.1"}, {Name: "16.10"}, {Name: "20.04"}, {Name: "20.10"}},
			order:           AlphabeticalOrderDesc,
			expectedVersion: Tag{Name: "16.04"},
		},
		{
			label:           "With Ubuntu code names",
			versions:        []Tag{{Name: "xenial"}, {Name: "yakkety"}, {Name: "zesty"}, {Name: "artful"}, {Name: "bionic"}},
			expectedVersion: Tag{Name: "zesty"},
		},
		{
			label:           "With Ubuntu code names descending",
			versions:        []Tag{{Name: "xenial"}, {Name: "yakkety"}, {Name: "zesty"}, {Name: "artful"}, {Name: "bionic"}},
			order:           AlphabeticalOrderDesc,
			expectedVersion: Tag{Name: "artful"},
		},
		{
			label:           "With Timestamps",
			versions:        []Tag{{Name: "1606234201"}, {Name: "1606364286"}, {Name: "1606334092"}, {Name: "1606334284"}, {Name: "1606334201"}},
			expectedVersion: Tag{Name: "1606364286"},
		},
		{
			label:           "With Unix Timestamps desc",
			versions:        []Tag{{Name: "1606234201"}, {Name: "1606364286"}, {Name: "1606334092"}, {Name: "1606334284"}, {Name: "1606334201"}},
			order:           AlphabeticalOrderDesc,
			expectedVersion: Tag{Name: "1606234201"},
		},
		{
			label:           "With Unix Timestamps prefix",
			versions:        []Tag{{Name: "rel-1606234201"}, {Name: "rel-1606364286"}, {Name: "rel-1606334092"}, {Name: "rel-1606334284"}, {Name: "rel-1606334201"}},
			expectedVersion: Tag{Name: "rel-1606364286"},
		},
		{
			label:           "With RFC3339",
			versions:        []Tag{{Name: "2021-01-08T21-18-21Z"}, {Name: "2020-05-08T21-18-21Z"}, {Name: "2021-01-08T19-20-00Z"}, {Name: "1990-01-08T00-20-00Z"}, {Name: "2023-05-08T00-20-00Z"}},
			expectedVersion: Tag{Name: "2023-05-08T00-20-00Z"},
		},
		{
			label:           "With RFC3339 desc",
			versions:        []Tag{{Name: "2021-01-08T21-18-21Z"}, {Name: "2020-05-08T21-18-21Z"}, {Name: "2021-01-08T19-20-00Z"}, {Name: "1990-01-08T00-20-00Z"}, {Name: "2023-05-08T00-20-00Z"}},
			order:           AlphabeticalOrderDesc,
			expectedVersion: Tag{Name: "1990-01-08T00-20-00Z"},
		},
		{
			label:     "Empty version list",
			versions:  []Tag{},
			expectErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			policy, err := NewAlphabetical(tt.order)
			if err != nil {
				t.Fatalf("returned unexpected error: %s", err)
			}
			latest, err := policy.Latest(tt.versions)
			if tt.expectErr && err == nil {
				t.Fatalf("expecting error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("returned unexpected error: %s", err)
			}

			if latest != tt.expectedVersion {
				t.Errorf("incorrect computed version returned, got '%s', expected '%s'", latest, tt.expectedVersion)
			}
		})
	}
}
