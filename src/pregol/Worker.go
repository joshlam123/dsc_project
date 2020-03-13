package pregol

import (
	"fmt"
	"net/http"
	"sync"
)

// Worker ...
type Worker struct {
	ID     int
	inChan chan ResultMsg
	inQueue []ResultMsg
	// outQueue    []*Message
	masterResp string
	vertices []Vertex  // workers' own vertices
	// allWorkers  map[int]map[int][]int // {workerID:  {partitionId: [vertexIDs]}}
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

// handleMessage from Master
func handleMessage(w http.ResponseWriter, r *http.Request) {

	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Println("handler")
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}

	// for {
	// 	select {
	// 	case msg := <-w.masterInChan:
	// 		switch {
	// 		case msg == "Superstep":
	// 			w.startSuperstep()
	// 		case msg == "SaveState":
	// 			//do something
	// 		}
	// 	}
	// }

}

// loadVertices loads assigned vertices received from Master
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

	var wg sync.WaitGroup
	for pID, vList := range partitions {
		wg.Add(1)
		defer wg.Done()

		go func() {
			for v := range w.vertices {
				// take udf and superstep from Master
				go v.compute(udf, w, superstep)
			}
		}()

	}
	// wait until done
	wg.Wait()

	for pID, vList := range partitions {
		go rcvFromVertex(pID) // fill inQueue, receive messages from vertices and add to inQueue
		go sendToWorkers(pID)    // fill outQueue, send outQueue
	}

	// wait until done -- how are we checking

	// inform Master that superstep has completed
	w.sendActiveVertices()

}

// receive messages and halt votes after superstep
func (w *Worker) rcvFromVertex(pID int) {
	count := 0

	for v := range w.inChan {
		go func() {
			for {
				select {
				case msg := <-v:
					defer wg.Done()
					w.inQueue = append(w.inQueue, msg)

				// once worker receives all messages
				case count == len(w.vertices):
					return

				}
			}
		}()
	}

}

func (w *Worker) sendActiveVertices() {
	// POST req to Master

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