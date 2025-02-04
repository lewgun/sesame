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

package gateway

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	"github.com/projectsesame/sesame/internal/gatewayapi"
	"github.com/projectsesame/sesame/internal/k8s"
	"github.com/projectsesame/sesame/pkg/config"
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi_v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// Tests in this block set up/tear down their own GatewayClasses and Gateways.
var _ = Describe("GatewayClass/Gateway admission tests", func() {
	var (
		SesameCmd            *gexec.Session
		SesameConfig         *config.Parameters
		SesameConfiguration  *sesame_api_v1alpha1.SesameConfiguration
		SesameConfigFile     string
		additionalSesameArgs []string
		controllerName       string
	)

	BeforeEach(func() {
		controllerName = fmt.Sprintf("projectsesame.io/projectsesame/sesame-%d", getRandomNumber())

		// Sesame config file contents, can be modified in nested
		// BeforeEach.
		SesameConfig = &config.Parameters{
			GatewayConfig: &config.GatewayParameters{
				ControllerName: controllerName,
			},
		}

		// Update sesame configuration to point to specified gateway.
		SesameConfiguration = e2e.DefaultSesameConfiguration()
		SesameConfiguration.Spec.Gateway = &sesame_api_v1alpha1.GatewayConfig{
			ControllerName: controllerName,
		}

		// Default sesame serve command line arguments can be appended to in
		// nested BeforeEach.
		additionalSesameArgs = []string{}
	})

	// JustBeforeEach is called after each of the nested BeforeEach are
	// called, so it is a final setup step before running a test.
	// A nested BeforeEach may have modified Sesame config, so we wait
	// until here to start Sesame.
	JustBeforeEach(func() {
		var err error
		SesameCmd, SesameConfigFile, err = f.Deployment.StartLocalSesame(SesameConfig, SesameConfiguration, additionalSesameArgs...)
		require.NoError(f.T(), err)

		// Wait for Envoy to be healthy.
		require.NoError(f.T(), f.Deployment.WaitForEnvoyDaemonSetUpdated())
	})

	AfterEach(func() {
		require.NoError(f.T(), f.Client.DeleteAllOf(context.Background(), &gatewayapi_v1alpha2.GatewayClass{}))
		require.NoError(f.T(), f.Deployment.StopLocalSesame(SesameCmd, SesameConfigFile))
	})

	f.NamespacedTest("gateway-multiple-gatewayclasses", func(namespace string) {
		Specify("only the oldest matching gatewayclass should be accepted", func() {
			newGatewayClass := func(name, controller string) *gatewayapi_v1alpha2.GatewayClass {
				return &gatewayapi_v1alpha2.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: gatewayapi_v1alpha2.GatewayClassSpec{
						ControllerName: gatewayapi_v1alpha2.GatewayController(controller),
					},
				}
			}

			// create a non-matching GC: should not be accepted
			nonMatching := newGatewayClass("non-matching-gatewayclass", "projectsesame.io/non-matching-controller")

			require.NoError(f.T(), f.Client.Create(context.Background(), nonMatching))
			require.Never(f.T(), func() bool {
				if err := f.Client.Get(context.Background(), k8s.NamespacedNameOf(nonMatching), nonMatching); err != nil {
					return true
				}
				return gatewayClassValid(nonMatching)
			}, 5*time.Second, time.Second)

			// create a matching GC: should be accepted
			oldest := newGatewayClass("oldest-matching-gatewayclass", controllerName)
			_, valid := f.CreateGatewayClassAndWaitFor(oldest, gatewayClassValid)
			require.True(f.T(), valid)

			// create another matching GC: should not be accepted since it's not oldest
			secondOldest := newGatewayClass("second-oldest-matching-gatewayclass", controllerName)
			_, notOldest := f.CreateGatewayClassAndWaitFor(secondOldest, func(gc *gatewayapi_v1alpha2.GatewayClass) bool {
				for _, cond := range gc.Status.Conditions {
					if cond.Type == string(gatewayapi_v1alpha2.GatewayClassConditionStatusAccepted) &&
						cond.Status == metav1.ConditionFalse &&
						cond.Reason == "Invalid" &&
						cond.Message == "Invalid GatewayClass: another older GatewayClass with the same Spec.Controller exists" {
						return true
					}
				}
				return false
			})
			require.True(f.T(), notOldest)

			// double-check that the oldest matching GC is still accepted
			require.NoError(f.T(), f.Client.Get(context.Background(), k8s.NamespacedNameOf(oldest), oldest))
			require.True(f.T(), gatewayClassValid(oldest))

			// delete the first matching GC: second one should now be accepted
			require.NoError(f.T(), f.Client.Delete(context.Background(), oldest))
			require.Eventually(f.T(), func() bool {
				if err := f.Client.Get(context.Background(), k8s.NamespacedNameOf(secondOldest), secondOldest); err != nil {
					return false
				}
				return gatewayClassValid(secondOldest)
			}, f.RetryTimeout, f.RetryInterval)
		})
	})

	f.NamespacedTest("gateway-multiple-gateways", func(namespace string) {
		Specify("only the oldest gateway for the accepted gatewayclass should be accepted", func() {
			// Create a matching gateway class.
			gc := &gatewayapi_v1alpha2.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sesame-gatewayclass",
				},
				Spec: gatewayapi_v1alpha2.GatewayClassSpec{
					ControllerName: gatewayapi_v1alpha2.GatewayController(controllerName),
				},
			}
			_, valid := f.CreateGatewayClassAndWaitFor(gc, gatewayClassValid)
			require.True(f.T(), valid)

			// Create a matching gateway and verify it's accepted.
			oldest := &gatewayapi_v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "oldest",
					Namespace: namespace,
				},
				Spec: gatewayapi_v1alpha2.GatewaySpec{
					GatewayClassName: gatewayapi_v1alpha2.ObjectName(gc.Name),
					Listeners: []gatewayapi_v1alpha2.Listener{
						{
							Name:     "http",
							Protocol: gatewayapi_v1alpha2.HTTPProtocolType,
							Port:     gatewayapi_v1alpha2.PortNumber(80),
							AllowedRoutes: &gatewayapi_v1alpha2.AllowedRoutes{
								Namespaces: &gatewayapi_v1alpha2.RouteNamespaces{
									From: gatewayapi.FromNamespacesPtr(gatewayapi_v1alpha2.NamespacesFromSame),
								},
							},
						},
					},
				},
			}
			_, valid = f.CreateGatewayAndWaitFor(oldest, gatewayValid)
			require.True(f.T(), valid)

			// Create another matching gateway and verify it's not accepted.
			secondOldest := &gatewayapi_v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "second-oldest",
					Namespace: namespace,
				},
				Spec: gatewayapi_v1alpha2.GatewaySpec{
					GatewayClassName: gatewayapi_v1alpha2.ObjectName(gc.Name),
					Listeners: []gatewayapi_v1alpha2.Listener{
						{
							Name:     "http",
							Protocol: gatewayapi_v1alpha2.HTTPProtocolType,
							Port:     gatewayapi_v1alpha2.PortNumber(80),
							AllowedRoutes: &gatewayapi_v1alpha2.AllowedRoutes{
								Namespaces: &gatewayapi_v1alpha2.RouteNamespaces{
									From: gatewayapi.FromNamespacesPtr(gatewayapi_v1alpha2.NamespacesFromSame),
								},
							},
						},
					},
				},
			}
			_, notScheduled := f.CreateGatewayAndWaitFor(secondOldest, func(gw *gatewayapi_v1alpha2.Gateway) bool {
				for _, cond := range gw.Status.Conditions {
					if cond.Type == "Scheduled" &&
						cond.Status == metav1.ConditionFalse &&
						cond.Reason == "OlderGatewayExists" {
						return true
					}
				}
				return false
			})
			require.True(f.T(), notScheduled)

			// Double-check that the oldest gateway is still accepted.
			require.NoError(f.T(), f.Client.Get(context.Background(), k8s.NamespacedNameOf(oldest), oldest))
			require.True(f.T(), gatewayValid(oldest))

			// Delete the oldest gateway and verify that the second
			// oldest is now accepted.
			require.NoError(f.T(), f.Client.Delete(context.Background(), oldest))
			require.Eventually(f.T(), func() bool {
				if err := f.Client.Get(context.Background(), k8s.NamespacedNameOf(secondOldest), secondOldest); err != nil {
					return false
				}
				return gatewayValid(secondOldest)
			}, f.RetryTimeout, f.RetryInterval)
		})
	})

	f.NamespacedTest("gateway-multiple-classes-and-gateways", func(namespace string) {
		Specify("gatewayclass and gateway admission transitions properly when older gatewayclasses are deleted", func() {
			// Create a matching gateway class.
			olderGC := &gatewayapi_v1alpha2.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "older-gc",
				},
				Spec: gatewayapi_v1alpha2.GatewayClassSpec{
					ControllerName: gatewayapi_v1alpha2.GatewayController(controllerName),
				},
			}
			_, valid := f.CreateGatewayClassAndWaitFor(olderGC, gatewayClassValid)
			require.True(f.T(), valid)

			// Create a matching gateway and verify it's accepted.
			olderGCGateway1 := &gatewayapi_v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "older-gc-gw-1",
					Namespace: namespace,
				},
				Spec: gatewayapi_v1alpha2.GatewaySpec{
					GatewayClassName: gatewayapi_v1alpha2.ObjectName(olderGC.Name),
					Listeners: []gatewayapi_v1alpha2.Listener{
						{
							Name:     "http",
							Protocol: gatewayapi_v1alpha2.HTTPProtocolType,
							Port:     gatewayapi_v1alpha2.PortNumber(80),
							AllowedRoutes: &gatewayapi_v1alpha2.AllowedRoutes{
								Namespaces: &gatewayapi_v1alpha2.RouteNamespaces{
									From: gatewayapi.FromNamespacesPtr(gatewayapi_v1alpha2.NamespacesFromSame),
								},
							},
						},
					},
				},
			}
			_, valid = f.CreateGatewayAndWaitFor(olderGCGateway1, gatewayValid)
			require.True(f.T(), valid)

			// Create a second matching gatewayclass & 2 associated gateways
			// and verify none of them are accepted.
			newerGC := &gatewayapi_v1alpha2.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "newer-gc",
				},
				Spec: gatewayapi_v1alpha2.GatewayClassSpec{
					ControllerName: gatewayapi_v1alpha2.GatewayController(controllerName),
				},
			}
			require.NoError(f.T(), f.Client.Create(context.Background(), newerGC))
			require.Never(f.T(), func() bool {
				if err := f.Client.Get(context.Background(), k8s.NamespacedNameOf(newerGC), newerGC); err != nil {
					return true
				}
				return gatewayClassValid(newerGC)
			}, 5*time.Second, time.Second)

			newerGCGateway1 := &gatewayapi_v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "newer-gc-gw-1",
					Namespace: namespace,
				},
				Spec: gatewayapi_v1alpha2.GatewaySpec{
					GatewayClassName: gatewayapi_v1alpha2.ObjectName(newerGC.Name),
					Listeners: []gatewayapi_v1alpha2.Listener{
						{
							Name:     "http",
							Protocol: gatewayapi_v1alpha2.HTTPProtocolType,
							Port:     gatewayapi_v1alpha2.PortNumber(80),
							AllowedRoutes: &gatewayapi_v1alpha2.AllowedRoutes{
								Namespaces: &gatewayapi_v1alpha2.RouteNamespaces{
									From: gatewayapi.FromNamespacesPtr(gatewayapi_v1alpha2.NamespacesFromSame),
								},
							},
						},
					},
				},
			}
			require.NoError(f.T(), f.Client.Create(context.Background(), newerGCGateway1))
			require.Never(f.T(), func() bool {
				if err := f.Client.Get(context.Background(), k8s.NamespacedNameOf(newerGCGateway1), newerGCGateway1); err != nil {
					return true
				}
				return gatewayValid(newerGCGateway1)
			}, 5*time.Second, time.Second)

			newerGCGateway2 := &gatewayapi_v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "newer-gc-gw-2",
					Namespace: namespace,
				},
				Spec: gatewayapi_v1alpha2.GatewaySpec{
					GatewayClassName: gatewayapi_v1alpha2.ObjectName(newerGC.Name),
					Listeners: []gatewayapi_v1alpha2.Listener{
						{
							Name:     "http",
							Protocol: gatewayapi_v1alpha2.HTTPProtocolType,
							Port:     gatewayapi_v1alpha2.PortNumber(80),
							AllowedRoutes: &gatewayapi_v1alpha2.AllowedRoutes{
								Namespaces: &gatewayapi_v1alpha2.RouteNamespaces{
									From: gatewayapi.FromNamespacesPtr(gatewayapi_v1alpha2.NamespacesFromSame),
								},
							},
						},
					},
				},
			}
			require.NoError(f.T(), f.Client.Create(context.Background(), newerGCGateway2))
			require.Never(f.T(), func() bool {
				if err := f.Client.Get(context.Background(), k8s.NamespacedNameOf(newerGCGateway2), newerGCGateway2); err != nil {
					return true
				}
				return gatewayValid(newerGCGateway2)
			}, 5*time.Second, time.Second)

			// Now delete the older gatewayclass and associated gateway.
			require.NoError(f.T(), f.Client.Delete(context.Background(), olderGCGateway1))
			require.NoError(f.T(), f.Client.Delete(context.Background(), olderGC))

			// Verify that the newer gatewayclass and its oldest gateway are now accepted.
			require.Eventually(f.T(), func() bool {
				if err := f.Client.Get(context.Background(), k8s.NamespacedNameOf(newerGC), newerGC); err != nil {
					return false
				}
				return gatewayClassValid(newerGC)
			}, f.RetryTimeout, f.RetryInterval)

			require.Eventually(f.T(), func() bool {
				if err := f.Client.Get(context.Background(), k8s.NamespacedNameOf(newerGCGateway1), newerGCGateway1); err != nil {
					return false
				}
				return gatewayValid(newerGCGateway1)
			}, f.RetryTimeout, f.RetryInterval)
		})
	})
})
