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

package e2e

import (
	"context"

	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Fixtures holds references to all of the E2E fixtures helpers.
type Fixtures struct {
	// Echo provides helpers for working with the ingress-conformance-echo
	// test fixture.
	Echo *Echo

	// EchoSecure provides helpers for working with the TLS-secured
	// ingress-conformance-echo-tls test fixture.
	EchoSecure *EchoSecure
}

// Echo manages the ingress-conformance-echo fixture.
type Echo struct {
	client client.Client
	t      ginkgo.GinkgoTInterface
}

// Deploy runs DeployN with a default of 1 replica.
func (e *Echo) Deploy(ns, name string) func() {
	return e.DeployN(ns, name, 1)
}

// DeployN creates the ingress-conformance-echo fixture, specifically
// the deployment and service, in the given namespace and with the given name, or
// fails the test if it encounters an error. Number of replicas of the deployment
// can be configured. Namespace is defaulted to "default"
// and name is defaulted to "ingress-conformance-echo" if not provided. Returns
// a cleanup function.
func (e *Echo) DeployN(ns, name string, replicas int32) func() {
	valOrDefault := func(val, defaultVal string) string {
		if val != "" {
			return val
		}
		return defaultVal
	}

	ns = valOrDefault(ns, "default")
	name = valOrDefault(name, "ingress-conformance-echo")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app.kubernetes.io/name": name},
				},
				Spec: corev1.PodSpec{
					TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
						{
							// Attempt to spread pods across different nodes if possible.
							TopologyKey:       "kubernetes.io/hostname",
							MaxSkew:           1,
							WhenUnsatisfiable: corev1.ScheduleAnyway,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app.kubernetes.io/name": name},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "conformance-echo",
							Image: "gcr.io/k8s-staging-ingressconformance/echoserver@sha256:dc59c3e517399b436fa9db58f16506bd37f3cd831a7298eaf01bd5762ec514e1",
							Env: []corev1.EnvVar{
								{
									Name:  "INGRESS_NAME",
									Value: name,
								},
								{
									Name:  "SERVICE_NAME",
									Value: name,
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http-api",
									ContainerPort: 3000,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(3000),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	require.NoError(e.t, e.client.Create(context.TODO(), deployment))

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromString("http-api"),
				},
			},
			Selector: map[string]string{"app.kubernetes.io/name": name},
		},
	}
	require.NoError(e.t, e.client.Create(context.TODO(), service))

	return func() {
		require.NoError(e.t, e.client.Delete(context.TODO(), service))
		require.NoError(e.t, e.client.Delete(context.TODO(), deployment))
	}
}

// EchoSecure manages the TLS-secured ingress-conformance-echo fixture.
type EchoSecure struct {
	client client.Client
	t      ginkgo.GinkgoTInterface
}

// Deploy creates the TLS-secured ingress-conformance-echo-tls fixture, specifically
// the deployment and service, in the given namespace and with the given name, or
// fails the test if it encounters an error. Namespace is defaulted to "default"
// and name is defaulted to "ingress-conformance-echo-tls" if not provided. Returns
// a cleanup function.
func (e *EchoSecure) Deploy(ns, name string) func() {
	valOrDefault := func(val, defaultVal string) string {
		if val != "" {
			return val
		}
		return defaultVal
	}

	ns = valOrDefault(ns, "default")
	name = valOrDefault(name, "ingress-conformance-echo-tls")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app.kubernetes.io/name": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "conformance-echo",
							Image: "gcr.io/k8s-staging-ingressconformance/echoserver@sha256:dc59c3e517399b436fa9db58f16506bd37f3cd831a7298eaf01bd5762ec514e1",
							Env: []corev1.EnvVar{
								{
									Name:  "INGRESS_NAME",
									Value: name,
								},
								{
									Name:  "SERVICE_NAME",
									Value: name,
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name:  "TLS_SERVER_CERT",
									Value: "/run/secrets/certs/tls.crt",
								},
								{
									Name:  "TLS_SERVER_PRIVKEY",
									Value: "/run/secrets/certs/tls.key",
								},
								{
									Name:  "TLS_CLIENT_CACERTS",
									Value: "/run/secrets/certs/ca.crt",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http-api",
									ContainerPort: 3000,
								},
								{
									Name:          "https-api",
									ContainerPort: 8443,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(3000),
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/run/secrets/certs",
									Name:      "backend-server-cert",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "backend-server-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "backend-server-cert",
								},
							},
						},
					},
				},
			},
		},
	}
	require.NoError(e.t, e.client.Create(context.TODO(), deployment))

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			Annotations: map[string]string{
				"projectsesame.io/upstream-protocol.tls": "443",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromString("http-api"),
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromString("https-api"),
				},
			},
			Selector: map[string]string{"app.kubernetes.io/name": name},
		},
	}
	require.NoError(e.t, e.client.Create(context.TODO(), service))

	return func() {
		require.NoError(e.t, e.client.Delete(context.TODO(), service))
		require.NoError(e.t, e.client.Delete(context.TODO(), deployment))
	}
}

