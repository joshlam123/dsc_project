package pregol

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int   
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {     
      
	w.Header().Set("Content-Type", "application/json") 
	resp := Response {
				  Status: "200 OK", 
				  StatusCode: 200
			 } 

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(js)
	
}

// Worker ...
type Worker struct {
	ID          int
	masterAdrss string
	allWorkers  map[int]map[int][]int // {workerID:  {partitionId: [vertexIDs]}}
	inQueue     []*Message            
	outQueue    []*Message            
	currVertex  Vertex                
	nextVertex  Vertex                
	vertices    []Vertex              // workers' own vertices
}

// InitWorkers ...
func (w *Worker) InitWorkers() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func(ip string, wg *sync.WaitGroup) {
		defer wg.Done()

		http.HandleFunc("/json", jsonHandler)
		http.ListenAndServe(":3000", nil)

	}(w.masterAdrss, &wg)

	wg.Wait()
	close(masterChan)
	fmt.Print("Worker ", w.ID, "connected.")
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

func (w* Worker) receive(pID int) {
	for v := range w.vertices {
		go func() {
			
		}
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
