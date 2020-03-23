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
var busyWorker = semaphore.NewWeighted(int64(1)) // flag: Check if any goroutines are still handling incoming messages from peer workers
var superstep = 0

// Worker ...
type Worker struct {
	ID          int
	inQueue     map[int][]float64
	outQueue    map[int][]float64
	activeVert  []int                   //TODO: change type
	partToVert  map[int]map[int]*Vertex // partId: {verticeID: Vertex}
	udf         UDF
	graphReader graphReader
}

func init() {
	w.inQueue = make(map[int][]float64)
	w.outQueue = make(map[int][]float64)
	w.partToVert = make(map[int]map[int]*Vertex)
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

		if _, ok := w.partToVert[partID]; !ok {
			w.partToVert[partID] = make(map[int]*Vertex)
		}
		w.partToVert[partID][vID] = &v
	}
	fmt.Println("Done loading, releasing pingpong.")
	printGraphReader(gr)
}

func startSuperstep() {

	defer pingPong.Release(1)

	// sending values to vertices through InMsg Channel
	fmt.Println("inqueue: ", w.inQueue)
	for vID, val := range w.inQueue {
		partID := getPartition(vID, w.graphReader.Info.NumPartitions)
		fmt.Println("new val: ", val)
		if _, ok := w.partToVert[partID][vID]; !ok {
			fmt.Println("not ok")
		}

		w.partToVert[partID][vID].SetInEdge(val)
	}

	// clearing queues so new values are not appended to old values (refresh for fresh superStep)
	// same for activeVert
	w.outQueue = make(map[int][]float64)
	w.activeVert = make([]int, 0)

	var wg sync.WaitGroup
	for _, vList := range w.partToVert {
		wg.Add(1)                        // add waitGroup for each partition: vertex list
		go func(vList map[int]*Vertex) { // for each partition, launch go routine to call compute for each of its vertex
			defer wg.Done()
			for _, v := range vList { // for each vertex in partition, compute().
				fmt.Println("Computing for vertice: ", v.Id)
				resultmsg := v.Compute(w.udf, superstep) //TODO: get superstep number
				fmt.Println("Finished computing for vertice: ", v.Id)
				processVertResult(resultmsg) //populate outQueue with return value of compute()
				fmt.Println("Populating out queue with computed value for vertex: ", v.Id)
			}
		}(vList)
	}
	wg.Wait()
	w.inQueue = make(map[int][]float64)

	fmt.Println("-----------------------------")
	fmt.Println("Ended all computations for vertices in partition. Disseminating msgs from outq")
	disseminateMsgFromOutQ() // send values to inqueue of respective worker nodes
	fmt.Println("Partition has finished superstep.")
	fmt.Println("-----------------------------")
	superstep++
}

func disseminateMsgFromOutQ() {
	// this function is called during startSuperstep() when the worker is disseminating the vertice values

	nodeToOutQ := make(map[int]map[int][]float64)

	// iterate over the worker's outqueue and prepare to disseminate it to the correct destination vertexID
	fmt.Println("OutQueue: ", w.outQueue)
	for m, n := range w.outQueue {

		outQLock.RLock()
		defer outQLock.RUnlock()

		partID := getPartition(m, w.graphReader.Info.NumPartitions)
		workerID := w.graphReader.PartitionToNode[partID]

		if _, ok := nodeToOutQ[workerID]; !ok {
			nodeToOutQ[workerID] = make(map[int][]float64)
		}
		nodeToOutQ[workerID][m] = n
	}
	fmt.Println("Finished building nodeToOutQ")

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
				defer wg.Done()

				for vID := range outQ {
					//w.inQueue[vID] = append(w.inQueue[vID], outQ[vID]...)
					for i := range outQ[vID] {
						w.inQueue[vID] = append(w.inQueue[vID], outQ[vID][i])
					}
					fmt.Println("Populating own inQ.")
				}

			}(nodeID, outQ)

		} else {
			workerIP := w.graphReader.ActiveNodes[nodeID].IP
			outQBytes, _ := json.Marshal(outQ)

			// send values to correct worker
			go func(workerIP string, outQBytes []byte) {
				defer wg.Done()
				fmt.Println("Sending InQ values to worker via json post")

				c := &http.Client{}
				//_, err := http.NewRequest("POST", "http://"+workerIP+":3000/incomingMsg", bytes.NewBuffer(outQBytes))
				fmt.Println("Sending to peer: ", outQ)
				req, err := http.NewRequest("POST", getURL(workerIP, "3000", "incomingMsg"), bytes.NewBuffer(outQBytes))
				if err != nil {
					log.Fatalln(err)
				}
				c.Do(req)
			}(workerIP, outQBytes)
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
		fmt.Println("Message: ", rm.msg)
		if msgList, ok := w.outQueue[dstVert]; ok {
			msgList = append(msgList, msg)
			w.outQueue[dstVert] = msgList
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
	fmt.Fprintln(rw, "received")
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
		log.Printf("Failed to acquire semaphore: %v to load graph", err)
		fmt.Fprintf(rw, "NOT OK")
	} else {
		defer pingPong.Release(1)
		fmt.Println("Acquired Sempahore to load graph")
		printGraphReader(*gr)
		fmt.Fprintln(rw, "ok")
	}
}

