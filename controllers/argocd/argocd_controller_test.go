package argocd_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	argoprojv1alpha1 "github.com/argoproj-labs/argocd-operator/pkg/apis/argoproj/v1alpha1"
	logr "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/snorwin/argocd-operator-extension/pkg/constants"
	"github.com/snorwin/argocd-operator-extension/pkg/helm"
	mock_helm "github.com/snorwin/argocd-operator-extension/pkg/mocks/helm"
	"helm.sh/helm/v3/pkg/chartutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client/fake"

	controller "github.com/snorwin/argocd-operator-extension/controllers/argocd"
)

var _ = Describe("Reconciler", func() {
	Context("Reconcile", func() {
		BeforeEach(func() {
			wd, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())
			d := filepath.Join(wd, "/../../helm/charts/argocd-operator-extension/resources")
			Ω(os.Setenv(constants.EnvHelmDirectory, d)).ShouldNot(HaveOccurred())

			Ω(os.Setenv(constants.EnvHelmDriver, "")).ShouldNot(HaveOccurred())

			Ω(os.Setenv(constants.EnvClusterArgoCDNamespacedNames, "")).ShouldNot(HaveOccurred())
		})
		It("should_install_helm_chart_and_add_finalizer", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
				},
			}

			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			mockHelm := mock_helm.NewMockClient(mockCtrl)
			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Finalizers).Should(ContainElement(constants.FinalizerName))
		})
		It("should_not_add_namespaces_for_cluster_argocd", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
				},
			}

			Ω(os.Setenv(constants.EnvClusterArgoCDNamespacedNames, fmt.Sprintf("%s/%s", argocd.Namespace, argocd.Name))).ShouldNot(HaveOccurred())

			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			mockHelm := mock_helm.NewMockClient(mockCtrl)
			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", nil), true).
				Return(nil)

			testReconcile(mockHelm, argocd)
		})
		It("should_include_labeled_namespaces", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
				},
			}

			namespaces := []*corev1.Namespace{
				{ObjectMeta: metav1.ObjectMeta{
					Name: "myapp1",
				}},
				{ObjectMeta: metav1.ObjectMeta{
					Name: "myapp2",
				}},
				{ObjectMeta: metav1.ObjectMeta{
					Name: "myapp3",
					Labels: map[string]string{
						constants.LabelArgoCDName:      argocd.Name,
						constants.LabelArgoCDNamespace: argocd.Namespace,
					},
				}},
				{ObjectMeta: metav1.ObjectMeta{
					Name: "myapp4",
					Labels: map[string]string{
						constants.LabelArgoCDName:      argocd.Name,
						constants.LabelArgoCDNamespace: argocd.Namespace,
					},
				}},
			}

			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			mockHelm := mock_helm.NewMockClient(mockCtrl)
			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"myapp3", "myapp4", "default"}), true).
				Return(nil)

			testReconcile(mockHelm, argocd, namespaces...)
		})
		It("should_uninstall_helm_chart_if_argocd_was_deleted", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "argocd",
					Namespace:         "default",
					Finalizers:        []string{constants.FinalizerName},
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			}

			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			mockHelm := mock_helm.NewMockClient(mockCtrl)
			mockHelm.
				EXPECT().
				Uninstall(argocd.Name).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Finalizers).ShouldNot(ContainElement(constants.FinalizerName))
		})
		It("should_ignore_uninstall_errors", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "argocd",
					Namespace:         "default",
					Finalizers:        []string{constants.FinalizerName},
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			}

			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			mockHelm := mock_helm.NewMockClient(mockCtrl)
			mockHelm.
				EXPECT().
				Uninstall(argocd.Name).
				Return(errors.New("uninstall: Release not loaded"))

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Finalizers).ShouldNot(ContainElement(constants.FinalizerName))
		})
		It("should_nop_if_argocd_does_not_exist", func() {
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			mockHelm := mock_helm.NewMockClient(mockCtrl)

			s := scheme.Scheme
			Ω(argoprojv1alpha1.SchemeBuilder.AddToScheme(s)).ShouldNot(HaveOccurred())
			cl := client.NewFakeClientWithScheme(s)

			r := &controller.Reconciler{
				Client: cl,
				Scheme: s,
				Log:    logr.NullLogger{},
				HelmFactory: func(_ string, _ ...helm.ClientOption) (helm.Client, error) {
					return mockHelm, nil
				},
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "argocd", Namespace: "default"},
			}

			result, err := r.Reconcile(req)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).ShouldNot(BeNil())
			Ω(result.Requeue).Should(BeFalse())
		})
	})
})

func testReconcile(mockHelm *mock_helm.MockClient, argocd *argoprojv1alpha1.ArgoCD, namespaces ...*corev1.Namespace) *argoprojv1alpha1.ArgoCD {
	s := scheme.Scheme
	Ω(argoprojv1alpha1.SchemeBuilder.AddToScheme(s)).ShouldNot(HaveOccurred())

	objects := []runtime.Object{argocd}
	for _, namespace := range namespaces {
		objects = append(objects, namespace)
	}
	cl := client.NewFakeClientWithScheme(s, objects...)

	r := &controller.Reconciler{
		Client: cl,
		Scheme: s,
		Log:    logr.NullLogger{},
		HelmFactory: func(_ string, _ ...helm.ClientOption) (helm.Client, error) {
			return mockHelm, nil
		},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{Name: argocd.Name, Namespace: argocd.Namespace},
	}

	result, err := r.Reconcile(req)
	Ω(err).ShouldNot(HaveOccurred())
	Ω(result).ShouldNot(BeNil())
	Ω(result.Requeue).Should(BeFalse())

	actual := &argoprojv1alpha1.ArgoCD{}
	Ω(cl.Get(context.TODO(), req.NamespacedName, actual)).ShouldNot(HaveOccurred())
	return actual
}

func Values(key string, value interface{}) gomock.Matcher {
	return valuesMatcher{key, value}
}

type valuesMatcher struct {
	key   string
	value interface{}
}

func (m valuesMatcher) Matches(x interface{}) bool {
	if jin, ok := x.(chartutil.Values); ok {
		return reflect.DeepEqual(jin[m.key], m.value)
	}
	return false
}

func (m valuesMatcher) String() string {
	return fmt.Sprintf("map[%v:%v ...]", m.key, m.value)
}
