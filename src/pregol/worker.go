package pregol

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

var w Worker = Worker{}
var UdfChan chan UDF = make(chan UDF)

// Worker ...
type Worker struct {
	ID          int
	inQueue     []float64 //TODO: do we need to send ID of senderVertex - ID is included in ResultMsg
	outQueue    map[int][]float64
	masterResp  string //TODO: change type
	partitions  map[int][]Vertex
	udf         UDF
	graphReader graphReader
}

func InitWorker() {
	w.udf = <-UdfChan
}

// loadVertices loads assigned vertices received from Master
func (w *Worker) createAndLoadVertices(gr graphReader) {
	// create Vertices
	for vID, vReader := range gr.Vertices {
		partID := getPartition(vID, gr.Info.NumPartitions)
		v := Vertex{vID, false, vReader.Value, make(map[int]float64), make(chan map[int]float64), make(map[int]float64)}

		// add to Worker's partition list
		if val, ok := w.partitions[partID]; ok {
			val = append(val, v)
		} else {
			w.partitions[partID] = []Vertex{v}
		}
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
	// call w.createAndLoadVertices(graphReader)
}

func startSuperstepHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "startedSuperstep")
}

func saveStateHandler(w http.ResponseWriter, r *http.Request) {

}

func pingHandler(w http.ResponseWriter, r *http.Request) {

}

func Run() {
	// TODO: gerald
	http.HandleFunc("/initConnection", initConnectionHandler)
	http.HandleFunc("/disseminateGraph", disseminateGraphHandler)
	http.HandleFunc("/startSuperstep", disseminateGraphHandler)
	http.HandleFunc("/saveState", saveStateHandler)
	http.HandleFunc("/ping", pingHandler)
	http.ListenAndServe(":3000", nil)
}
