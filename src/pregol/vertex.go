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
	inEdges  chan map[string]float64
	outEdges map[string]float64
	vertices []Vertex // do i know my peers?
}

func (v *Vertex) compute(UDF func(placeholder interface{}) interface{}, owner Worker) {
	// do computations by iterating over messages from each incoming edge.
	select {
	case msg := <-v.inEdges:
		v.flag = true
		for i := range msg {
			//TODO: Get return values from UDF

		}

		//TODO: Map return values {outEdge (string): value float64 ...}
		//TODO: set outgoing edges v.outEdges = xxx
		//TODO: worker-side need channel to receive incoming messages for this super step : inChan []chan map[string]float64
		owner.inChan[v.id] <- v.outEdges

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
