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
	"fmt"
	"sort"
)

const (
	// NewestOrderAsc ascending order
	NewestOrderAsc = "ASC"
	// NewestOrderDesc descending order
	NewestOrderDesc = "DESC"
)

// Newest representes a newest build ordering policy
type Newest struct {
	Order string
}

// NewNewest constructs a Newest object validating the provided
// order argument
func NewNewest(order string) (*Newest, error) {
	switch order {
	case "":
		order = NewestOrderDesc
	case NewestOrderAsc, NewestOrderDesc:
		break
	default:
		return nil, fmt.Errorf("invalid order argument provided: '%s', must be one of: %s, %s", order, NewestOrderAsc, NewestOrderDesc)
	}

	return &Newest{
		Order: order,
	}, nil
}

// Latest returns latest version from a provided list of strings
func (p *Newest) Latest(timestamp []Tag) (Tag, error) {
	if len(timestamp) == 0 {
		return Tag{}, fmt.Errorf("timestamp list argument cannot be empty")
	}

	sorted := ByCreated(timestamp)
	if p.Order == NewestOrderAsc {
		sort.Sort(sorted)
	} else {
		sort.Sort(sort.Reverse(sorted))
	}
	return sorted[0], nil
}
