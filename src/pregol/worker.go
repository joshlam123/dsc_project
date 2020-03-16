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
var inChanLock = sync.RWMutex{}
var semMaster = semaphore.NewWeighted(int64(1))

// Worker ...
type Worker struct {
	ID int
	//inQueue     []float64 //TODO: do we need to send ID of senderVertex - ID is included in ResultMsg
	//outQueue    map[int][]float64

	inQueue     map[int][]float64
	outQueue    map[int][]float64
	masterResp  []int                  //TODO: change type
	partitions  map[int]map[int]Vertex // partId: {verticeID: Vertex}
	udf         UDF
	graphReader graphReader
}

// init Worker
func initWorker(id int) Worker {

	w := Worker{
		ID:         id,
		inQueue:    make(map[int][]float64),
		outQueue:   make(map[int][]float64),
		partitions: make(map[int]map[int]Vertex),
	}
	return w
}

// SetUdf sets the user-defined function for `w`
func SetUdf(udf UDF) {
	w.udf = udf
}

// loadVertices loads assigned vertices received from Master
func createAndLoadVertices(gr graphReader) {
	// create Vertices
	for vID, vReader := range gr.Vertices {
		partID := getPartition(vID, gr.Info.NumPartitions)
		v := Vertex{vID, false, vReader.Value, make([]float64, 0), make(chan []float64), make(map[int]float64)}

		// add to Worker's partition list
		w.partitions[partID][v.Id] = v
	}
}

func startSuperstep() {

	if err := sem.Acquire(ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)

	}

	defer sem.Release(1)

	for nodeID, val := range w.inQueue {
    	w.partitions[w.ID][nodeID].InMsg <- val
  	}

	// proxyOut := make(map[int][]float64)

	// for i := range w.inQueue {
	// 	for j, k := range w.inQueue[i] {
	// 		proxyOut[j] = append(proxyOut[j], k)
	// 	}
	// }
	// w.outQueue = proxyOut

	var wg sync.WaitGroup
	// add waitgroup for each partition: vertex list

	for _, vList := range w.partitions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for vID, v := range vList {
				ret := v.Compute(w.udf, superstep) //TODO: get superstep number
				// TODO: call readMessage(ret)
			}
		}()
	}

	wg.Wait()

	// inform Master that superstep has completed
	w.sendActiveVertices()
}

func disseminateMsg() {
	nodeToOutQ := make(map[int]map[int][]float64)

	for m, n := range w.outQueue {
		//belong := false
		//for o := range w.partitions[w.ID] {
		//	if o == m {
		//		//send to own vertices directly if they belong in own partition
		//		belong = true
		//		w.partitions[w.ID][m].InMsg <- n //TODO: fix referencing for correct vertex
		//	}
		//}

		partID := getPartition(m, w.graphReader.Info.NumPartitions)
		workerID := w.graphReader.PartitionToNode[partID]
		nodeToOutQ[workerID][m] = n
	}

	for nodeID, outQ := range nodeToOutQ {

		if nodeID == w.graphReader.Info.NodeID {
			// send to own vertices

			go func(nodeID int, outQ map[int][]float64) {
				inChanLock.Lock()
				defer inChanLock.Unlock()

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
}

// reorder messages from vertices into outQueue and activeVertices
func readMessage(rm ResultMsg) {

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

	// fill list of active vertices to send to Master
	
	go func() {
		w.masterResp = append(w.masterResp, activeVertices)
	}()

}

func sendActiveVertices() {
	// TODO: POST req to Master

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
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)

	// get graph
	gr := getGraphFromJSONByte(bodyBytes)
	go w.createAndLoadVertices(gr)
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
	w.partitions

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
