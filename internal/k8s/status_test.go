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

package k8s_test

import (
	"testing"

	"github.com/projectsesame/sesame/internal/k8s"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func TestStatusUpdateHandlerRequiresLeaderElection(t *testing.T) {
	var s manager.LeaderElectionRunnable = &k8s.StatusUpdateHandler{}
	require.True(t, s.NeedLeaderElection())
}
