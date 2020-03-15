package main

import (
	"math"

	"../../src/pregol"
)

// MakeShortestPath returns a function which is a pregel program for finding the
// single-source shortest paths from the vertex with id `sourceID`
func MakeShortestPath(sourceID int) pregol.UDF {
	var msgs map[int]float64
	if superstep == 0 {
		if vertex.Id == SourceId {
			vertex.Val = 0
		} else {
			vertex.Val = math.Inf(+1)
		}
	}
	newMin = false
	for _, msg := range vertex.InEdges {
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

func main() {
	shortestPath := MakeShortestPath(sourceID)
	pregol.SetUdf(shortestPath)
	pregol.Run()
}
