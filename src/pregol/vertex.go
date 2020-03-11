package pregol

import (
	"fmt"
	"strconv"
)

type Message struct {
	recID int
	val   int
}

type Vertex struct {
	id       int
	flag     bool
	val      float64
	inEdges  []chan map[string]float64
	outEdges map[string]float64
	vertices []Vertex // do i know my peers?
}

func (v *Vertex) compute(UDF func(placeholder interface{}) interface{}) {
	// do computations by iterating over messages from each incoming edge.
	select {
	case msg := <-v.inEdges:
		v.flag = true
		for i := range msg {
			//TODO: Get return values from UDF

		}

		//TODO: Map return values {outEdge (string): value float64 ...}
		//TODO: set outgoing edges v.outEdges = xxx

		for i, j := range v.outEdges {
			v.vertices[strconv.Atoi(i)].inEdges <- j
		}

	default:
		v.flag = false
		v.vote_to_halt()
	}
}

// TODO: Worker-side need a channel to receive votes to halt from each worker
// halt []chan
func (v *Vertex) vote_to_halt() {
	v.worker.halt[v.id] <- 0
}

func main() {
	fmt.Println("hello world")
}
