package constants

const (
	LabelArgoCDName      = "argocd.snorwin.io/name"
	LabelArgoCDNamespace = "argocd.snorwin.io/namespace"

	FinalizerName = "uninstall.finalizers.argocd.snorwin.io"

	// Helm storage driver (default: secret)
	EnvHelmDriver = "HELM_DRIVER"
	// Directory of the Helm chart
	EnvHelmDirectory = "HELM_DIRECTORY"
	// Comma separated list of NamespacedNames (namespace/name) of Argo CD instances which run in cluster mode
	EnvClusterArgoCDNamespacedNames = "CLUSTER_ARGOCD_NAMESPACEDNAMES"
)
