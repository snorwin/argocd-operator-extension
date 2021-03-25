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
	"github.com/snorwin/argocd-operator-extension/pkg/utils"
	"helm.sh/helm/v3/pkg/chart/loader"
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
		var (
			mockCtrl *gomock.Controller
			mockHelm *mock_helm.MockClient
		)
		BeforeEach(func() {
			// Set environment variables
			wd, err := os.Getwd()
			Ω(err).ShouldNot(HaveOccurred())
			d := filepath.Join(wd, "/../../helm/charts/argocd-operator-extension/resources")
			Ω(os.Setenv(constants.EnvHelmDirectory, d)).ShouldNot(HaveOccurred())

			Ω(os.Setenv(constants.EnvHelmDriver, "")).ShouldNot(HaveOccurred())

			Ω(os.Setenv(constants.EnvClusterArgoCDNamespacedNames, "")).ShouldNot(HaveOccurred())

			// Create helm mock client
			mockCtrl = gomock.NewController(GinkgoT())
			mockHelm = mock_helm.NewMockClient(mockCtrl)
		})
		AfterEach(func() {
			mockCtrl.Finish()

			Ω(os.Unsetenv(constants.EnvArgoCDImage)).ShouldNot(HaveOccurred())
			Ω(os.Unsetenv(constants.EnvDexImage)).ShouldNot(HaveOccurred())
			Ω(os.Unsetenv(constants.EnvRedisImage)).ShouldNot(HaveOccurred())
		})
		It("should_install_helm_chart_and_add_finalizer", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Finalizers).Should(ContainElement(constants.FinalizerName))
			Ω(actual.Annotations).Should(HaveKey(constants.AnnotationHelmHash))
		})
		It("should_set_argocd_image_and_version_if_not_present", func() {
			image := "argocd"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvArgoCDImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyIfNotPresent,
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Image).Should(Equal(image))
			Ω(actual.Spec.Version).Should(Equal(tag))
		})
		It("should_only_set_argocd_image_and_version_if_not_present", func() {
			image := "argocd"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvArgoCDImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyIfNotPresent,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Version: "latest",
					Image:   "myargocd",
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Image).Should(Equal(argocd.Spec.Image))
			Ω(actual.Spec.Version).Should(Equal(argocd.Spec.Version))
		})
		It("should_set_argocd_image_and_version_always", func() {
			image := "argocd"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvArgoCDImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Version: "latest",
					Image:   "myargocd",
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Image).Should(Equal(image))
			Ω(actual.Spec.Version).Should(Equal(tag))
		})
		It("should_set_argocd_image_only", func() {
			image := "argocd"

			Ω(os.Setenv(constants.EnvArgoCDImage, image)).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Version: "latest",
					Image:   "myargocd",
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Image).Should(Equal(image))
			Ω(actual.Spec.Version).Should(Equal(argocd.Spec.Version))
		})
		It("should_set_argocd_version_only", func() {
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvArgoCDImage, ":"+tag)).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Version: "latest",
					Image:   "myargocd",
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Image).Should(Equal(argocd.Spec.Image))
			Ω(actual.Spec.Version).Should(Equal(tag))
		})
		It("should_set_dex_image_and_version_if_not_present", func() {
			image := "dex"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvDexImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyIfNotPresent,
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Dex.Image).Should(Equal(image))
			Ω(actual.Spec.Dex.Version).Should(Equal(tag))
		})
		It("should_only_set_dex_image_and_version_if_not_present", func() {
			image := "dex"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvDexImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyIfNotPresent,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Dex: argoprojv1alpha1.ArgoCDDexSpec{
						Version: "latest",
						Image:   "mydex",
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Dex.Image).Should(Equal(argocd.Spec.Dex.Image))
			Ω(actual.Spec.Dex.Version).Should(Equal(argocd.Spec.Dex.Version))
		})
		It("should_set_dex_image_and_version_always", func() {
			image := "dex"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvDexImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Dex: argoprojv1alpha1.ArgoCDDexSpec{
						Version: "latest",
						Image:   "mydex",
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Dex.Image).Should(Equal(image))
			Ω(actual.Spec.Dex.Version).Should(Equal(tag))
		})
		It("should_set_dex_image_only", func() {
			image := "dex"

			Ω(os.Setenv(constants.EnvDexImage, image)).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Dex: argoprojv1alpha1.ArgoCDDexSpec{
						Version: "latest",
						Image:   "mydex",
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Dex.Image).Should(Equal(image))
			Ω(actual.Spec.Dex.Version).Should(Equal(argocd.Spec.Dex.Version))
		})
		It("should_set_dex_version_only", func() {
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvDexImage, ":"+tag)).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Dex: argoprojv1alpha1.ArgoCDDexSpec{
						Version: "latest",
						Image:   "mydex",
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Dex.Image).Should(Equal(argocd.Spec.Dex.Image))
			Ω(actual.Spec.Dex.Version).Should(Equal(tag))
		})
		It("should_set_redis_image_and_version_if_not_present", func() {
			image := "redis"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvRedisImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyIfNotPresent,
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Redis.Image).Should(Equal(image))
			Ω(actual.Spec.Redis.Version).Should(Equal(tag))
		})
		It("should_only_set_redis_image_and_version_if_not_present", func() {
			image := "redis"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvDexImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyIfNotPresent,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Redis: argoprojv1alpha1.ArgoCDRedisSpec{
						Version: "latest",
						Image:   "myredis",
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Redis.Image).Should(Equal(argocd.Spec.Redis.Image))
			Ω(actual.Spec.Redis.Version).Should(Equal(argocd.Spec.Redis.Version))
		})
		It("should_set_redis_image_and_version_always", func() {
			image := "redis"
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvRedisImage, fmt.Sprintf("%s:%s", image, tag))).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Redis: argoprojv1alpha1.ArgoCDRedisSpec{
						Version: "latest",
						Image:   "myredis",
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Redis.Image).Should(Equal(image))
			Ω(actual.Spec.Redis.Version).Should(Equal(tag))
		})
		It("should_set_redis_image_only", func() {
			image := "redis"

			Ω(os.Setenv(constants.EnvRedisImage, image)).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Redis: argoprojv1alpha1.ArgoCDRedisSpec{
						Version: "latest",
						Image:   "myredis",
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Redis.Image).Should(Equal(image))
			Ω(actual.Spec.Redis.Version).Should(Equal(argocd.Spec.Redis.Version))
		})
		It("should_set_redis_version_only", func() {
			tag := "v1.2.3"

			Ω(os.Setenv(constants.EnvRedisImage, ":"+tag)).ShouldNot(HaveOccurred())

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationImageVersionUpdatePolicy: constants.ImageVersionUpdatePolicyAlways,
					},
				},
				Spec: argoprojv1alpha1.ArgoCDSpec{
					Redis: argoprojv1alpha1.ArgoCDRedisSpec{
						Version: "latest",
						Image:   "myredis",
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Spec.Redis.Image).Should(Equal(argocd.Spec.Redis.Image))
			Ω(actual.Spec.Redis.Version).Should(Equal(tag))
		})
		It("should_not_upgrade_helm_chart_if_not_needed", func() {
			chart, err := loader.Load(os.Getenv(constants.EnvHelmDirectory))
			Ω(err).ShouldNot(HaveOccurred())

			values := map[string]interface{}{
				"namespaces": []string{"default"},
			}

			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationHelmHash: utils.Hash(chart, values),
					},
					ResourceVersion: "2",
					Finalizers: []string{
						constants.FinalizerName,
					},
				},
			}

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.ResourceVersion).Should(Equal(argocd.ResourceVersion))
		})
		It("should_upgrade_helm_chart_and_update_helm_hash", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
					Annotations: map[string]string{
						constants.AnnotationHelmHash: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
					},
					ResourceVersion: "2",
					Finalizers: []string{
						constants.FinalizerName,
					},
				},
			}

			mockHelm.
				EXPECT().
				Upgrade(argocd.Name, gomock.Any(), Values("namespaces", []string{"default"}), true).
				Return(nil)

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Annotations).Should(HaveKeyWithValue(constants.AnnotationHelmHash, Not(Equal(argocd.Annotations[constants.AnnotationHelmHash]))))
		})

		It("should_not_add_namespaces_for_cluster_argocd", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
				},
			}

			Ω(os.Setenv(constants.EnvClusterArgoCDNamespacedNames, fmt.Sprintf("%s/%s", argocd.Namespace, argocd.Name))).ShouldNot(HaveOccurred())

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

			mockHelm.
				EXPECT().
				Uninstall(argocd.Name).
				Return(errors.New("uninstall: Release not loaded"))

			actual := testReconcile(mockHelm, argocd)
			Ω(actual.Finalizers).ShouldNot(ContainElement(constants.FinalizerName))
		})
		It("should_nop_if_argocd_does_not_exist", func() {
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
