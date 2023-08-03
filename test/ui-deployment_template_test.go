package test

import (
	"testing"

	"k8s.io/client-go/kubernetes/scheme"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/gruntwork-io/terratest/modules/helm"
)

func TestPodTemplateRendersContainerImage(t *testing.T) {
	// Path to the helm chart we will test
	helmChartPath := "../charts/featbit"

	// Setup the args. For this test, we will set the following input values:
	// - image=nginx:1.15.8
	options := &helm.Options{
		SetValues: map[string]string{"ui.image.registry": "docker.io",
			"ui.image.repository": "featbit/featbit-ui-asdfasdf",
			"ui.image.pullPolicy": "IfNotPresent",
			"ui.image.tag":        "2.4.1"},
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode

	// Run RenderTemplate to render the template and capture the output.
	output := helm.RenderTemplate(t, options, helmChartPath, "deployment", []string{"templates/ui-deployment.yaml"})

	obj, gKV, _ := decode([]byte(output), nil, nil)

	// Now we use kubernetes/client-go library to render the template output into the Pod struct. This will
	// ensure the Pod resource is rendered correctly.

	// helm.UnmarshalK8SYaml(t, output, &deployment)

	if gKV.Kind == "Deployment" {
		// var deployment appsv1.Deployment
		deployment := obj.(*appsv1.Deployment)
		// Finally, we verify the pod spec is set to the expected container image value
		expectedContainerImage := "docker.io/featbit/featbit-ui:2.4.1"
		podContainers := deployment.Spec.Template.Spec.Containers
		if podContainers[0].Image != expectedContainerImage {
			t.Fatalf("Rendered container image (%s) is not expected (%s)", podContainers[0].Image, expectedContainerImage)
		}
	} else {
		t.Fatalf("Not a deployment")
	}

}
