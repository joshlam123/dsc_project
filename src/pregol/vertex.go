package pregol

import "fmt"

type ResultMsg struct {
	sendId int
	halt   bool
	msg    map[int]float64
}

type Vertex struct {
	Id       int
	flag     bool
	Val      float64
	InEdges  []float64
	InMsg    chan []float64
	outMsg   map[int]float64
	OutEdges []edgeReader
}

type UDF func(vertex *Vertex, superstep int) (bool, map[int]float64)

func (v *Vertex) setInEdge(newVal []float64) {
	v.InEdges = append(v.InEdges, newVal...)
}

func (v *Vertex) Compute(udf UDF, superstep int) ResultMsg {
	// do computations by iterating over messages from each incoming edge.
	//select {
	//case v.InEdges = <-v.InMsg:
	//	v.flag, v.outMsg = udf(v, superstep)
	//	return ResultMsg{v.Id, false, v.outMsg}
	//
	//default:
	//	if v.flag {
	//		return v.VoteToHalt()
	//	} else {
	//		v.flag, v.outMsg = udf(v, superstep)
	//		return ResultMsg{v.Id, false, v.outMsg}
	//	}
	//}

	if v.InEdges != nil {
		v.flag, v.outMsg = udf(v, superstep)
		v.InEdges = nil
		return ResultMsg{v.Id, false, v.outMsg}
	} else {
		if v.flag {
			return v.VoteToHalt()
		} else {
			v.flag, v.outMsg = udf(v, superstep)
			return ResultMsg{v.Id, false, v.outMsg}
		}
	}
}

func (v *Vertex) VoteToHalt() ResultMsg {
	var m map[int]float64
	return ResultMsg{v.Id, true, m}
}

func main() {
	fmt.Println("hello world")
}
