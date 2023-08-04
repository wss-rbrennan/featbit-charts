// **NOTE**: we have build tags to differentiate kubernetes tests from non-kubernetes tests, and further differentiate helm
// tests. This is done because minikube is heavy and can interfere with docker related tests in terratest. Similarly, helm
// can overload the minikube system and thus interfere with the other kubernetes tests. Specifically, many of the tests
// start to fail with `connection refused` errors from `minikube`. To avoid overloading the system, we run the kubernetes
// tests and helm tests separately from the others. This may not be necessary if you have a sufficiently powerful machine.
// We recommend at least 4 cores and 16GB of RAM if you want to run all the tests together.

package test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
)

// This file contains examples of how to use terratest to test helm chart template logic by rendering the templates
// using `helm template`, and then reading in the rendered templates.
// There are two tests:
// - TestHelmBasicExampleTemplateRenderedDeployment: An example of how to read in the rendered object and check the
//   computed values.
// - TestHelmBasicExampleTemplateRequiredTemplateArgs: An example of how to check that the required args are indeed
//   required for the template to render.

// An example of how to verify the rendered template object of a Helm Chart given various inputs.
func TestHelmBasicExampleTemplateRenderedDeployment(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../charts/featbit")
	releaseName := "helm-basic"
	require.NoError(t, err)

	// Since we aren't deploying any resources, there is no need to setup kubectl authentication or helm home.

	// Set up the namespace; confirm that the template renders the expected value for the namespace.
	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())
	fullnameOverride := "featbit"
	logger.Logf(t, "Namespace: %s\n", namespaceName)

	// Setup the args. For this test, we will set the following input values:
	// - containerImageRepo=nginx
	// - containerImageTag=1.15.8
	options := &helm.Options{
		SetValues: map[string]string{
			"fullnameOverride":       fullnameOverride,
			"ui.image.registry":      "docker.io",
			"ui.image.repository":    "featbit/featbit-ui",
			"ui.image.pullPolicy":    "IfNotPresent",
			"ui.image.tag":           "2.4.1",
			"ui.autoscaling.enabled": "false",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	// Run RenderTemplate to render the template and capture the output. Note that we use the version without `E`, since
	// we want to assert that the template renders without any errors.
	// Additionally, although we know there is only one yaml file in the template, we deliberately path a templateFiles
	// arg to demonstrate how to select individual templates to render.
	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/ui-deployment.yaml"})

	// Now we use kubernetes/client-go library to render the template output into the Deployment struct. This will
	// ensure the Deployment resource is rendered correctly.
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	// Verify the namespace matches the expected supplied namespace.
	assert.Equal(t, namespaceName, deployment.Namespace)

	// Finally, we verify the deployment pod template spec is set to the expected container image value
	expectedContainerImage := "docker.io/featbit/featbit-ui:2.4.1"

	assert.Equal(t, deployment.Namespace, namespaceName)
	require.Equal(t, deployment.Name, "featbit-ui")
	require.Equal(t, deployment.Labels, map[string]string{
		"app.kubernetes.io/component":  "ui",
		"app.kubernetes.io/instance":   releaseName,
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "featbit",
		"helm.sh/chart":                "featbit-0.0.2",
	})
	require.Equal(t, deployment.Annotations, map[string]string{
		"meta.helm.sh/release-name":      releaseName,
		"meta.helm.sh/release-namespace": namespaceName,
	})

	require.Equal(t, int(*deployment.Spec.Replicas), int(1))
	require.Equal(t, intstr.IntOrString(*deployment.Spec.Strategy.RollingUpdate.MaxSurge), intstr.FromString("25%"))
	require.Equal(t, intstr.IntOrString(*deployment.Spec.Strategy.RollingUpdate.MaxUnavailable), intstr.FromString("25%"))
	require.Equal(t, deployment.Spec.Selector.MatchLabels, map[string]string{
		"app.kubernetes.io/component": "ui",
		"app.kubernetes.io/instance":  releaseName,
		"app.kubernetes.io/name":      "featbit",
	})

	require.Equal(t, deployment.Spec.Template.ObjectMeta.Labels, map[string]string{
		"app.kubernetes.io/name":      "featbit",
		"app.kubernetes.io/instance":  releaseName,
		"app.kubernetes.io/component": "ui",
	})

	require.Equal(t, deployment.Spec.Template.Spec.ServiceAccountName, "featbit")
	var podSecContext *corev1.PodSecurityContext = &corev1.PodSecurityContext{}
	require.Equal(t, deployment.Spec.Template.Spec.SecurityContext, podSecContext)
	//this getting long so lets shorten it
	podSpec := deployment.Spec.Template.Spec
	deploymentContainers := deployment.Spec.Template.Spec.Containers
	//InitContainers Section
	require.Equal(t, len(podSpec.InitContainers), 1)
	require.Equal(t, podSpec.InitContainers[0].Name, "wait-for-other-components")
	require.Equal(t, podSpec.InitContainers[0].Image, "docker.io/busybox:1.34")

	initExpectedCommand := []string{"/bin/sh", "-c", CreateCommandString(namespaceName, fullnameOverride, 5000, 5100, 8200)}

	require.Equal(t, podSpec.InitContainers[0].Command, initExpectedCommand)

	// var podSecContext *corev1.PodSecurityContext = &corev1.PodSecurityContext{}
	var containerSecContext *corev1.SecurityContext = &corev1.SecurityContext{}
	var contPullPolicy corev1.PullPolicy = corev1.PullIfNotPresent
	expectedCommand := []string{"/scripts/setup.sh"}
	//Containers Section
	require.Equal(t, len(deploymentContainers), 1)
	require.Equal(t, deploymentContainers[0].Name, "featbit-ui")
	require.Equal(t, deploymentContainers[0].SecurityContext, containerSecContext)
	require.Equal(t, deploymentContainers[0].Image, expectedContainerImage)
	require.Equal(t, deploymentContainers[0].ImagePullPolicy, contPullPolicy)
	require.Equal(t, deploymentContainers[0].Command, expectedCommand)
	require.Equal(t, deploymentContainers[0].Ports[0].Name, "http")
	require.Equal(t, int(deploymentContainers[0].Ports[0].ContainerPort), int(80))

}

// * TODO: toggle api, das, els
func CreateCommandString(namespaceName string, releaseName string, apiPort int, elsPort int, dasPort int) string {
	var strBytes bytes.Buffer

	strBytes.WriteString("\nuntil (nc -vz ")
	strBytes.WriteString(releaseName)
	strBytes.WriteString("-api.")
	strBytes.WriteString(namespaceName)
	strBytes.WriteString(".svc.cluster.local ")
	strBytes.WriteString(strconv.Itoa(apiPort))
	strBytes.WriteString("); do\n    echo \"waiting for API\"; sleep 1;\ndone\n\n\nuntil (nc -vz ")
	strBytes.WriteString(releaseName)
	strBytes.WriteString("-els.")
	strBytes.WriteString(namespaceName)
	strBytes.WriteString(".svc.cluster.local ")
	strBytes.WriteString(strconv.Itoa(elsPort))
	strBytes.WriteString("); do\n    echo \"waiting for Evaluation Server\"; sleep 1;\ndone\n\n\nuntil (nc -vz ")
	strBytes.WriteString(releaseName)
	strBytes.WriteString("-das.")
	strBytes.WriteString(namespaceName)
	strBytes.WriteString(".svc.cluster.local ")
	strBytes.WriteString(strconv.Itoa(dasPort))
	strBytes.WriteString("); do\n  echo \"waiting for DA Server\"; sleep 1;\ndone\n")

	return fmt.Sprint(strBytes.String())
}

// An example of how to verify required values for a helm chart.
// func TestHelmBasicExampleTemplateRequiredTemplateArgs(t *testing.T) {
// 	t.Parallel()

// 	// Path to the helm chart we will test
// 	helmChartPath, err := filepath.Abs("../charts/featbit")
// 	releaseName := "helm-basic"
// 	require.NoError(t, err)

// 	// Since we aren't deploying any resources, there is no need to setup kubectl authentication, helm home, or
// 	// namespaces

// 	// Here, we use a table driven test to iterate through all the required values as subtests. You can learn more about
// 	// go subtests here: https://blog.golang.org/subtests
// 	// The struct captures the inputs that we will pass to helm template and a human friendly name so we can identify it
// 	// in the test output. In this case, each test case will be a complete values input except for one of the required
// 	// values missing, to test that neglecting a required value will cause the template rendering to fail.
// 	testCases := []struct {
// 		name   string
// 		values map[string]string
// 	}{
// 		{
// 			"MissingContainerImageRepo",
// 			map[string]string{"containerImageTag": "1.15.8"},
// 		},
// 		{
// 			"MissingContainerImageTag",
// 			map[string]string{"containerImageRepo": "nginx"},
// 		},
// 	}

// 	// Now we iterate over each test case and spawn a sub test
// 	for _, testCase := range testCases {
// 		// Here, we capture the range variable and force it into the scope of this block. If we don't do this, when the
// 		// subtest switches contexts (because of t.Parallel), the testCase value will have been updated by the for loop
// 		// and will be the next testCase!
// 		testCase := testCase

// 		// The actual sub test spawning. We name the sub test using the human friendly name. Note that we name the sub
// 		// test T struct to subT to make it clear which T struct corresponds to which test. However, in most cases you
// 		// will not reference the main test T so you can name it the same.
// 		t.Run(testCase.name, func(subT *testing.T) {
// 			subT.Parallel()

// 			// Now we try rendering the template, but verify we get an error
// 			options := &helm.Options{SetValues: testCase.values}
// 			_, err := helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{})
// 			require.Error(t, err)
// 		})
// 	}
// }
