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

// Worker ...
type Worker struct {
	ID          int
	inQueue     map[int][]float64
	outQueue    map[int][]float64
	activeVert  []int                   //TODO: change type
	partToVert  map[int]map[int]*Vertex // partId: {verticeID: Vertex}
	udf         UDF
	graphReader graphReader
	superstep   int

	ctx            context.Context
	inQLock        sync.RWMutex
	outQLock       sync.RWMutex
	activeVertLock sync.RWMutex        // ensure that one partition access activeVert variable at a time
	pingPong       *semaphore.Weighted // flag: (A) whether superstep is completed; (B) whether initState is done
	busyWorker     *semaphore.Weighted // flag: Check if any goroutines are still handling incoming messages from peer workers
}

// NewWorker creates and returns an initialized worker
func NewWorker(udf UDF) *Worker {
	w := &Worker{udf: udf}
	w.Init()
	return w
}

// Init initializes a worker
func (w *Worker) Init() {
	w.ctx = context.Background()
	w.inQLock = sync.RWMutex{}
	w.outQLock = sync.RWMutex{}
	w.activeVertLock = sync.RWMutex{}              // ensure that one partition access activeVert variable at a time
	w.pingPong = semaphore.NewWeighted(int64(1))   // flag: (A) whether superstep is completed; (B) whether initState is done
	w.busyWorker = semaphore.NewWeighted(int64(1)) // flag: Check if any goroutines are still handling incoming messages from peer workers

	w.inQueue = make(map[int][]float64)
	w.outQueue = make(map[int][]float64)
	w.partToVert = make(map[int]map[int]*Vertex)
	w.superstep = 0
}

// loadVertices loads assigned vertices received from Master
func (w *Worker) initState(gr graphReader) {
	// create Vertices
	w.partToVert = make(map[int]map[int]*Vertex)

	if err := w.pingPong.Acquire(w.ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)
	}
	defer w.pingPong.Release(1)

	for vID, vReader := range gr.Vertices {
		partID := getPartition(vID, gr.Info.NumPartitions)
		v := Vertex{vID,
			vReader.Flag, //false = active
			vReader.Value,
			make([]float64, 0),
			make(chan []float64),
			make(map[int]float64),
			gr.Edges[vID]}

		if _, ok := w.partToVert[partID]; !ok {
			w.partToVert[partID] = make(map[int]*Vertex)
		}
		w.partToVert[partID][vID] = &v
		w.activeVert = gr.ActiveVerts
		w.outQueue = gr.outQueue
	}
	w.superstep = gr.superstep
	fmt.Println("Done loading, releasing pingpong.")
	printGraphReader(gr)
}

func (w *Worker) startSuperstep() {

	defer w.pingPong.Release(1)

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
				resultmsg := v.Compute(w.udf, w.superstep)
				fmt.Println("Finished computing for vertice: ", v.Id)
				w.processVertResult(resultmsg) // populate outQueue with return value of compute()
				fmt.Println("Populating out queue with computed value for vertex: ", v.Id)
			}
		}(vList)
	}
	wg.Wait()
	w.inQueue = make(map[int][]float64)

	fmt.Println("-----------------------------")
	fmt.Println("Ended all computations for vertices in partition. Disseminating msgs from outq")
	w.disseminateMsgFromOutQ() // send values to inqueue of respective worker nodes
	fmt.Println("Partition has finished superstep.")
	fmt.Println("-----------------------------")
	w.superstep++
}

func (w *Worker) disseminateMsgFromOutQ() {
	// this function is called during startSuperstep() when the worker is disseminating the vertice values

	nodeToOutQ := make(map[int]map[int][]float64)

	// iterate over the worker's outqueue and prepare to disseminate it to the correct destination vertexID
	fmt.Println("OutQueue: ", w.outQueue)
	for destVert, vals := range w.outQueue {

		w.outQLock.RLock()
		defer w.outQLock.RUnlock()

		partID := getPartition(destVert, w.graphReader.Info.NumPartitions)
		workerID := w.graphReader.PartitionToNode[partID]

		if _, ok := nodeToOutQ[workerID]; !ok {
			nodeToOutQ[workerID] = make(map[int][]float64)
			nodeToOutQ[workerID][destVert] = vals
		} else {
			nodeToOutQ[workerID][destVert] = append(nodeToOutQ[workerID][destVert], vals...)
		}
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
				w.inQLock.Lock()
				defer w.inQLock.Unlock()
				defer wg.Done()

				for vID := range outQ {
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
				fmt.Println("Sending to peer: ", outQ)
				fmt.Println("Stirng OutQBytes: ", string(outQBytes))
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
func (w *Worker) processVertResult(rm ResultMsg) {
	// Populate outQueue with outgoing messages
	w.outQLock.Lock()
	for dstVert, msg := range rm.msg {
		fmt.Println("Message: ", rm.msg)
		if msgList, ok := w.outQueue[dstVert]; ok {
			msgList = append(msgList, msg)
			w.outQueue[dstVert] = msgList
		} else {
			w.outQueue[dstVert] = []float64{msg}
		}
	}
	w.outQLock.Unlock()

	// Populate activeVert with vertices which are active at the end of superstep
	if rm.halt == false {
		w.activeVertLock.Lock()
		w.activeVert = append(w.activeVert, rm.sendId)
		w.activeVertLock.Unlock()
	}
}

func initConnectionHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "connected")
}

func (w *Worker) disseminateGraphHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "received")
	bodyBytes, err := ioutil.ReadAll(r.Body) // arr of bytes
	if err != nil {
		panic(err)
	}

	// get graph
	gr := getGraphFromJSONByte(bodyBytes)
	w.graphReader = *gr

	go w.initState(*gr)

	if err := w.pingPong.Acquire(w.ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v to load graph", err)
		fmt.Fprintf(rw, "NOT OK")
	} else {
		defer w.pingPong.Release(1)
		fmt.Println("Acquired Sempahore to load graph")
		//printGraphReader(*gr)
		fmt.Fprintln(rw, "ok")
	}
}

