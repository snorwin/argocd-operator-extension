module github.com/snorwin/argocd-operator-extension

go 1.15

require (
	github.com/Azure/go-autorest/autorest/adal v0.9.10 // indirect
	github.com/argoproj-labs/argocd-operator v0.0.15
	github.com/go-logr/logr v0.3.0
	github.com/golang/mock v1.5.0
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.10.5
	github.com/snorwin/jsonpatch v1.4.0
	go.uber.org/zap v1.16.0
	helm.sh/helm/v3 v3.4.2
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.2.0
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	github.com/mikefarah/yaml/v2 => gopkg.in/yaml.v2 v2.4.0
	github.com/openshift/api => github.com/openshift/api v0.0.0-20190916204813-cdbe64fb0c91
	k8s.io/api => k8s.io/api v0.19.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.4
	k8s.io/client-go => k8s.io/client-go v0.19.4
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.0.0
	k8s.io/kubectl => k8s.io/kubectl v0.19.4
)
