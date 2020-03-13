package pregol

import (
	"fmt"
	"net/http"
	"sync"
)

// Worker ...
type Worker struct {
	ID     		int
	inQueue 	[]ResultMsg
	outQueue    []ResultMsg
	masterResp 	string
	partitions 	map[int][]Vertex
}

func newWorker(id int, ma string) *Worker {
	w := Worker{}
	w.ID = id

}

// InitWorkers ...
func (w *Worker) InitWorkers() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		http.HandleFunc("/", handleMessage)
		http.ListenAndServe(":3000", nil)
	}(&wg)

	wg.Wait()
	fmt.Print("Worker ", w.ID, "connected.")
}

// handleMaster incoming messages
// TODO: 
func handleMaster(w http.ResponseWriter, r *http.Request) {

	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Println("handler")
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

// loadVertices loads assigned vertices received from Master
func (w *Worker) loadVertices(gr graphReader) {
	partitionsMap := w.allWorkers[w.ID]

	for k, v := range partitionsMap {
		// TODO: populate vertices
	}

}

func (w *Worker) startSuperstep() {
	partitions := w.allWorkers[w]

	// TODO: read inQueue
	// TODO: send outQueue

	var wg sync.WaitGroup
	for pID, vList := range partitions {
		wg.Add(1)
		

		for pID, vList := range partitions {
			go func(vList []Vertex, udf, superstep){
				defer wg.Done()
				for v := range vList {
					ret = v.compute(udf, w, superstep)
					w.readMessage(ret)
					w.outQueue = append(w.outQueue, ret)
				}
			}(vList, udf, superstep)
		}

	}
	wg.Wait()

	// inform Master that superstep has completed
	w.sendActiveVertices()

}

func (w *Worker) readMessage(msg ResultMsg) {
	// fill outQueue 
	// TODO: Luoqi collate messages to each vertice

	activeVertices := []int

	// get list/number of active vertices
	for message := range w.inQueue {
		if message.halt == false {
			activeVertices = append(activeVertices, message.sendID)
		}
	}

	// send list/number to Master
	w.masterResp = activeVertices

}

func (w *Worker) sendActiveVertices() {
	// TODO: POST req to Master
}

func (w *Worker) Run() {
	// TODO: gerald
}