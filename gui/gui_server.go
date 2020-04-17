package main

import ("math/rand"
		"net/http"
		"encoding/json"
		"fmt"
		"os")
		// "time")


// the types of data needed by the master to disseminate into the graph are:

// static things
// 4. Number of Partitions 
// 6. total size of the graph


// dynamic things
// 1. line graph of the cost function for each vertice over time (with the time step)
// 2. histogram of the number of active nodes being processed (in each worker) 
// 3. Number of supersteps progressed so far
// 4. # of active nodes per partition being processed
// 9. Total Alive Time for Each Node (how many supersteps participated in)

// 7. timing for each superstep (tbd)
// 8. average timing across all supersteps (tbd)


// needs to be changed
type serverStats struct {
	doneSignal		 bool
	// static things
	graphFile 		 string
	numPartitions    int
	numVertices		 int

	// dynamic things
	nodeVertCostFn	 map[int]int	
	currentIteration int
	numActiveNodes	 int
	activeNodesVert  map[string][]int
	totalAliveTime	 map[string]int

	avgTiming		 map[string]float64
}

type httpReplyMsg struct {
	numPartitions    int
	currentIteration int
	activeNodes      []activeNode
	graphsToNodes    []graphReader
}

func initGUI(originalFile string, nodeAdrs map[string]bool) *serverStats {

	// instantiate the GUI
	graph := getGraphFromFile(originalFile)

	ipStr := make(map[string]int, 0)
	timing := make(map[string]float64, 0)
	activeNodesVert := make(map[string]int, 0)

	for ip, parts := range graph.ActiveNodes {
		ipStr[ip] = 0
		timing[ip] = 0.0
		activeNodesVert[ip] = []int
	}

	guistats := &serverStats{
		doneSignal:			false,
		graphFile: 			originalFile,
		numPartitions:		graph.Info.NumPartitions,
		numVertices:		graph.Info.NumVertices,

		nodeVertCostFn:		make(map[int]float64, 0),
		currentIteration:	0,
		numActiveNodes:		0,
		activeNodesVert:	activeNodesVert,
		totalAliveTime:		ipStr,
		avgTiming:			timing,
	}
	return guistats
}

func (guistats *serverStats) checkPath() *graphReader {
	if _, err := os.Stat(checkpointPATH); os.IsNotExist(err) {
		// the file does not exist. Read the very original graph
		graph := getGraphFromFile(guistats.graphFile)
	} else {
		// the file exists and read from checkpoint.json
		graph := getGraphFromFile(checkpointPATH)
	}

	return graph
}

func contains(s []int, e int) bool {
    for _, a := range s {
        if a.IP == e {
            return true
        }
    }
    return false
}

func (guistats *serverStats) sendGraphStats (w http.ResponseWriter, request *http.Request) {
	
	graph := checkPath()

	guistats.currentIteration = graph.Superstep

	// append the cost function for each node at each superstep to nodeVertCostFn
	for k,v := range graph.Vertices {
		guistats.nodeVertCostFn[k] = v.Value
	}

	for ip, activeVert := range guistats.activeNodesVert {
		guistats.activeNodesVert[ip] = append(guistats.activeNodesVert[ip], 0)
	}

    if len(graph.ActiveNodes) > 0 {

		guistats.numActiveNodes = len(graph.ActiveNodes)

		// get the total length of active currently active vertices for each partition
		for ip, activeVert := range guistats.activeNodesVert {
			if contains(graph.ActiveNodes, ip) == true {
				guistats.activeNodesVert[ip] = append(guistats.activeNodesVert[ip], len(partitionList))
			}
		}
		
		// update total alive time
		for ip, step := range guistats.totalAliveTime {

			// if the ip address of a node is contained in ActiveNodes, then change the timing
			if contains(graph.ActiveNodes, ip) == true {
				guistats.totalAliveTime[item.IP] = graph.Superstep
			}
		} 

	    // append the average timing
	    for idx, timing := range guistats.totalAliveTime {

    		total := 0  
    		for _, number := range numbs {  
    			total = total + number  
    		}
    		avg := total / len(timing)

    		guistats.avgTiming[idx] = append(guistats.avgTiming[idx], avg)
    	}


		log.Printf("current Status is: %v", guistats)


    } else {
    	guistats.doneSignal = true
    }


	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(guistats); err != nil {
        panic(err)
    }

 	// req, _ := http.NewRequest("POST", getURL(ip, "3000", "guiserver"), bytes.NewBuffer([]byte msg)
}

func (guistats *serverStats) terminateHandler(rw http.ResponseWriter, r *http.Request) {
	// handles the terminate handler from master
	printGraphReader(w.graphReader)
	for _, vert := range w.partToVert {
		for _, v := range vert {
			fmt.Println("Value of vertex ", v.Id, ": ", v.Val)
		}
	}
}


func getPortPath(path, port string) string {
	return strings.TrimSpace(path) + "/" + strings.TrimSpace(port)
}

func runServer(server string) {
	log.Printf("GUI Server running from port %s", server)
	http.HandleFunc(getPortPath("/guiserver", port), sendGraphStats)
	http.HandleFunc(getPortPath("/terminate", port), sendGraphStats)
	http.ListenAndServe(fmt.Sprintf(":%s"server), nil)
}

func RunGUI(server string, originalFile string, ip nodeAdrs) {
	guistats := initGUI(originalFile, ip)
	guistats.runServer(server, guistats)
	// for {
	// }
}