func startSuperstepHandler(rw http.ResponseWriter, r *http.Request) {
	if err := pingPong.Acquire(ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v to start superstep", err)
	} else {
		fmt.Println("Acquired semaphore to startedSuperstep")
		fmt.Println("Starting superstep.")
		go startSuperstep()
	}
}

func workerToWorkerHandler(rw http.ResponseWriter, r *http.Request) {
	// map[int][]float64
	go func() {
		busyWorker.Acquire(ctx, 1)
		defer busyWorker.Release(1)

		fmt.Println("Receiving messages from peers")
		fmt.Fprintf(rw, "Start receive from peers")
		defer r.Body.Close()
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			// do something
		}
		var dstToVals map[int][]float64
		json.Unmarshal(bodyBytes, &dstToVals)

		inQLock.Lock()
		defer inQLock.Unlock()
		for dst, vals := range dstToVals {
			fmt.Println("Receving values from peer: ", vals)
			w.inQueue[dst] = append(w.inQueue[dst], vals...)
		}
	}()
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
	fmt.Println("received ping")
	fmt.Println("bodystring: ", bodyString)

	if bodyString == "Completed graphHandler?" {
		if pingPong.TryAcquire(1) == false {
			log.Printf("Failed to acquire semaphore for graphHandler")
			//rw.Write([]byte("Still not done"))
			fmt.Println("I'm not done with graphHandler.")
			fmt.Fprintf(rw, "still not done")
		} else {
			defer pingPong.Release(1)
			fmt.Println("I am done with graphHandler.")
			fmt.Fprintf(rw, "done")
		}
	} else if bodyString == "Completed Superstep?" {
		// lock when accessing masterResponse
		// if unable to access semaphore, send "still not done" to master
		if pingPong.TryAcquire(1) == false {
			log.Printf("Failed to acquire semaphore")
			fmt.Println("Master pinged, but I'm not done with my superstep :(")
			//rw.Write([]byte("Still not done"))
			fmt.Fprintf(rw, "still not done")
		} else {
			fmt.Println("Acquired sempahore to signal superstep completed.")
			defer pingPong.Release(1)

			if busyWorker.TryAcquire(1) == false {
				fmt.Println("Acquire busyworker, still handling peer messages.")

			} else {
				//resp := map[string][]int{
				//	"Active Nodes": w.activeVert,
				//}
				defer busyWorker.Release(1)

				outBytes, _ := json.Marshal(w.activeVert)

				rw.Write(outBytes)
				fmt.Println("Outbytes: ", outBytes)
				fmt.Println("Sending active vertices to master: ", w.activeVert)
			}
		}
	}
}

func terminateHandler(rw http.ResponseWriter, r *http.Request) {
	printGraphReader(w.graphReader)
	for _, vert := range w.partToVert {
		for _, v := range vert {
			fmt.Println("Value of vertex ", v.Id, ": ", v.Val)
		}
	}
	//fmt.Println("Max Value: ", w.partToVert[1][0].Val)
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
	http.HandleFunc("/terminate", terminateHandler)
	http.ListenAndServe(":3000", nil)
}
