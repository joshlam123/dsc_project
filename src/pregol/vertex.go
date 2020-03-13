package pregol

import "fmt"

type ResultMsg struct {
	msg map[int]float64
}

type Vertex struct {
	Id     int
	flag   bool
	Val    float64
	Edges  map[int]float64
	InMsg  chan map[int]float64
	outMsg map[int]float64
	//vertices []Vertex // do i know my peers?
}

type UDF func(vertex *Vertex, superstep int) (bool, map[int]float64)

func (v *Vertex) compute(udf UDF, owner Worker, superstep int) {
	// do computations by iterating over messages from each incoming edge.
	select {
	case v.Edges = <-v.InMsg:
		v.flag, v.outMsg = udf(v, superstep)

		//TODO: worker-side need channel to receive incoming messages for this super step : inChan []chan map[int]float64
		owner.inChan <- v.outMsg

	default:
		if v.flag {
			v.VoteToHalt(owner)
		}
	}
}

// TODO: Worker-side need a channel to receive votes to halt from each worker: halt []chan int
func (v *Vertex) VoteToHalt(owner Worker) {
	owner.halt[v.Id] <- 0
}

func main() {
	fmt.Println("hello world")
}
