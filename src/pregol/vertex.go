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

func (v *Vertex) SetInEdge(newVal []float64) {
	fmt.Println(v.InEdges)
	//v.InEdges = append(v.InEdges, newVal...)

	for _, i := range newVal {
		v.InEdges = append(v.InEdges, i)
	}
	fmt.Println("in edge updated in setinEdge: ", v.InEdges)
}

func (v *Vertex) Compute(udf UDF, superstep int) ResultMsg {
	v.outMsg = make(map[int]float64)
	if len(v.InEdges) != 0 {
		v.flag, v.outMsg = udf(v, superstep)
		v.InEdges = make([]float64, 0)
		return ResultMsg{v.Id, false, v.outMsg}
	} else {
		if v.flag {
			return v.VoteToHalt()
		} else {
			v.flag, v.outMsg = udf(v, superstep)
			v.InEdges = make([]float64, 0)
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
