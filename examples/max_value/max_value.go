package main

import (
	"../../src/pregol"
)

func MaxStep(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
	var msgs map[int]float64
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
		for target := range vertex.OutEdges {
			msgs[target] = vertex.Val
		}
	}
	// 3. always halt, and send messages if any
	return true, msgs
}