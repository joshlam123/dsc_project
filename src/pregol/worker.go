package pregol

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

// Worker ...
type Worker struct {
	ID         int
	inQueue    []float64 //TODO: do we need to send ID of senderVertex - ID is included in ResultMsg
	outQueue   map[int][]float64
	masterResp string //TODO: change type
	partitions map[int][]Vertex
}

func newWorker(id int, ma string) *Worker {
	w := Worker{}

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

// TODO: handleMaster incoming messages
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
	// add waitgroup for each partition: vertex list
	for _, _ = range partitions {
		wg.Add(1)

		for _, vList := range partitions {
			go func() {
				defer wg.Done()
				for v := range vList {
					ret = v.compute(udf, w, superstep)
					w.readMessage(ret)
				}
			}()
		}
	}
	wg.Wait()

	// inform Master that superstep has completed
	w.sendActiveVertices()

}

// reorder messages from vertices into outQueue and activeVertices
func (w *Worker) readMessage(rm ResultMsg) {

	for dest, m := range rm.msg {
		if v, ok := w.outQueue[dest]; ok {
			v = append(v, m)
		} else {
			w.outQueue[dest] = []float64{m}
		}
	}

	var activeVertices []int
	if rm.halt == false {
		activeVertices = append(activeVertices, rm.sendId)
	}

	// send list/number to Master
	w.masterResp = activeVertices

}

func (w *Worker) sendActiveVertices() {
	// TODO: POST req to Master
}

func initConnectionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "connected")
}

func disseminateGraphHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "received")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	// Handle Graph here
}

func startSuperstepHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "startedSuperstep")
}

func saveStateHandler(w http.ResponseWriter, r *http.Request) {

}

func pingHandler(w http.ResponseWriter, r *http.Request) {

}

func (w *Worker) Run() {
	// TODO: gerald
	http.HandleFunc("/initConnection", initConnectionHandler)
	http.HandleFunc("/disseminateGraph", disseminateGraphHandler)
	http.HandleFunc("/startSuperstep", disseminateGraphHandler)
	http.HandleFunc("/saveState", saveStateHandler)
	http.HandleFunc("/ping", pingHandler)
	http.ListenAndServe(":3000", nil)
}
