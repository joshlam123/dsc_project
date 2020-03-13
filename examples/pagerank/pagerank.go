package main

import (
	"../../src/pregol"
)

func PageRank(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
	var msgs map[int]float64
	var halt bool = false
	if superstep != 0 {
		var sum float64 = 0
		for msg := vertex.InMsgs {
			sum += msg
		}
		vertex.Val = 0.15 / NumVertices() + 0.85 * sum
	}

	if superstep < 30 {
		n := len(vertex.InMsgs)
		contrib := vertex.Val / n
		for target := range vertex.OutEdges {
			msgs[target] = contrib
		}
	} else {
		halt = true
	}
	return halt, msgs
}
