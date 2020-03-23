package main

import (
	"fmt"

	"../../src/pregol"
)

// MaxValue represents the Pregel program for finding the max value among all vertices
func MaxValue(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
	fmt.Println("At superstep: ", superstep)
	msgs := make(map[int]float64)
	// algorithm:
	// 1. take max(value, incomingValues...)
	newMax := false
	for _, val := range vertex.InEdges {
		if val > vertex.Val {
			vertex.Val = val
			newMax = true
		}
	}
	// 2. if at superstep 0, or if found new max, send value to target of each outgoing edge
	if superstep == 0 || newMax {
		for _, edge := range vertex.OutEdges {
			msgs[edge.VerticeID] = vertex.Val
		}
	}
	// 3. always halt, and send messages if any
	return true, msgs
}

func main() {
	pregol.RunUDF(MaxValue)
}
