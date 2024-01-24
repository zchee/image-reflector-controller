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
	"time"
)

type Tag struct {
	Name    string
	Created time.Time
}

type ByName []Tag

func (x ByName) Len() int           { return len(x) }
func (x ByName) Less(i, j int) bool { return x[i].Name < x[j].Name }
func (x ByName) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type ByCreated []Tag

func (x ByCreated) Len() int           { return len(x) }
func (x ByCreated) Less(i, j int) bool { return x[i].Created.Unix() < x[j].Created.Unix() }
func (x ByCreated) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// Policer is an interface representing a policy implementation type
type Policer interface {
	Latest([]Tag) (Tag, error)
}
