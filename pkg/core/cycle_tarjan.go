package core

import (
	"sort"
)

type TarjanSCCDetector struct {
	dag   *DAG
	index int
	stack []string
	onStack map[string]bool
	indices map[string]int
	lowlink map[string]int
	sccs    [][]string
}

func NewTarjanSCCDetector(dag *DAG) *TarjanSCCDetector {
	return &TarjanSCCDetector{
		dag:     dag,
		stack:   make([]string, 0),
		onStack: make(map[string]bool),
		indices: make(map[string]int),
		lowlink: make(map[string]int),
		sccs:    make([][]string, 0),
	}
}

// FindCycles returns all strongly connected components with size > 1 or self-loops.
func (d *TarjanSCCDetector) FindCycles() [][]string {
	nodes := d.dag.Nodes()
	for _, n := range nodes {
		if _, visited := d.indices[n.ID]; !visited {
			d.strongConnect(n.ID)
		}
	}

	var cycles [][]string
	for _, scc := range d.sccs {
		if len(scc) > 1 {
			sort.Strings(scc)
			cycles = append(cycles, scc)
		} else if len(scc) == 1 {
			// Check for self loop
			nodeID := scc[0]
			for _, dep := range d.dag.GetDependencies(nodeID) {
				if dep == nodeID {
					cycles = append(cycles, scc)
					break
				}
			}
		}
	}
	return cycles
}

func (d *TarjanSCCDetector) strongConnect(v string) {
	d.indices[v] = d.index
	d.lowlink[v] = d.index
	d.index++
	d.stack = append(d.stack, v)
	d.onStack[v] = true

	for _, w := range d.dag.GetDependents(v) {
		if _, visited := d.indices[w]; !visited {
			d.strongConnect(w)
			if d.lowlink[w] < d.lowlink[v] {
				d.lowlink[v] = d.lowlink[w]
			}
		} else if d.onStack[w] {
			if d.indices[w] < d.lowlink[v] {
				d.lowlink[v] = d.indices[w]
			}
		}
	}

	if d.lowlink[v] == d.indices[v] {
		var scc []string
		for {
			w := d.stack[len(d.stack)-1]
			d.stack = d.stack[:len(d.stack)-1]
			d.onStack[w] = false
			scc = append(scc, w)
			if w == v {
				break
			}
		}
		d.sccs = append(d.sccs, scc)
	}
}
