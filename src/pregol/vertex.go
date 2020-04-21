package pregol

import "fmt"

type ResultMsg struct {
	sendId int
	halt   bool
	msg    map[int]float64
}

type Vertex struct {
	Id          int
	flag        bool
	Val         float64
	InMsg       []float64
	outMsg      map[int]float64
	OutEdges    []edgeReader
	NumVertices float64
}

type UDF func(vertex *Vertex, superstep int) (bool, map[int]float64)

func (v *Vertex) SetInEdge(newVal []float64) {
	fmt.Println(v.InMsg)
	//v.InMsg = append(v.InMsg, newVal...)

	for _, i := range newVal {
		v.InMsg = append(v.InMsg, i)
	}
	//fmt.Println("in edge updated in setinEdge: ", v.InMsg)
}

func (v *Vertex) Compute(udf UDF, superstep int) ResultMsg {
	v.outMsg = make(map[int]float64)
	if len(v.InMsg) != 0 {
		v.flag, v.outMsg = udf(v, superstep)
		v.InMsg = make([]float64, 0)
		return ResultMsg{v.Id, false, v.outMsg}
	} else {
		if v.flag {
			return v.VoteToHalt()
		} else {
			v.flag, v.outMsg = udf(v, superstep)
			v.InMsg = make([]float64, 0)
			return ResultMsg{v.Id, false, v.outMsg}
		}
	}
}

func (v *Vertex) VoteToHalt() ResultMsg {
	var m map[int]float64
	return ResultMsg{v.Id, true, m}
}
