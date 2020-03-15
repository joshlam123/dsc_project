package main

import (
	"../../src/pregol"
)

// PageRank represents a pregel program for calculating the PageRank score from a graph
func PageRank(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
	var msgs map[int]float64
	halt := false
	if superstep != 0 {
		var sum float64 = 0
		for _, msg := vertex.InEdges {
			sum += msg
		}
		vertex.Val = 0.15 / NumVertices() + 0.85 * sum
	}

	if superstep < 30 {
		n := len(vertex.InEdges)
		contrib := vertex.Val / n
		for target := range vertex.OutEdges {
			msgs[target] = contrib
		}
	} else {
		halt = true
	}
	return halt, msgs
}

func main() {
	pregol.SetUdf(PageRank)
	pregol.Run()
}
