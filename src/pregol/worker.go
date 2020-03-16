package pregol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"golang.org/x/sync/semaphore"
)

var w Worker = Worker{}
var inQLock = sync.RWMutex{}
var outQLock = sync.RWMutex{}
var activeVertLock = sync.RWMutex{}            // ensure that one partition access activeVert variable at a time
var pingPong = semaphore.NewWeighted(int64(1)) // flag: (A) whether superstep is completed; (B) whether initVertices is done

// Worker ...
type Worker struct {
	ID          int
	inQueue     map[int][]float64
	outQueue    map[int][]float64
	activeVert  []int                  //TODO: change type
	partToVert  map[int]map[int]Vertex // partId: {verticeID: Vertex}
	udf         UDF
	graphReader graphReader
}

// InitWorker ...
func InitWorker(id int) {
	w.inQueue = make(map[int][]float64)
	w.outQueue = make(map[int][]float64)
	w.partToVert = make(map[int]map[int]Vertex)
}

// SetUdf sets the user-defined function for `w`
func SetUdf(udf UDF) {
	w.udf = udf
}

// loadVertices loads assigned vertices received from Master
func initVertices(gr graphReader) {
	// create Vertices

	if err := pingPong.Acquire(ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)
	}
	defer pingPong.Release(1)

	for vID, vReader := range gr.Vertices {
		partID := getPartition(vID, gr.Info.NumPartitions)
		v := Vertex{vID,
			false,
			vReader.Value,
			make([]float64, 0),
			make(chan []float64),
			make(map[int]float64),
			gr.Edges[vID]}

		// add to Worker's partition list
		w.partToVert[partID][v.Id] = v
	}
}

func startSuperstep() {

	if err := pingPong.Acquire(ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)
	}

	defer pingPong.Release(1)

	// sending values to vertices through InMsg Channel
	for vID, val := range w.inQueue {
		w.partToVert[w.ID][vID].setInEdge(val)
	}

	var wg sync.WaitGroup
	// add waitgroup for each partition: vertex list

	for _, vList := range w.partToVert {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for vID, v := range vList {
				ret := v.Compute(w.udf, superstep) //TODO: get superstep number
				// TODO: call processVertResult(ret)
			}
		}()
	}

	wg.Wait()
}

func disseminateMsgFromOutQ() {
	nodeToOutQ := make(map[int]map[int][]float64)

	for m, n := range w.outQueue {
		partID := getPartition(m, w.graphReader.Info.NumPartitions)
		workerID := w.graphReader.PartitionToNode[partID]
		nodeToOutQ[workerID][m] = n
	}

	for nodeID, outQ := range nodeToOutQ {
		if nodeID == w.graphReader.Info.NodeID {
			// send to own vertices

			go func(nodeID int, outQ map[int][]float64) {
				inQLock.Lock()
				defer inQLock.Unlock()

				for vID := range outQ {
					w.inQueue[vID] = append(w.inQueue[vID], outQ[vID]...)
				}
			}(nodeID, outQ)
		} else {
			workerIP := w.graphReader.ActiveNodes[nodeID].IP
			outQBytes, _ := json.Marshal(outQ)

			// TODO: send values to correct worker
			go func() {
				request, err := http.NewRequest("POST", "http://"+workerIP+":3000/incomingMsg", bytes.NewBuffer(outQBytes))
				if err != nil {
					log.Fatalln(err)
				}
			}()
		}
	}
	select {}
}

// Process results from vertices:
//     a) Populate outQueue with outgoing messages
//     b) Populate activeVert with vertices which are active at the end of superstep
// Requires concurrency controls as each partition will run it's own goroutine and call processVertResult multiple times
func processVertResult(rm ResultMsg) {

	// Populate outQueue with outgoing messages
	outQLock.Lock()
	for dstVert, msg := range rm.msg {
		if msgList, ok := w.outQueue[dstVert]; ok {
			msgList = append(msgList, msg)
		} else {
			w.outQueue[dstVert] = []float64{msg}
		}
	}
	outQLock.Unlock()

	// Populate activeVert with vertices which are active at the end of superstep
	if rm.halt == false {
		activeVertLock.Lock()
		w.activeVert = append(w.activeVert, rm.sendId)
		activeVertLock.Unlock()
	}
}

func initConnectionHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "connected")
}

func disseminateGraphHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "received")
	bodyBytes, err := ioutil.ReadAll(r.Body) //arr of bytes
	if err != nil {
		panic(err)
	}
	// bodyString := string(bodyBytes)
	// fmt.Println(bodyString)

	// get graph
	gr := getGraphFromJSONByte(bodyBytes)
	w.graphReader = gr

	go initVertices(gr)
}

func workerToWorkerHandler() {

}

func startSuperstepHandler(rw http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(rw, "startedSuperstep")
}

func saveStateHandler(rw http.ResponseWriter, r *http.Request) {
	// takes the format: writeToJson(jsonFile interface{}, name string) from util.go
	// send back graphReader and In/Out Queue

	// get the get request
	resp, err := r.Get(getURL(ip, "3000", "saveStateHandler"))
	if err != nil {
		log.Fatalln(err)
	}

	// define the format of the response
	w.graphReader
	w.partToVert

	// send back the response here - encoded as json or something

}

func pingHandler(rw http.ResponseWriter, r *http.Request) {
	// read the ping request
	resp, err := r.Get(getURL(ip, "3000", "pingHandler"))

	if err != nil {
		log.Fatalln(err)
	}

	// lock when accessing masterResponse
	// if unable to access semaphore, send "still not done"
	if err := sem.Acquire(ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)
		w.Write([]byte("Still not done"))
	}

	go func() {
		defer sem.Release(1)
		resp := map[string][]int{
			"Active Nodes": w.masterResp,
		}
		outBytes, error := json.Marshal(resp)
		if error != nil {
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(outBytes)
	}()
}

// Run ...
func Run() {
	// TODO: gerald
	http.HandleFunc("/initConnection", initConnectionHandler)
	http.HandleFunc("/disseminateGraph", disseminateGraphHandler)
	http.HandleFunc("/startSuperstep", disseminateGraphHandler)
	http.HandleFunc("/saveState", saveStateHandler)
	http.HandleFunc("/ping", pingHandler)
	http.ListenAndServe(":3000", nil)
}
