package pregol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"golang.org/x/sync/semaphore"
)

var ctx = context.Background()
var w Worker = Worker{}
var inQLock = sync.RWMutex{}
var outQLock = sync.RWMutex{}
var activeVertLock = sync.RWMutex{}              // ensure that one partition access activeVert variable at a time
var pingPong = semaphore.NewWeighted(int64(1))   // flag: (A) whether superstep is completed; (B) whether initVertices is done
var busyWorker = semaphore.NewWeighted(int64(0)) // flag: Check if any goroutines are still handling incoming messages from peer workers

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

	defer pingPong.Release(1)

	// sending values to vertices through InMsg Channel
	for vID, val := range w.inQueue {
		w.partToVert[w.ID][vID].setInEdge(val)
	}

	// clearing queues so new values are not appended to old values (refresh for fresh superStep)
	// same for activeVert
	w.inQueue = make(map[int][]float64)
	w.outQueue = make(map[int][]float64)
	w.activeVert = make([]int, 0)

	var wg sync.WaitGroup
	for _, vList := range w.partToVert {
		wg.Add(1)   // add waitGroup for each partition: vertex list
		go func() { // for each partition, launch go routine to call compute for each of its vertex
			defer wg.Done()
			for _, v := range vList { // for each vertex in partition, compute().
				resultmsg := v.Compute(w.udf, 1) //TODO: get superstep number
				processVertResult(resultmsg)     //populate outQueue with return value of compute()
			}
		}()
	}
	wg.Wait()
	disseminateMsgFromOutQ() // send values to inqueue of respective worker nodes
}

func disseminateMsgFromOutQ() {
	// this function is called during startSuperstep() when the worker is disseminating the vertice values

	nodeToOutQ := make(map[int]map[int][]float64)

	// iterate over the worker's outqueue and prepare to disseminate it to the correct destination vertexID
	for m, n := range w.outQueue {

		outQLock.RLock()
		defer outQLock.RUnlock()

		partID := getPartition(m, w.graphReader.Info.NumPartitions)
		workerID := w.graphReader.PartitionToNode[partID]
		nodeToOutQ[workerID][m] = n
	}

	// set a waitgroup to wait for disseminating to all other workers (incl. yourself)
	var wg sync.WaitGroup

	for nodeID, outQ := range nodeToOutQ {
		wg.Add(1)

		if nodeID == w.graphReader.Info.NodeID {
			// send to own vertices

			// concurrent writes will happen in the inqueue
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

			// send values to correct worker
			go func() {
				_, err := http.NewRequest("POST", "http://"+workerIP+":3000/incomingMsg", bytes.NewBuffer(outQBytes))
				if err != nil {
					log.Fatalln(err)
				}
			}()
		}
	}
	wg.Wait()
}

// Process results from vertices:
//     a) Populate outQueue with outgoing messages
//     b) Populate activeVert with vertices which are active at the end of superstep
// Requires concurrency controls as each partition will run its own goroutine and call processVertResult multiple times
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
	w.graphReader = *gr

	go initVertices(*gr)

	if err := pingPong.Acquire(ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)
		fmt.Fprintf(rw, "NOT OK")
	} else {
		defer pingPong.Release(1)
		fmt.Fprintf(rw, "OK")
	}
}

func startSuperstepHandler(rw http.ResponseWriter, r *http.Request) {
	if err := pingPong.Acquire(ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)
	} else {
		fmt.Fprintf(rw, "startedSuperstep")
	}
}

func workerToWorkerHandler(rw http.ResponseWriter, r *http.Request) {
	// map[int][]float64
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// do something
	}
	var dstToVals map[int][]float64
	json.Unmarshal(bodyBytes, &dstToVals)
	busyWorker.Release(1)
	go func(dstToVals map[int][]float64) {
		defer busyWorker.Acquire(ctx, 1)
		inQLock.Lock()
		defer inQLock.Unlock()
		for dst, vals := range dstToVals {
			w.inQueue[dst] = append(w.inQueue[dst], vals...)
		}
	}(dstToVals)
}

func saveStateHandler(rw http.ResponseWriter, r *http.Request) {
	// takes the format: writeToJson(jsonFile interface{}, name string) from util.go
	// send back graphReader and In/Out Queue

	// get the get request

	// define the format of the response

	// send back the response here - encoded as json or something

}

func pingHandler(rw http.ResponseWriter, r *http.Request) {
	// read the ping request
	bodyByte, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyByte)

	if bodyString == "Completed graphHandler?" {
		if pingPong.TryAcquire(1) == false {
			log.Printf("Failed to acquire semaphore")
			//rw.Write([]byte("Still not done"))
			fmt.Fprintf(rw, "still not done")
		} else {
			defer pingPong.Release(1)
			fmt.Fprintf(rw, "done")
		}
	} else if bodyString == "Completed Superstep?" {
		// lock when accessing masterResponse
		// if unable to access semaphore, send "still not done" to master
		if pingPong.TryAcquire(1) == false {
			log.Printf("Failed to acquire semaphore")
			//rw.Write([]byte("Still not done"))
			fmt.Fprintf(rw, "still not done")
		} else {

			defer pingPong.Release(1)

			if busyWorker.TryAcquire(1) == false {

			} else {
				defer busyWorker.Release(1)
				//resp := map[string][]int{
				//	"Active Nodes": w.activeVert,
				//}

				outBytes, error := json.Marshal(w.activeVert)

				if error != nil {
					http.Error(rw, error.Error(), http.StatusInternalServerError)
					return
				}
				rw.Write(outBytes)
			}
		}
	}
}

// Run ...
func Run() {
	// TODO: gerald
	http.HandleFunc("/initConnection", initConnectionHandler)
	http.HandleFunc("/disseminateGraph", disseminateGraphHandler)
	http.HandleFunc("/startSuperstep", startSuperstepHandler)
	http.HandleFunc("/saveState", saveStateHandler)
	http.HandleFunc("/incomingMsg", workerToWorkerHandler)
	http.HandleFunc("/ping", pingHandler)
	http.ListenAndServe(":3000", nil)
}
