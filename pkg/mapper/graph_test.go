package mapper_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/snorwin/argocd-operator-extension/pkg/mapper"
)

var _ = Describe("Graph", func() {
	var (
		graph *mapper.DependencyGraph
	)
	Context("AddDependency", func() {
		BeforeEach(func() {
			graph = &mapper.DependencyGraph{}
		})
		It("add_dependency", func() {
			a := mapper.Reference{Name: "A"}
			b := mapper.Reference{Name: "B"}

			// A - B
			graph.AddDependency(a, b)

			Ω(graph.HasDependency(a, b)).Should(BeTrue())
			Ω(graph.HasDependency(b, a)).Should(BeTrue())
		})
		It("add_multiple_dependency", func() {
			a := mapper.Reference{Name: "A"}
			b := mapper.Reference{Name: "B"}
			c := mapper.Reference{Name: "C"}

			// A - B - C
			graph.AddDependency(a, b)
			graph.AddDependency(b, c)

			Ω(graph.HasDependency(a, b)).Should(BeTrue())
			Ω(graph.HasDependency(b, a)).Should(BeTrue())
			Ω(graph.HasDependency(c, b)).Should(BeTrue())
			Ω(graph.HasDependency(b, c)).Should(BeTrue())
			Ω(graph.HasDependency(c, a)).Should(BeFalse())
			Ω(graph.HasDependency(a, c)).Should(BeFalse())
		})
	})
	Context("RemoveDependency", func() {
		BeforeEach(func() {
			graph = &mapper.DependencyGraph{}
		})
		It("remove_dependency", func() {
			a := mapper.Reference{Name: "A"}
			b := mapper.Reference{Name: "B"}
			c := mapper.Reference{Name: "C"}

			// A - B
			//  \ /
			//   C
			graph.AddDependency(a, b)
			graph.AddDependency(a, c)
			graph.AddDependency(b, c)

			Ω(graph.HasDependency(a, b)).Should(BeTrue())
			Ω(graph.HasDependency(b, a)).Should(BeTrue())
			Ω(graph.HasDependency(c, b)).Should(BeTrue())
			Ω(graph.HasDependency(b, c)).Should(BeTrue())
			Ω(graph.HasDependency(c, a)).Should(BeTrue())
			Ω(graph.HasDependency(a, c)).Should(BeTrue())

			// A   B
			//  \ /
			//   C
			graph.RemoveDependency(a, b)

			Ω(graph.HasDependency(a, b)).Should(BeFalse())
			Ω(graph.HasDependency(b, a)).Should(BeFalse())
			Ω(graph.HasDependency(c, b)).Should(BeTrue())
			Ω(graph.HasDependency(b, c)).Should(BeTrue())
			Ω(graph.HasDependency(c, a)).Should(BeTrue())
			Ω(graph.HasDependency(a, c)).Should(BeTrue())
		})
		It("remove_dependency_which_does_not_exist", func() {
			a := mapper.Reference{Name: "A"}
			b := mapper.Reference{Name: "B"}

			graph.RemoveDependency(a, b)

			Ω(graph.HasDependency(a, b)).Should(BeFalse())
			Ω(graph.HasDependency(b, a)).Should(BeFalse())
		})
	})
	Context("GetAllDependenciesFor", func() {
		BeforeEach(func() {
			graph = &mapper.DependencyGraph{}
		})
		It("get_all_dependencies", func() {
			a := mapper.Reference{Name: "A"}
			b := mapper.Reference{Name: "B"}
			c := mapper.Reference{Name: "C"}
			e := mapper.Reference{Name: "D"}
			d := mapper.Reference{Name: "E"}

			// A - B - D    E
			//  \ /
			//   C
			graph.AddDependency(a, b)
			graph.AddDependency(a, c)
			graph.AddDependency(b, c)
			graph.AddDependency(b, e)

			Ω(graph.GetAllDependenciesFor(a)).Should(ConsistOf(b, c))
			Ω(graph.GetAllDependenciesFor(b)).Should(ConsistOf(a, c, e))
			Ω(graph.GetAllDependenciesFor(c)).Should(ConsistOf(a, b))
			Ω(graph.GetAllDependenciesFor(e)).Should(ConsistOf(b))
			Ω(graph.GetAllDependenciesFor(d)).Should(BeEmpty())
		})
	})
	Context("RemoveAllDependenciesFor", func() {
		BeforeEach(func() {
			graph = &mapper.DependencyGraph{}
		})
		It("remove_all_dependencies", func() {
			a := mapper.Reference{Name: "A"}
			b := mapper.Reference{Name: "B"}
			c := mapper.Reference{Name: "C"}

			// A - B
			//  \ /
			//   C
			graph.AddDependency(a, b)
			graph.AddDependency(a, c)

			// A - B
			//
			//   C
			graph.RemoveAllDependenciesFor(c)

			Ω(graph.HasDependency(a, b)).Should(BeTrue())
			Ω(graph.HasDependency(b, a)).Should(BeTrue())
			Ω(graph.HasDependency(c, b)).Should(BeFalse())
			Ω(graph.HasDependency(b, c)).Should(BeFalse())
			Ω(graph.HasDependency(c, a)).Should(BeFalse())
			Ω(graph.HasDependency(a, c)).Should(BeFalse())
		})
	})
	Context("HasDependency", func() {
		BeforeEach(func() {
			graph = &mapper.DependencyGraph{}
		})
		It("has_dependency_composite_keys", func() {
			aa := mapper.Reference{Name: "A", Namespace: "AA"}
			ab := mapper.Reference{Name: "A", Namespace: "AB"}
			b := mapper.Reference{Name: "B"}

			// {A AA} - {A AB} - {B}
			graph.AddDependency(ab, b)
			graph.AddDependency(ab, aa)

			Ω(graph.HasDependency(aa, ab)).Should(BeTrue())
			Ω(graph.HasDependency(ab, aa)).Should(BeTrue())
			Ω(graph.HasDependency(ab, b)).Should(BeTrue())
			Ω(graph.HasDependency(b, ab)).Should(BeTrue())
			Ω(graph.HasDependency(aa, b)).Should(BeFalse())
			Ω(graph.HasDependency(b, aa)).Should(BeFalse())
		})
	})
})
