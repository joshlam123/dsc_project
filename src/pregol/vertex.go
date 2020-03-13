package pregol

type Message struct {
	recID int
	val   int
}

type Vertex struct {
	id       int
	flag     bool
	val      float64
	nowEdge  map[int]float64
	inEdges  chan map[int]float64
	outEdges map[int]float64
	vertices []Vertex // do i know my peers?
}

type UDF func(vertex *Vertex) (bool, map[int]float64)

func (v *Vertex) compute(udf UDF, owner Worker) {
	// do computations by iterating over messages from each incoming edge.
	select {
	case v.nowEdge = <-v.inEdges:
		v.flag, v.outEdges = udf(v)

		//TODO: worker-side need channel to receive incoming messages for this super step : inChan []chan map[string]float64
		owner.inChan[v.id] <- v.outEdges

	default:
		if v.flag {
			v.VoteToHalt(owner)
		}
	}
}

// TODO: Worker-side need a channel to receive votes to halt from each worker
// halt []chan
func (v *Vertex) VoteToHalt(owner Worker) {
	owner.halt[v.id] <- 0
}
