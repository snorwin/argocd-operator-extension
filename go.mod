module github.com/snorwin/argocd-operator-extension

go 1.15

require (
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.10 // indirect
	github.com/argoproj-labs/argocd-operator v0.0.14
	github.com/go-logr/logr v0.3.0
	github.com/golang/mock v1.4.3
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	go.uber.org/zap v1.14.1
	helm.sh/helm/v3 v3.4.2
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.2.0
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	k8s.io/api => k8s.io/api v0.19.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.4
	k8s.io/client-go => k8s.io/client-go v0.19.4
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.0.0
	k8s.io/kubectl => k8s.io/kubectl v0.19.4
)
