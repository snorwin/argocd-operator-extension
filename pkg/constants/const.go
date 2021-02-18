package constants

const (
	// LabelArgoCDName - namespace label to specify the ArgoCD name
	LabelArgoCDName = "argocd.snorwin.io/name"
	// LabelArgoCDNamespace - namespace label to specify the ArgoCD namespace
	LabelArgoCDNamespace = "argocd.snorwin.io/namespace"

	// FinalizerName - name of the finalizer added to the ArgoCD instance
	FinalizerName = "uninstall.finalizers.argocd.snorwin.io"

	// EnvHelmDriver - helm storage driver (default: secret)
	EnvHelmDriver = "HELM_DRIVER"
	// EnvHelmMaxHistory - limit the maximum number of revisions saved per release. Use 0 for no limit. Default 10
	EnvHelmMaxHistory = "HELM_MAX_HISTORY"
	// EnvHelmDirectory - directory of the Helm chart
	EnvHelmDirectory = "HELM_DIRECTORY"
	// EnvClusterArgoCDNamespacedNames - comma separated list of NamespacedNames (namespace/name) of Argo CD instances which run in cluster mode
	EnvClusterArgoCDNamespacedNames = "CLUSTER_ARGOCD_NAMESPACEDNAMES"
)
