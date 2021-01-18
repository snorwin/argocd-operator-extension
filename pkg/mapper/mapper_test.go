package mapper_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	argoprojv1alpha1 "github.com/argoproj-labs/argocd-operator/pkg/apis/argoproj/v1alpha1"
	"github.com/snorwin/argocd-operator-extension/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/snorwin/argocd-operator-extension/pkg/mapper"
)

var _ = Describe("Mapper", func() {
	var (
		m *mapper.Mapper
	)
	Context("Map", func() {
		BeforeEach(func() {
			m = &mapper.Mapper{}
		})
		It("pod_with_two_configmaps_as_dependency", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "mypod", Namespace: "default"},
			}
			cm1 := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm1", Namespace: "default"},
			}
			cm2 := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm2", Namespace: "default"},
			}

			m.Graph.AddDependency(mapper.ReferenceFromObject(pod), mapper.ReferenceFromObject(cm1))
			m.Graph.AddDependency(mapper.ReferenceFromObject(pod), mapper.ReferenceFromObject(cm2))

			Ω(m.Map(handler.MapObject{Meta: pod, Object: pod})).
				Should(ConsistOf([]reconcile.Request{
					{NamespacedName: types.NamespacedName{Name: cm1.Name, Namespace: cm1.Namespace}},
					{NamespacedName: types.NamespacedName{Name: cm2.Name, Namespace: cm2.Namespace}},
				}))
		})
		It("namespace_add_missing_dependency", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "argoproj.io/v1alpha1",
					Kind:       "ArgoCD",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
				},
			}

			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mynamespace",
					Labels: map[string]string{
						constants.LabelArgoCDName:      argocd.Name,
						constants.LabelArgoCDNamespace: argocd.Namespace,
					},
				},
			}

			Ω(m.Map(handler.MapObject{Meta: namespace, Object: namespace})).
				Should(ConsistOf([]reconcile.Request{
					{NamespacedName: types.NamespacedName{Name: argocd.Name, Namespace: argocd.Namespace}},
				}))

			Ω(m.Graph.HasDependency(mapper.ReferenceFromObject(namespace), mapper.ReferenceFromObject(argocd))).Should(BeTrue())
			Ω(m.Graph.HasDependency(mapper.ReferenceFromObject(argocd), mapper.ReferenceFromObject(namespace))).Should(BeTrue())
		})
		It("namespace_with_dependency", func() {
			argocd := &argoprojv1alpha1.ArgoCD{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "argoproj.io/v1alpha1",
					Kind:       "ArgoCD",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
				},
			}

			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mynamespace",
				},
			}

			m.Graph.AddDependency(mapper.ReferenceFromObject(argocd), mapper.ReferenceFromObject(namespace))

			Ω(m.Map(handler.MapObject{Meta: namespace, Object: namespace})).
				Should(ConsistOf([]reconcile.Request{
					{NamespacedName: types.NamespacedName{Name: argocd.Name, Namespace: argocd.Namespace}},
				}))
		})
	})
})

var _ = Describe("Helper", func() {
	Context("ReferenceFromObject", func() {
		It("pod", func() {
			obj := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "mypod", Namespace: "default"},
			}

			ref := mapper.ReferenceFromObject(obj)

			Ω(ref.APIGroup).Should(Equal(obj.GroupVersionKind().Group))
			Ω(ref.Kind).Should(Equal(obj.GroupVersionKind().Kind))
			Ω(ref.Namespace).Should(Equal(obj.Namespace))
			Ω(ref.Name).Should(Equal(obj.Name))
		})
		It("namespace", func() {
			obj := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: "mynamespace"},
			}

			ref := mapper.ReferenceFromObject(obj)

			Ω(ref.APIGroup).Should(Equal(obj.GroupVersionKind().Group))
			Ω(ref.Kind).Should(Equal(obj.GroupVersionKind().Kind))
			Ω(ref.Namespace).Should(Equal(obj.Namespace))
			Ω(ref.Name).Should(Equal(obj.Name))
		})
		It("argocd", func() {
			obj := &argoprojv1alpha1.ArgoCD{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "argoproj.io/v1alpha1",
					Kind:       "ArgoCD",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argocd",
					Namespace: "default",
				},
			}
			ref := mapper.ReferenceFromObject(obj)

			Ω(ref.APIGroup).Should(Equal(obj.GroupVersionKind().Group))
			Ω(ref.Kind).Should(Equal(obj.GroupVersionKind().Kind))
			Ω(ref.Namespace).Should(Equal(obj.Namespace))
			Ω(ref.Name).Should(Equal(obj.Name))
		})
	})
	Context("ReferenceFromMapObject", func() {
		It("pod", func() {
			obj := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "mypod", Namespace: "default"},
			}

			mapObj := handler.MapObject{
				Meta:   obj,
				Object: obj,
			}
			ref := mapper.ReferenceFromMapObject(mapObj)

			Ω(ref.APIGroup).Should(Equal(obj.GroupVersionKind().Group))
			Ω(ref.Kind).Should(Equal(obj.GroupVersionKind().Kind))
			Ω(ref.Namespace).Should(Equal(obj.Namespace))
			Ω(ref.Name).Should(Equal(obj.Name))
		})
		It("namespace", func() {
			obj := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: "mynamespace"},
			}

			mapObj := handler.MapObject{
				Meta:   obj,
				Object: obj,
			}
			ref := mapper.ReferenceFromMapObject(mapObj)

			Ω(ref.APIGroup).Should(Equal(obj.GroupVersionKind().Group))
			Ω(ref.Kind).Should(Equal(obj.GroupVersionKind().Kind))
			Ω(ref.Namespace).Should(Equal(obj.Namespace))
			Ω(ref.Name).Should(Equal(obj.Name))
		})
	})
})
