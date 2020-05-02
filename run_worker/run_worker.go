package main

import (
	"dsc_project/src/pregol"
	"fmt"
	"math"
	"os"
)

// MaxValue represents the Pregel program for finding the max value among all vertices
func MaxValue(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
	fmt.Println("running maxvalue")
	fmt.Println("at superstep: ", superstep)
	var msgs map[int]float64 = make(map[int]float64)
	// algorithm:
	// 1. take max(value, incomingValues...)
	newMax := false
	for _, val := range vertex.InMsg {
		fmt.Println("inedges val: ", vertex.InMsg)
		if val > vertex.Val {
			vertex.Val = val
			fmt.Println("Weight of vertex ", vertex.Id, ": ", vertex.Val)
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

func PageRank(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
	var msgs map[int]float64
	halt := false
	if superstep != 0 {
		var sum float64 = 0
		for _, msg := range vertex.InMsg {
			sum += msg
		}
		vertex.Val = 0.15/vertex.NumVertices + 0.85*sum
	}

	if superstep < 30 {
		n := len(vertex.InMsg)
		contrib := vertex.Val / float64(n)
		for _, edge := range vertex.OutEdges {
			msgs[edge.VerticeID] = contrib
		}
	} else {
		halt = true
	}
	return halt, msgs
}

func MakeShortestPath(sourceID int) pregol.UDF {

	return func(vertex *pregol.Vertex, superstep int) (bool, map[int]float64) {
		var msgs = make(map[int]float64)

		newMin := false

		if superstep == 0 {
			if vertex.Id == sourceID {
				vertex.Val = 0.0
				newMin = true
			} else {
				vertex.Val = math.Inf(+1)
			}
		}

		for _, msg := range vertex.InMsg {
			if msg < vertex.Val {
				vertex.Val = msg
				newMin = true
			}
		}
		if newMin {
			for _, edge := range vertex.OutEdges {
				fmt.Println("Edge value:", math.Mod(edge.Weight, 1.0) == 0)
				msgs[edge.VerticeID] = vertex.Val + edge.Weight
			}
		}
		return true, msgs
	}
}

func main() {
	ports := os.Args[1:]
	pregol.RunUDF(MaxValue, ports)
	// pregol.RunUDF(MakeShortestPath(0), ports)
}
