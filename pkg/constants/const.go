package constants

const (
	// AnnotationImageVersionUpdatePolicy - specify the update policy of the images and versions,
	// allowed values are: 'None', 'Always' or 'IfNotPresent' (default: 'None')
	AnnotationImageVersionUpdatePolicy = "argocd.snorwin.io/image-update-policy"
	// AnnotationHelmHash - hash to track the helm chart and values installed for this ArgoCD instance
	AnnotationHelmHash = "argocd.snorwin.io/helm-hash"

	// ImageVersionUpdatePolicy
	ImageVersionUpdatePolicyNone         = "None"
	ImageVersionUpdatePolicyAlways       = "Always"
	ImageVersionUpdatePolicyIfNotPresent = "IfNotPresent"

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
	// EnvClusterArgoCDNamespacedNames - comma separated list of NamespacedNames (namespace/name) of ArgoCD instances which run in cluster mode
	EnvClusterArgoCDNamespacedNames = "CLUSTER_ARGOCD_NAMESPACEDNAMES"
	// EnvArgoCDImage - ArgoCD image and version (<image>:<version>) used for automated version updates
	EnvArgoCDImage = "ARGOCD_IMAGE"
	// EnvDexImage - Dex image and version (<image>:<version>) used for automated version updates
	EnvDexImage = "DEX_IMAGE"
	// EnvRedisImage - Redis image and version (<image>:<version>) used for automated version updates
	EnvRedisImage = "REDIS_IMAGE"
)
