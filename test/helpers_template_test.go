// **NOTE**: we have build tags to differentiate kubernetes tests from non-kubernetes tests, and further differentiate helm
// tests. This is done because minikube is heavy and can interfere with docker related tests in terratest. Similarly, helm
// can overload the minikube system and thus interfere with the other kubernetes tests. Specifically, many of the tests
// start to fail with `connection refused` errors from `minikube`. To avoid overloading the system, we run the kubernetes
// tests and helm tests separately from the others. This may not be necessary if you have a sufficiently powerful machine.
// We recommend at least 4 cores and 16GB of RAM if you want to run all the tests together.

package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestItShouldRemoveChartNameWhenSameAsReleaseName(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../charts/featbit")
	releaseName := "featbit"
	require.NoError(t, err)
	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())
	// fullnameOverride := "featbit"

	// Setup the args. For this test, we will set the following input values:
	options := &helm.Options{
		SetValues: map[string]string{
			"das.service.type": "ClusterIp",
		},
		Logger:         logger.Discard,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	// Run RenderTemplate to render the template and capture the output. Note that we use the version without `E`, since
	// we want to assert that the template renders without any errors.
	// Additionally, although we know there is only one yaml file in the template, we deliberately path a templateFiles
	// arg to demonstrate how to select individual templates to render.
	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/api-deployment.yaml"})

	// Now we use kubernetes/client-go library to render the template output into the Deployment struct. This will
	// ensure the Deployment resource is rendered correctly.
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	require.Equal(t, "featbit-api", deployment.Name)

}

func TestItShouldIncludeChartNameWhenNotSameAsReleaseName(t *testing.T) {
	t.Parallel()
	// logger.New(logger.Discard)
	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../charts/featbit")
	releaseName := "notsame"
	require.NoError(t, err)
	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())
	// fullnameOverride := "featbit"

	// Setup the args. For this test, we will set the following input values:
	options := &helm.Options{
		SetValues: map[string]string{
			"das.service.type": "ClusterIp",
		},
		Logger:         logger.Discard,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	// Run RenderTemplate to render the template and capture the output. Note that we use the version without `E`, since
	// we want to assert that the template renders without any errors.
	// Additionally, although we know there is only one yaml file in the template, we deliberately path a templateFiles
	// arg to demonstrate how to select individual templates to render.
	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/api-deployment.yaml"})

	// Now we use kubernetes/client-go library to render the template output into the Deployment struct. This will
	// ensure the Deployment resource is rendered correctly.
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	require.Equal(t, "notsame-featbit-api", deployment.Name)

}

// This file contains examples of how to use terratest to test helm chart template logic by rendering the templates
// using `helm template`, and then reading in the rendered templates.
// There are two tests:
// - TestHelmBasicExampleTemplateRenderedDeployment: An example of how to read in the rendered object and check the
//   computed values.
// - TestHelmBasicExampleTemplateRequiredTemplateArgs: An example of how to check that the required args are indeed
//   required for the template to render.

// An example of how to verify the rendered template object of a Helm Chart given various inputs.
// func TestApiServiceTemplateRenderedDeployment(t *testing.T) {
// 	t.Parallel()

// 	// Path to the helm chart we will test
// 	helmChartPath, err := filepath.Abs("../charts/featbit")
// 	releaseName := "helm-basic"
// 	require.NoError(t, err)

// 	// Since we aren't deploying any resources, there is no need to setup kubectl authentication or helm home.

// 	// Set up the namespace; confirm that the template renders the expected value for the namespace.
// 	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())
// 	fullnameOverride := "featbit"
// 	logger.Logf(t, "Namespace: %s\n", namespaceName)

// 	// Setup the args. For this test, we will set the following input values:
// 	// - containerImageRepo=nginx
// 	// - containerImageTag=1.15.8
// 	options := &helm.Options{
// 		SetValues: map[string]string{
// 			"fullnameOverride":        fullnameOverride,
// 			"api.image.registry":      "docker.io",
// 			"api.image.repository":    "featbit/featbit-api",
// 			"api.image.pullPolicy":    "IfNotPresent",
// 			"api.image.tag":           "2.4.1",
// 			"api.autoscaling.enabled": "false",
// 			"api.service.type":        "ClusterIp",
// 			"api.service.port":        "5000",
// 			"api.service.nodePort":    "30050",
// 		},
// 		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
// 	}

// 	// Run RenderTemplate to render the template and capture the output. Note that we use the version without `E`, since
// 	// we want to assert that the template renders without any errors.
// 	// Additionally, although we know there is only one yaml file in the template, we deliberately path a templateFiles
// 	// arg to demonstrate how to select individual templates to render.
// 	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/_helpers.tpl"})

// 	// Now we use kubernetes/client-go library to render the template output into the Service struct. This will
// 	// ensure the Service resource is rendered correctly.
// 	var service corev1.Service
// 	helm.UnmarshalK8SYaml(t, output, &service)

// 	// Verify the namespace matches the expected supplied namespace.
// 	assert.Equal(t, namespaceName, service.Namespace)

// 	require.Equal(t, "featbit-api", service.Name)
// 	require.Equal(t, map[string]string{
// 		"app.kubernetes.io/component":  "api",
// 		"app.kubernetes.io/instance":   releaseName,
// 		"app.kubernetes.io/managed-by": "Helm",
// 		"app.kubernetes.io/name":       "featbit",
// 		"helm.sh/chart":                "featbit-0.0.2",
// 	}, service.Labels)
// 	require.Equal(t, map[string]string{
// 		"meta.helm.sh/release-name":      releaseName,
// 		"meta.helm.sh/release-namespace": namespaceName,
// 	}, service.Annotations)

// 	require.Equal(t, corev1.ServiceType("ClusterIp"), service.Spec.Type)
// 	// if loadBalancer
// 	//  require.Equal(t, "loadBalancerIP", service.Spec.LoadBalancerIP) //refered to as staticIp in values
// 	require.Equal(t, intstr.FromInt(5000), service.Spec.Ports[0].TargetPort)
// 	//if ingress 80 else 5000 or custom value
// 	require.Equal(t, int32(5000), service.Spec.Ports[0].Port)
// 	require.Equal(t, corev1.ProtocolTCP, service.Spec.Ports[0].Protocol)
// 	//if nodeport
// 	// require.Equal(t, int32(30050), service.Spec.Ports[0].NodePort)

// 	require.Equal(t, service.Spec.Selector, map[string]string{
// 		"app.kubernetes.io/component": "api",
// 		"app.kubernetes.io/instance":  releaseName,
// 		"app.kubernetes.io/name":      "featbit",
// 	})

// }
