// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v3

// HTTPProxy helpers

import (
	sesame_api_v1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
)

func matchconditions(first sesame_api_v1.MatchCondition, rest ...sesame_api_v1.MatchCondition) []sesame_api_v1.MatchCondition {
	return append([]sesame_api_v1.MatchCondition{first}, rest...)
}

func prefixMatchCondition(prefix string) sesame_api_v1.MatchCondition {
	return sesame_api_v1.MatchCondition{
		Prefix: prefix,
	}
}

func headerContainsMatchCondition(name, value string) sesame_api_v1.MatchCondition {
	return sesame_api_v1.MatchCondition{
		Header: &sesame_api_v1.HeaderMatchCondition{
			Name:     name,
			Contains: value,
		},
	}
}

func headerNotContainsMatchCondition(name, value string) sesame_api_v1.MatchCondition {
	return sesame_api_v1.MatchCondition{
		Header: &sesame_api_v1.HeaderMatchCondition{
			Name:        name,
			NotContains: value,
		},
	}
}

func headerExactMatchCondition(name, value string) sesame_api_v1.MatchCondition {
	return sesame_api_v1.MatchCondition{
		Header: &sesame_api_v1.HeaderMatchCondition{
			Name:  name,
			Exact: value,
		},
	}
}

func headerNotExactMatchCondition(name, value string) sesame_api_v1.MatchCondition {
	return sesame_api_v1.MatchCondition{
		Header: &sesame_api_v1.HeaderMatchCondition{
			Name:     name,
			NotExact: value,
		},
	}
}

func headerPresentMatchCondition(name string) sesame_api_v1.MatchCondition {
	return sesame_api_v1.MatchCondition{
		Header: &sesame_api_v1.HeaderMatchCondition{
			Name:    name,
			Present: true,
		},
	}
}
