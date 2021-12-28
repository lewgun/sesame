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

package httpproxy

import (
	. "github.com/onsi/ginkgo"
	Sesamev1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testHostHeaderRewrite(namespace string) {
	Specify("hostname can be rewritten with policy on route", func() {
		t := f.T()

		f.Fixtures.Echo.Deploy(namespace, "ingress-conformance-echo")

		p := &Sesamev1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "host-header-rewrite",
			},
			Spec: Sesamev1.HTTPProxySpec{
				VirtualHost: &Sesamev1.VirtualHost{
					Fqdn: "hostheaderrewrite.projectsesame.io",
				},
				Routes: []Sesamev1.Route{
					{
						Services: []Sesamev1.Service{
							{
								Name: "ingress-conformance-echo",
								Port: 80,
							},
						},
						RequestHeadersPolicy: &Sesamev1.HeadersPolicy{
							Set: []Sesamev1.HeaderValue{
								{
									Name:  "Host",
									Value: "rewritten.com",
								},
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
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		assert.Equal(t, "rewritten.com", f.GetEchoResponseBody(res.Body).Host)
	})
}
