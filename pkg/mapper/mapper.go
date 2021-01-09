package mapper

import (
	"github.com/snorwin/argocd-operator-extension/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Mapper struct {
	Graph DependencyGraph
}

func (m *Mapper) Map(obj handler.MapObject) []reconcile.Request {
	var ret []reconcile.Request

	ref := ReferenceFromMapObject(obj)

	labels := obj.Meta.GetLabels()
	if labels[constants.LabelArgoCDName] != "" && labels[constants.LabelArgoCDNamespace] != "" {
		m.Graph.AddDependency(ref, Reference{
			APIGroup:  "argoproj.io",
			Kind:      "ArgoCD",
			Namespace: labels[constants.LabelArgoCDNamespace],
			Name:      labels[constants.LabelArgoCDName],
		})
	}

	for _, dependency := range m.Graph.GetAllDependenciesFor(ref) {
		ret = append(ret, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: dependency.Namespace,
				Name:      dependency.Name,
			},
		})
	}

	return ret
}

type Reference struct {
	APIGroup  string
	Kind      string
	Namespace string
	Name      string
}

func ReferenceFromMapObject(obj handler.MapObject) Reference {
	return Reference{
		APIGroup:  obj.Object.GetObjectKind().GroupVersionKind().Group,
		Kind:      obj.Object.GetObjectKind().GroupVersionKind().Kind,
		Namespace: obj.Meta.GetNamespace(),
		Name:      obj.Meta.GetName(),
	}
}

type Object interface {
	metav1.Object
	runtime.Object
}

func ReferenceFromObject(obj Object) Reference {
	return ReferenceFromMapObject(handler.MapObject{Meta: obj, Object: obj})
}