func (w *Worker) startSuperstepHandler(rw http.ResponseWriter, r *http.Request) {
	if err := w.pingPong.Acquire(w.ctx, 1); err != nil {
		log.Printf("Failed to acquire semaphore: %v to start superstep", err)
	} else {
		fmt.Println("Acquired semaphore to startedSuperstep")
		fmt.Println("Starting superstep.")
		go w.startSuperstep()
	}
}

func (w *Worker) workerToWorkerHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// do something
	}
	go func(bodyBytes []byte) {
		w.busyWorker.Acquire(w.ctx, 1)
		defer w.busyWorker.Release(1)

		fmt.Println("Receiving messages from peers")

		var dstToVals map[int][]float64

		json.Unmarshal(bodyBytes, &dstToVals)
		fmt.Println("String bodybytes: ", string(bodyBytes))

		fmt.Println("This is the stuff i received:, ", dstToVals)
		w.inQLock.Lock()
		defer w.inQLock.Unlock()
		for dst, vals := range dstToVals {
			fmt.Println("Receving values from peer: ", vals)
			w.inQueue[dst] = append(w.inQueue[dst], vals...)
		}
	}(bodyBytes)
}

func (w *Worker) saveStateHandler(rw http.ResponseWriter, r *http.Request) {
	gr := newGraphReader()

	fmt.Println("Worker in saveStateHandler.")

	gr.outQueue = w.outQueue

	for _, vert := range w.partToVert {
		for vID, v := range vert {
			vr := gr.Vertices[vID]
			vr.Value = v.Val
			vr.Flag = v.flag
			gr.Vertices[vID] = vr
		}
	}

	gr.superstep = w.superstep
	fmt.Println(gr.superstep)
	gr.ActiveVerts = w.activeVert
	fmt.Println(gr.ActiveVerts)
	fmt.Println("Sending checkpoint to master: ")
	printGraphReader(gr)

	bytes, _ := json.Marshal(&gr)
	fmt.Println(string(bytes), len(bytes))
	rw.Write(bytes)
	fmt.Println("Sending saved state to master.")

	// TODO: parse json to send to Master - graphReader and In/Out Queue

	// wait for the Master's the GET request - isit a ping????

	// define the format of the response

	// send back the response here - encoded as json or something

}

func (w *Worker) pingHandler(rw http.ResponseWriter, r *http.Request) {
	// read the ping request
	bodyByte, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyByte)
	fmt.Println("received ping")
	fmt.Println("bodystring: ", bodyString)

	if bodyString == "Completed graphHandler?" {
		if w.pingPong.TryAcquire(1) == false {
			log.Printf("Failed to acquire semaphore for graphHandler")
			fmt.Println("I'm not done with graphHandler.")
			fmt.Fprintf(rw, "still not done")
		} else {
			defer w.pingPong.Release(1)
			fmt.Println("I am done with graphHandler.")
			fmt.Fprintf(rw, "done")
		}
	} else if bodyString == "Completed Superstep?" {
		// lock when accessing masterResponse
		// if unable to access semaphore, send "still not done" to master
		if w.pingPong.TryAcquire(1) == false {
			log.Printf("Failed to acquire semaphore")
			fmt.Println("Master pinged, but I'm not done with my superstep :(")
			fmt.Fprintf(rw, "still not done")
		} else {
			fmt.Println("Acquired sempahore to signal superstep completed.")
			defer w.pingPong.Release(1)

			if w.busyWorker.TryAcquire(1) == false {
				fmt.Println("Acquire busyworker, still handling peer messages.")

			} else {
				defer w.busyWorker.Release(1)

				outBytes, _ := json.Marshal(w.activeVert)

				rw.Write(outBytes)
				fmt.Println("Outbytes: ", outBytes)
				fmt.Println("Sending active vertices to master: ", w.activeVert)
			}
		}
	}
}

func (w *Worker) terminateHandler(rw http.ResponseWriter, r *http.Request) {
	printGraphReader(w.graphReader)
	for _, vert := range w.partToVert {
		for _, v := range vert {
			fmt.Println("Value of vertex ", v.Id, ": ", v.Val)
		}
	}
}

// Run ...
func (w *Worker) Run() {
	// TODO: gerald
	http.HandleFunc("/initConnection", initConnectionHandler)
	http.HandleFunc("/disseminateGraph", w.disseminateGraphHandler)
	http.HandleFunc("/startSuperstep", w.startSuperstepHandler)
	http.HandleFunc("/saveState", w.saveStateHandler)
	http.HandleFunc("/incomingMsg", w.workerToWorkerHandler)
	http.HandleFunc("/ping", w.pingHandler)
	http.HandleFunc("/terminate", w.terminateHandler)
	http.ListenAndServe(":3000", nil)
}

// RunUDF creates a new worker with the given UDF and runs the worker
func RunUDF(udf UDF) {
	w := NewWorker(udf)
	w.Run()
}
