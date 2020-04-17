package main

import (
	"fmt"
	"os"
	"pregol"
)

// MaxValue represents the Pregel program for finding the max value among all vertices
func MaxValue(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
	fmt.Println("running maxvalue")
	fmt.Println("at superstep: ", superstep)
	var msgs map[int]float64 = make(map[int]float64)
	// algorithm:
	// 1. take max(value, incomingValues...)
	newMax := false
	for _, val := range vertex.InEdges {
		fmt.Println("inedges val: ", vertex.InEdges)
		if val > vertex.Val {
			vertex.Val = val
			fmt.Println("Value of vertex ", vertex.Id, ": ", vertex.Val)
			newMax = true
		}
	}
	// 2. if at superstep 0, or if found new max, send value to target of each outgoing edge
	if superstep == 0 || newMax {
		for _, edge := range vertex.OutEdges {
			msgs[edge.VerticeID] = vertex.Val
		}
		fmt.Println("OutEdge: ", msgs)

	}
	// 3. always halt, and send messages if any
	//if len(msgs) != 0{
	//	return false, msgs
	//}

	return true, msgs
}

func Run() {
	ports := os.Args[1:]
	pregol.RunUDF(MaxValue, ports)
}
