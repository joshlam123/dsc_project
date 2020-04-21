package pregol

import ("net/http"
		"encoding/json"
		"fmt"
		"io/ioutil"
		"strings"
		"log"
		"time"
		"math/rand"
		"sync"
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
	GraphName 		 string
	DoneSignal		 int
	// static things
	GraphFile 		 string
	NumPartitions    int
	NumVertices		 int

	// dynamic things
	NodeVertCostFn	 map[int]map[int]float64
	CurrentIteration int
	NumActiveNodes	 []int
	ActiveNodesVert  map[string][]int
	TotalAliveTime	 map[string]int
	AvgTiming		 []float64
	Mux 			 sync.RWMutex

}

type serverData struct {
	GraphName 		 string
	DoneSignal		 int
	// static things
	NumPartitions    int
	NumVertices		 int

	// dynamic things
	NodeVertCostFn	 map[int]map[int]float64	
	CurrentIteration int
	NumActiveNodes	 []int
	ActiveNodesVert  map[string][]int
	TotalAliveTime	 map[string]int
	AvgTiming		 []float64

}

type httpReplyMsg struct {
	numPartitions    int
	currentIteration int
	activeNodes      []activeNode
	graphsToNodes    []graphReader
}

func initGUI(originalFile string, nodeAdrs []string, graphName string) *serverStats {

	// instantiate the GUI
	graph := getGraphFromFile(originalFile)

	timing := make(map[string]int, 0)
	activeNodesVert := make(map[string][]int, 0)

	for _, nodes := range nodeAdrs {
		timing[strings.TrimSpace(nodes)] = 0
		activeNodesVert[strings.TrimSpace(nodes)] = make([]int, 0)
	}

	guistats := &serverStats{
		GraphName: 			graphName,
		DoneSignal:			0,
		GraphFile: 			originalFile,
		NumPartitions:		graph.Info.NumPartitions,
		NumVertices:		len(graph.Vertices),

		NodeVertCostFn:		make(map[int]map[int]float64, 0),
		CurrentIteration:	0,
		NumActiveNodes:		make([]int,0),
		ActiveNodesVert:	activeNodesVert,
		TotalAliveTime:		timing,
		AvgTiming:			make([]float64,0),
		Mux: 				sync.RWMutex{},

	}
	log.Printf("GUI Initialised! Current state of GUI at %d iterations", guistats.CurrentIteration)
	return guistats
}

func (guistats *serverStats) checkPath() *graphReader {
	var graph *graphReader

	if _, err := os.Stat(checkpointPATH); os.IsNotExist(err) {
		// the file does not exist. Read the very original graph
		graph = getGraphFromFile(guistats.GraphFile)
	} else {
		// the file exists and read from checkpoint.json
		graph = getGraphFromFile(checkpointPATH)
	}

	return graph
}

func contains(s []activeNode, e string) bool {
    for _, a := range s {
        if a.IP == e {
            return true
        }
    }
    return false
}

func delay() int {
  // random delay for each message
	randomAmt := rand.Intn(5000)
	amt := time.Duration(randomAmt)
	time.Sleep(time.Millisecond * amt)
	return randomAmt/1000
}

