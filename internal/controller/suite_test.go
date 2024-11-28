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

package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	routelayerv1 "github.com/fergalsomers/routelayer/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.31.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = routelayerv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Layer Reconciler", func() {
	Context("When reconciling a Layer", func() {
		const (
			resourceName = "test-layer"
			namespace    = "default"
		)

		ctx := context.Background()

		namespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}

		log := zap.New()

		var layer *routelayerv1.Layer

		// The test emvironment has a K8s control plane - which respects CREATE/GET/UPDATE/DELETE but we
		// have not registered a controller directly. This means we need to create a LayerReconciler and call
		// the reconcile method directly.

		BeforeEach(func() {
			layer = &routelayerv1.Layer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
			}
			err := k8sClient.Create(ctx, layer)
			Expect(err).NotTo(HaveOccurred())
			log.Info("layer", "resourceVersion", layer.ObjectMeta.ResourceVersion, "state", layer.Status.State)
		})

		AfterEach(func() {
			log.Info("layer", "resourceVersion", layer.ObjectMeta.ResourceVersion, "state", layer.Status.State)
			err := k8sClient.Delete(ctx, layer)
			Expect(err).NotTo(HaveOccurred())
			// force reconcilliation after delete to delete the item.
			// note this will remove the finalizers (i.e. we implicitly test deletion of finalizers works)
			req := ctrl.Request{NamespacedName: namespacedName}
			lc := &LayerReconciler{Client: k8sClient}
			lc.Reconcile(ctx, req)
		})

		It("Should set and remove finalizers correctly", func() {
			// Tests finalizers are set and unset correctly by the code
			// This ensures the item can be created.
			req := ctrl.Request{NamespacedName: namespacedName}
			lc := &LayerReconciler{Client: k8sClient}
			lc.Reconcile(ctx, req)

			// Finalizer should be set
			layer = &routelayerv1.Layer{}
			err := k8sClient.Get(ctx, namespacedName, layer)
			Expect(err).NotTo(HaveOccurred())
			Expect(layer.Finalizers).To(ContainElement(RouteLayerFinalizer))
		})

		It("Layer with no parent should be ready", func() {
			req := ctrl.Request{NamespacedName: namespacedName}
			lc := &LayerReconciler{Client: k8sClient}
			lc.Reconcile(ctx, req)

			l := &routelayerv1.Layer{}
			err := k8sClient.Get(ctx, namespacedName, l)
			Expect(err).NotTo(HaveOccurred())
			Expect(l.Status.State).To(Equal(ReadyState))
		})

		It("Layer with a parent should be waiting", func() {
			layer.Spec.Parent = "a-parent"
			err := k8sClient.Update(ctx, layer)
			Expect(err).NotTo(HaveOccurred())

			req := ctrl.Request{NamespacedName: namespacedName}
			lc := &LayerReconciler{Client: k8sClient}
			lc.Reconcile(ctx, req)

			l := &routelayerv1.Layer{}
			err = k8sClient.Get(ctx, namespacedName, l)
			Expect(err).NotTo(HaveOccurred())
			Expect(l.Status.State).To(Equal(WaitingState))
		})

		It("should update a Layer", func() {
			// TODO: implement update logic
		})

		It("should delete a Layer", func() {
			// TODO: implement delete logic
		})
	})
})
