package pregol

// Worker represent a node in the distributed system
type Worker struct {
	ID         int
	allWorkers map[int]map[int][]int // {workerID:  {partitionId: [vertexIDs]}}
	inQueue    []string              // ["msg1", "msg2"] current incoming messages
	outQueue   []string              // next outgoing messages
	currVertex Vertex                // current active vertex
	nextVertex Vertex                // next active vertex
	Vertices   []*Vertex             // workers' own vertices
}

func (w *Worker) loadVertices() {
	// check whether assigned vertex is in assigned partition
	partitionsMap := w.allWorkers[w.ID]

	for k, v := range partitionsMap {
		// TODO
	}

	// if yes, add to w.Vertices
	// if no, send message to remote peer
}

func (w *Worker) startSuperstep() {
	partitions := w.allWorkers[w]

	// read inQueue

	for pID, vList := range partitions {
		go w.superstep(pID)
	}

	// wait until done -- how are we checking

	for pID, vList := range partitions {
		go receive(pID) // fill inQueue, receive messages from vertices and add to inQueue
		go send(pID)    // fill outQueue, send outQueue
	}

	// wait until done -- how are we checking

	// inform Master that superstep has completed
	w.sendActiveVertices()

}

func (w *Worker) superstep(pID int) {
	for v := range w.Vertices {
		go v.compute()
	}

}

func (w *Worker) sendActiveVertices() {
	// get list/number of active vertices
	// send list/number to Master
}

func (w *Worker) run() {
	for {
		select {
		case msg := <-w.masterInChan:
			switch {
			case msg == "Superstep":
				w.startSuperstep()
			case msg == "SaveState":
				//do something
			}
		}

	}
}
