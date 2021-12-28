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

//go:build e2e
// +build e2e

package incluster

import (
	sesamev1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testSimpleSmoke(namespace string) {
	Specify("simple smoke test", func() {
		f.Fixtures.Echo.Deploy(namespace, "echo")

		p := &sesamev1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "smoke-test",
			},
			Spec: sesamev1.HTTPProxySpec{
				VirtualHost: &sesamev1.VirtualHost{
					Fqdn: "smoke-test.projectsesame.io",
				},
				Routes: []sesamev1.Route{
					{
						Services: []sesamev1.Service{
							{
								Name: "echo",
								Port: 80,
							},
						},
					},
				},
			},
		}
		f.CreateHTTPProxyAndWaitFor(p, e2e.HTTPProxyValid)

		res, ok := f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(f.T(), res, "request never succeeded")
		require.Truef(f.T(), ok, "expected 200 response code, got %d", res.StatusCode)
	})
}