func (guistats *serverStats) sendGraphStats (w http.ResponseWriter, request *http.Request) {

		graph := guistats.checkPath()
		log.Printf("Read graph from %s: %v", guistats.GraphFile, guistats.CurrentIteration)

		// guistats.CurrentIteration = graph.Superstep
		guistats.CurrentIteration = guistats.CurrentIteration + graph.Superstep + 1
		guistats.NodeVertCostFn[guistats.CurrentIteration] = make(map[int]float64)
		// append the cost function for each node at each superstep to nodeVertCostFn

		guistats.Mux.Lock()

		for k,_ := range graph.Vertices {
			// only for testing
			// guistats.NodeVertCostFn[k] = v.Value
			r := 0 + rand.Float64() * (10 - 0)
			guistats.NodeVertCostFn[guistats.CurrentIteration][k] = r
		}	

		for ip, _ := range guistats.ActiveNodesVert {
			guistats.ActiveNodesVert[ip] = append(guistats.ActiveNodesVert[ip], 0)
		}

	    if len(graph.ActiveNodes) == 0 {

			guistats.NumActiveNodes = append(guistats.NumActiveNodes, len(graph.ActiveNodes))

			for _, activenode := range graph.ActiveNodes {
				// get the total length of active currently active vertices for each partition
				for ip, _ := range guistats.ActiveNodesVert {
					if activenode.IP == ip {
						guistats.ActiveNodesVert[ip] = append(guistats.ActiveNodesVert[ip], len(activenode.PartitionList))
					}
				}
			}
			
			// only for testing purposes
			randomMap := make(map[string]int)
			for ip,_ := range guistats.TotalAliveTime {
				// r := 0 + rand.Float64() * (10 - 0)
				r := rand.Intn(10)
				randomMap[ip] = r
			}
			// remember to delete later
			
			// update total alive time
			for ip, _ := range guistats.TotalAliveTime {

				// if the ip address of a node is contained in ActiveNodes, then change the timing
				// if contains(graph.ActiveNodes, ip) == true {
				// 	guistats.TotalAliveTime[ip] = float64(graph.Superstep)
				// }
				guistats.TotalAliveTime[ip] = randomMap[ip]
			} 


			total := 0
		    // append the average timing
		    for _, timing := range guistats.TotalAliveTime {
				total = total + timing
	    	}
	    	newtotal := float64(total)
	    	avg := newtotal / float64(len(guistats.TotalAliveTime))
	    	guistats.AvgTiming = append(guistats.AvgTiming, avg)

			log.Printf("current Timing is: %v", guistats.AvgTiming)
		
			guistats.DoneSignal = 1
	    } else {
	    	guistats.DoneSignal = 1
	    }


	 //    guidata, err := json.Marshal(guistats)
		
		// if err != nil {
		// 	fmt.Fprintf(w, "Error: %s", err)
		// }

		// // log.Printf("data: %v", guidata)

		data := serverData{GraphName:guistats.GraphName, DoneSignal:guistats.DoneSignal, NumPartitions:guistats.NumPartitions,
							NumVertices:guistats.NumVertices, NodeVertCostFn:guistats.NodeVertCostFn, CurrentIteration:guistats.CurrentIteration,
							NumActiveNodes:guistats.NumActiveNodes, ActiveNodesVert:guistats.ActiveNodesVert, TotalAliveTime:guistats.TotalAliveTime,
							AvgTiming:guistats.AvgTiming}


		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
	    // w.WriteHeader(http.StatusOK)
	    // fmt.Fprintf(w, "%v", guidata)
	    json.NewEncoder(w).Encode(data)

	    guistats.Mux.Unlock()

	    log.Printf("Wrote data")
	    // if err := json.NewEncoder(w).Encode(guistats); err != nil {
	    //     panic(err)
	    // } 

	delay()
 	// req, _ := http.NewRequest("POST", getURL(ip, "3000", "guiserver"), bytes.NewBuffer([]byte msg)
}


func (guistats *serverStats) runServer(server string) {
	http.HandleFunc("/guiserver", guistats.sendGraphStats)
	http.ListenAndServe(fmt.Sprint(":",server), nil)
	log.Printf("GUI Server running from port %s", server)
}

func RunGUI(server string, originalFile string, ip string, graphName string) {

	ipaddrs := make([]string,0)

	dat, err := ioutil.ReadFile(ip)
	if err != nil {
		panic(err)
	}

	for _, ip := range strings.Split(string(dat), "\n") {
		ipaddrs = append(ipaddrs, ip)
	}

	guistats := initGUI(originalFile, ipaddrs, graphName)
	guistats.runServer(server)

	var input string
	fmt.Scanln(&input)
	fmt.Println("Quitting Application")
	// for {
	// }
}
