package mapper

import (
	"sync"
)

type DependencyGraph struct {
	sync.RWMutex
	edges map[Reference]map[Reference]bool
}

func (g *DependencyGraph) AddDependency(obj1, obj2 Reference) {
	g.set(obj1, obj2, true)
}

func (g *DependencyGraph) RemoveDependency(obj1, obj2 Reference) {
	g.set(obj1, obj2, false)
}

func (g *DependencyGraph) GetAllDependenciesFor(obj1 Reference) []Reference {
	return g.getAll(obj1)
}

func (g *DependencyGraph) RemoveAllDependenciesFor(obj1 Reference) {
	for _, obj2 := range g.GetAllDependenciesFor(obj1) {
		g.RemoveDependency(obj1, obj2)
	}
}

func (g *DependencyGraph) HasDependency(obj1, obj2 Reference) bool {
	return g.get(obj1, obj2)
}

func (g *DependencyGraph) set(src, dst Reference, b bool) {
	g.Lock()
	defer g.Unlock()

	if g.edges == nil {
		g.edges = make(map[Reference]map[Reference]bool)
	}

	if _, ok := g.edges[src]; !ok {
		g.edges[src] = make(map[Reference]bool)
	}
	g.edges[src][dst] = b

	if _, ok := g.edges[dst]; !ok {
		g.edges[dst] = make(map[Reference]bool)
	}
	g.edges[dst][src] = b
}

func (g *DependencyGraph) get(src, dst Reference) bool {
	g.RLock()
	defer g.RUnlock()

	if g.edges == nil {
		return false
	}
	if _, ok := g.edges[src]; !ok {
		return false
	}
	if _, ok := g.edges[dst]; !ok {
		return false
	}

	return g.edges[dst][src] && g.edges[src][dst]
}

func (g *DependencyGraph) getAll(src Reference) []Reference {
	g.RLock()
	defer g.RUnlock()

	var ret []Reference

	if g.edges == nil {
		return ret
	}
	if _, ok := g.edges[src]; !ok {
		return ret
	}
	for dst, ok := range g.edges[src] {
		if ok {
			ret = append(ret, dst)
		}
	}

	return ret
}
