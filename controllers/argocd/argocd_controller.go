package argocd

import (
	"context"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/snorwin/argocd-operator-extension/pkg/constants"
	"github.com/snorwin/argocd-operator-extension/pkg/helm"
	"github.com/snorwin/argocd-operator-extension/pkg/mapper"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	argoprojv1alpha1 "github.com/argoproj-labs/argocd-operator/pkg/apis/argoproj/v1alpha1"
)

// Reconciler reconciles a ArgoCD object
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	// HelmFactory is a factory function to create new Helm clients
	HelmFactory helm.ClientFactory

	// mapper relates namespaces to ArgoCD instances and vice versa
	mapper mapper.Mapper
}

// +kubebuilder:rbac:groups=argoproj.io,resources=argocds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=argocds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete

// SetupWithManager register the ArgoCD Reconciler to the Manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set default factory if it was not set before
	if r.HelmFactory == nil {
		r.HelmFactory = helm.NewClientForNamespace
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&argoprojv1alpha1.ArgoCD{}).
		Watches(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: &r.mapper}).
		Complete(r)
}

// Reconcile create the RBAC (rolebindings, roles and servicaccounts) for an ArgoCD instance
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("argocd", req.NamespacedName)

	// get reconciled object
	obj := argoprojv1alpha1.ArgoCD{}
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		if errors.IsNotFound(err) {
			// return and don't requeue
			return reconcile.Result{}, nil
		}
		// error reading the object - requeue the request
		return reconcile.Result{}, err
	}

	// create a helm client and inject logger and storage driver
	helm, err := r.HelmFactory(req.Namespace, helm.WithLogger(logger), helm.WithHelmDriver(os.Getenv(constants.EnvHelmDriver)))
	if err != nil {
		return reconcile.Result{}, err
	}

	// remove all dependencies form mapper
	ref := mapper.ReferenceFromObject(&obj)
	r.mapper.Graph.RemoveAllDependenciesFor(ref)

	// handle finalizer during deletion
	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		if contains(obj.ObjectMeta.Finalizers, constants.FinalizerName) {
			// uninstall helm chart
			_ = helm.Uninstall(req.Name)

			// remove finalizer
			obj.ObjectMeta.Finalizers = remove(obj.ObjectMeta.Finalizers, constants.FinalizerName)
			if err := r.Update(ctx, &obj); err != nil {
				return ctrl.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// add finalizer
	if !contains(obj.ObjectMeta.Finalizers, constants.FinalizerName) {
		obj.ObjectMeta.Finalizers = add(obj.ObjectMeta.Finalizers, constants.FinalizerName)
		if err := r.Update(ctx, &obj); err != nil {
			return ctrl.Result{}, err
		}
	}

	// load helm chart and update dependency
	chart, err := loader.Load(os.Getenv(constants.EnvHelmDirectory))
	if err != nil {
		return reconcile.Result{}, err
	}

	// load values from chart
	values := chart.Values
	if values == nil {
		values = chartutil.Values{}
	}

	// specify namespaces if the ArgoCD instance is not running in cluster mode
	if !contains(strings.Split(os.Getenv(constants.EnvClusterArgoCDNamespacedNames), ","), req.NamespacedName.String()) {
		// load namespaces with matching labels and update dependencies
		namespaces := corev1.NamespaceList{}
		if err := r.List(ctx, &namespaces, client.MatchingLabels{
			constants.LabelArgoCDName:      req.Name,
			constants.LabelArgoCDNamespace: req.Namespace,
		}); err != nil {
			return reconcile.Result{}, err
		}

		var slice []string
		for _, namespace := range namespaces.Items {
			slice = append(slice, namespace.Name)

			// add dependency
			r.mapper.Graph.AddDependency(ref, mapper.ReferenceFromObject(&namespace))
		}
		slice = add(slice, req.Namespace)

		values["namespaces"] = slice
	}

	// upgrade or install helm chart
	if err = helm.Upgrade(req.Name, chart, values, true); err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// contains check if a string in a []string exists
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// remove a string from a []string if it exist
func remove(slice []string, str string) []string {
	for i, v := range slice {
		if v == str {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// add a string to a []string if it not exist
func add(slice []string, str string) []string {
	for _, v := range slice {
		if v == str {
			return slice
		}
	}
	return append(slice, str)
}
