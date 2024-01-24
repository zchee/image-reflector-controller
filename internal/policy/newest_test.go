/*
Copyright 2024 The Flux authors

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
	"time"
)

func TestNewNewest(t *testing.T) {
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
			order: NewestOrderAsc,
		},
		{
			label: "With valid desc order",
			order: NewestOrderDesc,
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

func TestTestNewNewest_Latest(t *testing.T) {
	dummyCreatedAt := time.Date(2024, 01, 27, 03, 07, 0, 0, time.UTC)

	cases := []struct {
		label           string
		order           string
		versions        []Tag
		expectedVersion Tag
		expectErr       bool
	}{
		{
			label: "With Ubuntu CalVer",
			versions: []Tag{
				{Name: "16.04", Created: dummyCreatedAt.Add(5 * time.Hour)},
				{Name: "16.04.1", Created: dummyCreatedAt.Add(4 * time.Hour)},
				{Name: "16.10", Created: dummyCreatedAt.Add(3 * time.Hour)},
				{Name: "20.04", Created: dummyCreatedAt.Add(2 * time.Hour)},
				{Name: "20.10", Created: dummyCreatedAt.Add(1 * time.Hour)},
			},
			expectedVersion: Tag{Name: "16.04", Created: dummyCreatedAt.Add(5 * time.Hour)},
		},
		{
			label: "With Ubuntu CalVer ascending",
			versions: []Tag{
				{Name: "16.04", Created: dummyCreatedAt.Add(5 * time.Hour)},
				{Name: "16.04.1", Created: dummyCreatedAt.Add(4 * time.Hour)},
				{Name: "16.10", Created: dummyCreatedAt.Add(3 * time.Hour)},
				{Name: "20.04", Created: dummyCreatedAt.Add(2 * time.Hour)},
				{Name: "20.10", Created: dummyCreatedAt.Add(1 * time.Hour)},
			},
			order:           NewestOrderAsc,
			expectedVersion: Tag{Name: "20.10", Created: dummyCreatedAt.Add(1 * time.Hour)},
		},
		{
			label: "With Ubuntu code names",
			versions: []Tag{
				{Name: "xenial", Created: dummyCreatedAt.Add(3 * time.Hour)},
				{Name: "yakkety", Created: dummyCreatedAt.Add(4 * time.Hour)},
				{Name: "zesty", Created: dummyCreatedAt.Add(5 * time.Hour)},
				{Name: "artful", Created: dummyCreatedAt.Add(1 * time.Hour)},
				{Name: "bionic", Created: dummyCreatedAt.Add(2 * time.Hour)},
			},
			expectedVersion: Tag{Name: "zesty", Created: dummyCreatedAt.Add(5 * time.Hour)},
		},
		{
			label: "With Ubuntu code names ascending",
			versions: []Tag{
				{Name: "xenial", Created: dummyCreatedAt.Add(3 * time.Hour)},
				{Name: "yakkety", Created: dummyCreatedAt.Add(4 * time.Hour)},
				{Name: "zesty", Created: dummyCreatedAt.Add(5 * time.Hour)},
				{Name: "artful", Created: dummyCreatedAt.Add(1 * time.Hour)},
				{Name: "bionic", Created: dummyCreatedAt.Add(2 * time.Hour)},
			},
			order:           NewestOrderAsc,
			expectedVersion: Tag{Name: "artful", Created: dummyCreatedAt.Add(1 * time.Hour)},
		},
		{
			label: "With Unix Timestamps",
			versions: []Tag{
				{Name: "1606234201", Created: time.Unix(1606234201, 0)},
				{Name: "1606364286", Created: time.Unix(1606364286, 0)},
				{Name: "1606334092", Created: time.Unix(1606334092, 0)},
				{Name: "1606334284", Created: time.Unix(1606334284, 0)},
				{Name: "1606334201", Created: time.Unix(1606334201, 0)},
			},
			expectedVersion: Tag{Name: "1606364286", Created: time.Unix(1606364286, 0)},
		},
		{
			label: "With Unix Timestamps asc",
			versions: []Tag{
				{Name: "1606234201", Created: time.Unix(1606234201, 0)},
				{Name: "1606364286", Created: time.Unix(1606364286, 0)},
				{Name: "1606334092", Created: time.Unix(1606334092, 0)},
				{Name: "1606334284", Created: time.Unix(1606334284, 0)},
				{Name: "1606334201", Created: time.Unix(1606334201, 0)},
			},
			order:           NewestOrderAsc,
			expectedVersion: Tag{Name: "1606234201", Created: time.Unix(1606234201, 0)},
		},
		{
			label: "With Unix Timestamps prefix",
			versions: []Tag{
				{Name: "rel-1606234201", Created: time.Unix(1606234201, 0)},
				{Name: "rel-1606364286", Created: time.Unix(1606364286, 0)},
				{Name: "rel-1606334092", Created: time.Unix(1606334092, 0)},
				{Name: "rel-1606334284", Created: time.Unix(1606334284, 0)},
				{Name: "rel-1606334201", Created: time.Unix(1606334201, 0)},
			},
			expectedVersion: Tag{Name: "rel-1606364286", Created: time.Unix(1606364286, 0)},
		},
		{
			label: "With RFC3339",
			versions: []Tag{
				{Name: "2021-01-08T21-18-21Z", Created: time.Date(2021, 01, 8, 21, 18, 21, 0, time.UTC)},
				{Name: "2020-05-08T21-18-21Z", Created: time.Date(2020, 05, 8, 21, 18, 21, 0, time.UTC)},
				{Name: "2021-01-08T19-20-00Z", Created: time.Date(2021, 01, 8, 19, 20, 0, 0, time.UTC)},
				{Name: "1990-01-08T00-20-00Z", Created: time.Date(1990, 01, 8, 00, 20, 00, 0, time.UTC)},
				{Name: "2023-05-08T00-20-00Z", Created: time.Date(2023, 05, 8, 00, 20, 21, 0, time.UTC)},
			},
			expectedVersion: Tag{Name: "2023-05-08T00-20-00Z", Created: time.Date(2023, 05, 8, 00, 20, 21, 0, time.UTC)},
		},
		{
			label: "With RFC3339 asc",
			versions: []Tag{
				{Name: "2021-01-08T21-18-21Z", Created: time.Date(2021, 01, 8, 21, 18, 21, 0, time.UTC)},
				{Name: "2020-05-08T21-18-21Z", Created: time.Date(2020, 05, 8, 21, 18, 21, 0, time.UTC)},
				{Name: "2021-01-08T19-20-00Z", Created: time.Date(2021, 01, 8, 19, 20, 0, 0, time.UTC)},
				{Name: "1990-01-08T00-20-00Z", Created: time.Date(1990, 01, 8, 00, 20, 00, 0, time.UTC)},
				{Name: "2023-05-08T00-20-00Z", Created: time.Date(2023, 05, 8, 00, 20, 21, 0, time.UTC)},
			},
			order:           NewestOrderAsc,
			expectedVersion: Tag{Name: "1990-01-08T00-20-00Z", Created: time.Date(1990, 01, 8, 00, 20, 00, 0, time.UTC)},
		},
		{
			label:     "Empty version list",
			versions:  []Tag{},
			expectErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			policy, err := NewNewest(tt.order)
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
