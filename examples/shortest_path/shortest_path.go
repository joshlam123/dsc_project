package main

import (
	"math"

	"../../src/pregol"
)

// MakeShortestPath returns a function which is a pregel program for finding the
// single-source shortest paths from the vertex with id `sourceID`
func MakeShortestPath(sourceID int) pregol.UDF {
	return func(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
		var msgs map[int]float64
		if superstep == 0 {
			if vertex.Id == sourceID {
				vertex.Val = 0
			} else {
				vertex.Val = math.Inf(+1)
			}
		}
		newMin := false
		for _, msg := range vertex.InEdges {
			if msg < vertex.Val {
				vertex.Val = msg
				newMin = true
			}
		}
		if newMin {
			for _, edge := range vertex.OutEdges {
				msgs[edge.VerticeID] = vertex.Val + edge.Value
			}
		}
		return true, msgs
	}
}

func main() {
	shortestPath := MakeShortestPath(1)
	pregol.RunUDF(shortestPath)
}
