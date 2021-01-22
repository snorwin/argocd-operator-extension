package argocd_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestArgocd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ArgoCD Controller Suite")
}