// DefaultSesameConfiguration returns a default SesameConfiguration object.
func DefaultSesameConfiguration() *sesame_api_v1alpha1.SesameConfiguration {
	return &sesame_api_v1alpha1.SesameConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress",
			Namespace: "projectsesame",
		},
		Spec: sesame_api_v1alpha1.SesameConfigurationSpec{
			Debug: sesame_api_v1alpha1.DebugConfig{
				Address:                 "127.0.0.1",
				Port:                    6060,
				DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
				KubernetesDebugLogLevel: 0,
			},
			Health: sesame_api_v1alpha1.HealthConfig{
				Address: "0.0.0.0",
				Port:    8000,
			},
			Envoy: sesame_api_v1alpha1.EnvoyConfig{
				DefaultHTTPVersions: []sesame_api_v1alpha1.HTTPVersionType{
					"HTTP/1.1", "HTTP/2",
				},
				Listener: sesame_api_v1alpha1.EnvoyListenerConfig{
					UseProxyProto:             false,
					DisableAllowChunkedLength: false,
					ConnectionBalancer:        "",
					TLS: sesame_api_v1alpha1.EnvoyTLS{
						MinimumProtocolVersion: "1.2",
						CipherSuites: []sesame_api_v1alpha1.TLSCipherType{
							"[ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]",
							"[ECDHE-RSA-AES128-GCM-SHA256|ECDHE-RSA-CHACHA20-POLY1305]",
							"ECDHE-ECDSA-AES256-GCM-SHA384",
							"ECDHE-RSA-AES256-GCM-SHA384",
						},
					},
				},
				Service: sesame_api_v1alpha1.NamespacedName{
					Name:      "envoy",
					Namespace: "projectsesame",
				},
				HTTPListener: sesame_api_v1alpha1.EnvoyListener{
					Address:   "0.0.0.0",
					Port:      8080,
					AccessLog: "/dev/stdout",
				},
				HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
					Address:   "0.0.0.0",
					Port:      8443,
					AccessLog: "/dev/stdout",
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8002,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8002,
				},
				Logging: sesame_api_v1alpha1.EnvoyLogging{
					AccessLogFormat: sesame_api_v1alpha1.EnvoyAccessLog,
				},
				Cluster: sesame_api_v1alpha1.ClusterParameters{
					DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
				},
				Network: sesame_api_v1alpha1.NetworkParameters{
					EnvoyAdminPort: 9001,
				},
			},
			HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
				DisablePermitInsecure: false,
			},
			EnableExternalNameService: false,
			Metrics: sesame_api_v1alpha1.MetricsConfig{
				Address: "0.0.0.0",
				Port:    8000,
			},
		},
	}
}

func IngressPathTypePtr(val networkingv1.PathType) *networkingv1.PathType {
	return &val
}
