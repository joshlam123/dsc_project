package main

import ("math/rand"
		"net/http"
		"encoding/json"
		"fmt")
		// "time")


// the types of data needed by the master to disseminate into the graph are:

// static things
// 4. Number of Partitions 
// 6. total size of the graph


// dynamic things
// 1. line graph of the cost function for each vertice over time (with the time step)
// 2. histogram of the number of active nodes being processed (in each worker) 
// 3. Number of supersteps progressed so far
// 4. # of active nodes being processed
// 9. Total Alive Time for Each Node (how many supersteps participated in)

// 7. timing for each superstep (tbd)
// 8. average timing across all supersteps (tbd)


// needs to be changed
type serverStats struct {
	// static things
	numPartitions    int
	numNodes		 int
	numEdges		 int

	// dynamic things
	nodeVertCostFn	 map[int]map[int]int	
	currentIteration int
	numActiveNodes	 int
	activeNodesVert  map[int]int
	totalAliveTime	 map[int]int

	timing			 map[int][]int
	avgTiming		 map[int]float64
}

type httpReplyMsg struct {
	numPartitions    int
	currentIteration int
	activeNodes      []activeNode
	graphsToNodes    []graphReader
}

func initGUI() *serverStats {
	guistats := &serverStats{
		numPartitions:		0,
		numNodes:			0,
		numEdges:			0,
		nodeVertCostFn:		make(map[int]map[int]int, 0),
		currentIteration:	0,
		numActiveNodes:		0,
		activeNodesVert:	make(map[int]int, 0),
		totalAliveTime:		make(map[int]int, 0),
		timing:				make(map[int][]int, 0),
		avgTiming:			make(map[int]float64, 0),
	}
	return guistats
}

func (guistats *serverStats) sendGraphStats (w http.ResponseWriter, request *http.Request) {
	jsonFile, err := os.Open("../../run_master/users.json")
	// if we os.Open returns an error then handle it
	if err != nil {
	    fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	graphBody, err := ioutil.ReadAll(r.Body) // arr of bytes
	if err != nil {
		panic(err)
	}

	fmt.Println("Received %s from master: ", string(graphBody))

	// change this to a struct
	var graphVals httpReplyMsg
	json.Unmarshal(graphBody, &graphVals)

	if err != nil {
		// this means that it received nothing from the post request
		// there must be a better way to handle this
        panic(err)
    }

    // if len(graphVals.activeNodes) > 0 {

    // }

    // TODO:	
    guistats.numPartitions = 0
    guistats.numNodes = 0
    gui.numEdges = 0


	// get graph
	// mRecv := getGraphFromJSONByte(graphVals)

	// this is what graphstoNodes looks like
	// for i := range m.graphsToNodes {
	// 	m.graphsToNodes[i] = newGraphReader()
	// 	m.graphsToNodes[i].Info = gOriginal.Info
	// 	m.graphsToNodes[i].Info.NodeID = i
	// 	m.graphsToNodes[i].Info.NumPartitions = m.numPartitions
	// 	m.graphsToNodes[i].ActiveNodes = m.activeNodes
	// 	m.graphsToNodes[i].PartitionToNode = g.PartitionToNode
	// 	m.graphsToNodes[i].Superstep = m.currentIteration
	// }

    for k,v := range gr.activeNodes {
    	totalAliveTime[]
    }


    var avgTiming 

	response := T{numPartitions:graphVals.numPartitions, numNodes: , numEdges: , numActiveNodes: , activeNodesVert: , 
			nodeVertCostFn: , totalAliveTime:, timing: 0.00, avgTiming: 3.14152689}

	log.Printf("Response is: %v", response)

	// w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// w.Header().Set("Access-Control-Allow-Origin", "*")
 //    w.WriteHeader(http.StatusOK)
 //    if err := json.NewEncoder(w).Encode(response); err != nil {
 //        panic(err)
 //    }
}

func getPortPath(path, port string) string {
	return strings.TrimSpace(path) + "/" + strings.TrimSpace(port)
}

func runServer(server string) {
	log.Printf("GUI Server running from port %s", server)
	http.HandleFunc(getPortPath("/guiserver", port), sendGraphStats)
	http.ListenAndServe(fmt.Sprintf(":%s"server), nil)
}

func RunGUI(server string) {
	guistats := initGUI()
	runServer(server, guistats)
	// for {
	// }
}
