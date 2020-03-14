package main

import (
	"math"

	"../../src/pregol"
)

const SOURCE_ID int = 1

func ShortestPath(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
	var msgs map[int]float64
	if superstep == 0 {
		if vertex.Id == SOURCE_ID {
			vertex.Val = 0
		} else {
			vertex.Val = math.Inf(+1)
		}
	}
	newMin = false
	for _, msg := range vertex.InMsgs {
		if msg < vertex.Val {
			vertex.Val = msg
			newMin = true
		}
	}
	if newMin {
		for target, weight := range vertex.OutEdges {
			msgs[target] = vertex.Val + weight
		}
	}
	return true, msgs
}
