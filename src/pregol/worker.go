package pregol

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

var w Worker = Worker{}

// Worker ...
type Worker struct {
	ID int
	//inQueue     []float64 //TODO: do we need to send ID of senderVertex - ID is included in ResultMsg
	//outQueue    map[int][]float64

	inQueue     []ResultMsg
	outQueue    map[int][]float64
	masterResp  string //TODO: change type
	partitions  map[int][]Vertex
	udf         UDF
	graphReader graphReader
}

// SetUdf sets the user-defined function for `w`
func SetUdf(udf UDF) {
	w.udf = udf
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

	sendOwn := make(map[int][]float64)
	proxyOut := make(map[int][]float64)

	for i := range w.inQueue {
		for j, k := range w.inQueue[i].msg {
			for l := range w.partitions[w.ID] {
				if j == w.partitions[w.ID][l].Id {
					sendOwn[j] = append(sendOwn[j], k)
				} else {
					proxyOut[j] = append(proxyOut[j], k)
				}
			}
		}
	}

	//send messages in outQueue to own vertices
	for m, n := range sendOwn {
		w.partitions[w.ID][m].InMsg <- n //TODO: fix referencing for correct vertex
	}

	w.outQueue = proxyOut
	for i := range w.outQueue {
		// TODO: send vertices to other workers
	}

	var wg sync.WaitGroup
	// add waitgroup for each partition: vertex list
	for range partitions {
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

func Run() {
	// TODO: gerald
	http.HandleFunc("/initConnection", initConnectionHandler)
	http.HandleFunc("/disseminateGraph", disseminateGraphHandler)
	http.HandleFunc("/startSuperstep", disseminateGraphHandler)
	http.HandleFunc("/saveState", saveStateHandler)
	http.HandleFunc("/ping", pingHandler)
	http.ListenAndServe(":3000", nil)
}
