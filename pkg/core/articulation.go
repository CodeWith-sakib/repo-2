package core

import (
	"sort"
)

// ArticulationPointDetector identifies critical single points of failure in the DAG topology.
type ArticulationPointDetector struct {
	dag     *DAG
	timer   int
	tin     map[string]int
	low     map[string]int
	visited map[string]bool
	cutVertices map[string]bool
}

func NewArticulationPointDetector(dag *DAG) *ArticulationPointDetector {
	return &ArticulationPointDetector{
		dag:         dag,
		tin:         make(map[string]int),
		low:         make(map[string]int),
		visited:     make(map[string]bool),
		cutVertices: make(map[string]bool),
	}
}

func (d *ArticulationPointDetector) FindCutVertices() []string {
	nodes := d.dag.Nodes()
	for _, n := range nodes {
		if !d.visited[n.ID] {
			d.dfs(n.ID, "")
		}
	}

	var cuts []string
	for k := range d.cutVertices {
		cuts = append(cuts, k)
	}
	sort.Strings(cuts)
	return cuts
}

func (d *ArticulationPointDetector) dfs(v string, p string) {
	d.visited[v] = true
	d.tin[v] = d.timer
	d.low[v] = d.timer
	d.timer++
	children := 0

	// Undirected projection of the DAG to find structural bottlenecks
	neighbors := d.getUndirectedNeighbors(v)
	for _, to := range neighbors {
		if to == p {
			continue
		}
		if d.visited[to] {
			if d.tin[to] < d.low[v] {
				d.low[v] = d.tin[to]
			}
		} else {
			d.dfs(to, v)
			if d.low[to] < d.low[v] {
				d.low[v] = d.low[to]
			}
			if d.low[to] >= d.tin[v] && p != "" {
				d.cutVertices[v] = true
			}
			children++
		}
	}

	if p == "" && children > 1 {
		d.cutVertices[v] = true
	}
}

func (d *ArticulationPointDetector) getUndirectedNeighbors(v string) []string {
	seen := make(map[string]bool)
	var list []string
	for _, dep := range d.dag.GetDependencies(v) {
		if !seen[dep] {
			seen[dep] = true
			list = append(list, dep)
		}
	}
	for _, dep := range d.dag.GetDependents(v) {
		if !seen[dep] {
			seen[dep] = true
			list = append(list, dep)
		}
	}
	return list
}
