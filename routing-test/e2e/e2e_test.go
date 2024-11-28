/*
Copyright 2024.

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

package e2e

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting routelayer integration test suite\n")
	RunSpecs(t, "e2e suite")
}

var _ = Describe("RouteTest", Ordered, func() {

	var namespaceName, kubeResourcePath string
	var options *k8s.KubectlOptions
	var log *zap.Logger

	t := GinkgoT() // get a testing.T object to use with terratest

	// Before running the tests create the namespace and kustomize the example
	BeforeAll(func() {
		var err error
		log, err = zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())
		log.Info("Runnning Before")

		// create the namespace
		namespaceName = fmt.Sprintf("routing-demo-%s", strings.ToLower(random.UniqueId()))
		options = k8s.NewKubectlOptions("", "", namespaceName)
		k8s.CreateNamespaceWithMetadata(t, options, v1.ObjectMeta{
			Name:   namespaceName,
			Labels: map[string]string{"istio-injection": "enabled"},
		})

		// Kustomize apply the resources
		kubeResourcePath, err = filepath.Abs("./routing-test/resources/")
		Expect(err).NotTo(HaveOccurred())
		k8s.KubectlApplyFromKustomize(t, options, kubeResourcePath)

		// check the test-pod has come up and the deploymenmts being routed to.
		// This is important otherwise you will have a race condition for testing.
		k8s.WaitUntilPodAvailable(t, options, "test-pod", 10, time.Second*5)
		k8s.WaitUntilDeploymentAvailable(t, options, "http-echo-deployment-v1", 10, time.Second*5)
		k8s.WaitUntilDeploymentAvailable(t, options, "http-echo-deployment-v2", 10, time.Second*5)
		log.Info("Test pod has come up")
	})

	// After all tests have been executed, clean up dynamic namespace
	AfterAll(func() {
		// kustomize delete the resources and delete the namespace
		k8s.KubectlDeleteFromKustomize(t, options, kubeResourcePath)
		k8s.DeleteNamespace(t, options, namespaceName)
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
	})

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	Context("RouteTest", func() {

		It("should have created the http-echo service", func() {
			log.Info("Starting http-echo service lookup")
			service := k8s.GetService(t, options, "http-echo")
			Expect(service.Name).To(Equal("http-echo"))
		})

		It("should route to v1 when exec a curl to http-echo service", func() {
			httpServiceEndpoint := fmt.Sprintf("http://http-echo.%s.svc.cluster.local:8080/", namespaceName)
			// Exec into test-pod and run curl
			result, err := k8s.RunKubectlAndGetOutputE(t, options, "exec", "test-pod", "--",
				"curl", "-s", httpServiceEndpoint)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("v1"))
		})

		It("should route to v2 when exec a curl to http-echo service with header v2", func() {
			httpServiceEndpoint := fmt.Sprintf("http://http-echo.%s.svc.cluster.local:8080/", namespaceName)
			// Exec into test-pod and run curl
			result, err := k8s.RunKubectlAndGetOutputE(t, options, "exec", "test-pod", "--",
				"curl", "-sH", "x-route: v2", httpServiceEndpoint)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("v2"))
		})

		It("should route to v1 when exec a curl to http-echo service with header v1", func() {
			httpServiceEndpoint := fmt.Sprintf("http://http-echo.%s.svc.cluster.local:8080/", namespaceName)
			// Exec into test-pod and run curl
			result, err := k8s.RunKubectlAndGetOutputE(t, options, "exec", "test-pod", "--",
				"curl", "-sH", "x-route: v1", httpServiceEndpoint)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("v1"))
		})
	})
})
