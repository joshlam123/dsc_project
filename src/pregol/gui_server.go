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


const savePATH = "../results/guiSave.json"

type serverStats struct {
	GraphName 		 string
	DoneSignal		 int
	// static things
	GraphFile 		 string
	NumPartitions    int
	NumVertices		 int

	// dynamic things
	nodeAddresses 	 map[int]string
	InactiveNodes 	 map[int]string
	ActiveNodes 	 map[int]string
	PartitionList	 map[int]map[string][]int
	NodeVertCostFn	 map[int]map[int]float64
	CurrentIteration int
	ActiveNodesVert  map[int]map[int][]int
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
	PartitionList	 map[string][]int
	NodeVertCostFn	 map[int]map[int]float64
	InactiveNodes 	 map[int]string
	CurrentIteration int
	NumActiveNodes	 int
	ActiveNodesVert  map[int]map[int][]int
	TotalAliveTime	 map[string]int
	AvgTiming		 []float64
}

type httpReplyMsg struct {
	numPartitions    int
	currentIteration int
	activeNodes      []activeNode
	graphsToNodes    []graphReader
}

type guiSave struct {
	CurrentIteration int
	GraphsToNodes    []graphReader
	NodeAdrs 		 map[string]bool
}

func getSaveFile(graphFile string) *guiSave {
	jsonFile, err := os.Open(graphFile)

	if err != nil {
		panic(err)
	}

	defer jsonFile.Close()

	byteVal, _ := ioutil.ReadAll(jsonFile)
	var g guiSave
	json.Unmarshal(byteVal, &g)
	return &g
}

func initGUI(originalFile string, nodeAdrs []string, graphName string) *serverStats {

	// instantiate the GUI
	graph := getGraphFromFile(originalFile)

	guistats := &serverStats{
		GraphName: 			graphName,
		DoneSignal:			0,
		GraphFile: 			originalFile,
		NumPartitions:		0,
		NumVertices:		len(graph.Vertices),

		PartitionList: 		make(map[int]map[string][]int),
		nodeAddresses: 		make(map[int]string),
		NodeVertCostFn:		make(map[int]map[int]float64, 0),
		CurrentIteration:	1,
		ActiveNodes:		make(map[int]string, 0),
		InactiveNodes: 		make(map[int]string, 0),
		ActiveNodesVert:	make(map[int]map[int][]int),
		TotalAliveTime:		make(map[string]int),
		AvgTiming:			make([]float64,0),
		Mux: 				sync.RWMutex{},

	}
	log.Printf("GUI Initialised! Current state of GUI at %d iterations", guistats.CurrentIteration)
	return guistats
}

func (guistats *serverStats) checkPath() *guiSave {

	var graph *guiSave

	if _, err := os.Stat(savePATH); os.IsNotExist(err) {

		// the file exists and read from checkpoint.json
		fmt.Println("Save file has not yet been created.")

	} else {

		// the file does not exist. Read the very original graph
		graph = getSaveFile(savePATH)

		guistats.NumPartitions = graph.GraphsToNodes[0].Info.NumPartitions

		if graph.CurrentIteration == 0 {

			for k,v := range graph.GraphsToNodes {
				guistats.nodeAddresses[v.Info.NodeID] = v.ActiveNodes[k].IP
				guistats.ActiveNodes[v.Info.NodeID] = v.ActiveNodes[k].IP
				guistats.TotalAliveTime[v.ActiveNodes[k].IP] = 1
			}

		} 

		log.Printf("Node Adr, %v", guistats.nodeAddresses)

		guistats.PartitionList[graph.CurrentIteration] = make(map[string][]int)
		for k, v := range graph.GraphsToNodes[0].PartitionToNode {
			log.Printf("Partition List %s", guistats.nodeAddresses[v])
			guistats.PartitionList[graph.CurrentIteration][guistats.nodeAddresses[v]] = append(guistats.PartitionList[graph.CurrentIteration][guistats.nodeAddresses[v]], k)
		}

		log.Printf("Read graph from %s: %v", savePATH, guistats.CurrentIteration)

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

func getIndex(address map[int]string, searchString string) int {
	var correctIndex int = -1
	for k,v := range address {
		if v == searchString {
			correctIndex = k
		}
	}
	return correctIndex
}

func (guistats *serverStats) sendGraphStats (w http.ResponseWriter, request *http.Request) {

		graph := guistats.checkPath()	

		if graph == nil  {

		    fmt.Println("is zero value")

		} else {

			guistats.Mux.Lock()

			// update total alive time
			for k, v := range graph.GraphsToNodes {
				if len(v.ActiveVerts) > 0 && guistats.CurrentIteration != graph.CurrentIteration {
					guistats.TotalAliveTime[v.ActiveNodes[k].IP] += 1
				}
			} 

			total := 0
		    // append the average timing
		    for _, timing := range guistats.TotalAliveTime {
				total = total + timing
	    	}

	    	fmt.Println("Total TIMING ", total)

	    	// function to calculate the average timing
	    	newtotal := float64(total)
	    	avg := newtotal / float64(len(guistats.TotalAliveTime))
	    	guistats.AvgTiming = append(guistats.AvgTiming, avg)

			log.Printf("current Timing is: %v", guistats.AvgTiming)

			log.Printf("Total Alive Time: %v", guistats.TotalAliveTime)

			// set the current iteration to the current superstep
			guistats.CurrentIteration = graph.CurrentIteration
				
			// initialise activenodesvert and nodevertcostfn for this superstep
			guistats.ActiveNodesVert[graph.CurrentIteration] = make(map[int][]int)
			guistats.NodeVertCostFn[graph.CurrentIteration] = make(map[int]float64)

			// append the cost function for each node at each superstep to nodeVertCostFn
			for k, v := range graph.GraphsToNodes[0].Vertices {
				guistats.NodeVertCostFn[guistats.CurrentIteration][k] = v.Value
			}	

			// number of nodes in the outqueue
			var nOQ int = 0 
			for _, v := range graph.GraphsToNodes {
				if len(v.OutQueue) == 0 {
					nOQ += 1
				}
			}

		    if len(guistats.nodeAddresses) == nOQ {

		    	// remove the no longer active nodes from active nodes
		    	for k,v := range graph.NodeAdrs {
		    		if v == true {
		    			index := getIndex(guistats.nodeAddresses, k)
		    			if index != -1 {
		    				guistats.ActiveNodes[index] = k
		    				if _, ok := guistats.InactiveNodes[index]; ok {
		    					delete(guistats.InactiveNodes, index)
		    				}
		    			} 
		    		} else {
		    			index := getIndex(guistats.nodeAddresses, k)
		    			if index != -1 {
		    				delete(guistats.ActiveNodes, index)
		    				guistats.InactiveNodes[index] = k
		    			}  

		    		}
		    	}


		    	// get the total length of active currently active vertices for each partition
				for _, activenode := range graph.GraphsToNodes {

					for _, v := range activenode.ActiveVerts {
						guistats.ActiveNodesVert[graph.CurrentIteration][activenode.Info.NodeID] = append(guistats.ActiveNodesVert[graph.CurrentIteration][activenode.Info.NodeID], v)
					}

				}

			
		    } else {

		    	// pregel is done
		    	guistats.DoneSignal = 1

		    }

		    // package the data to be sent to the GUI in the serverdata struct
			data := serverData{GraphName:guistats.GraphName, DoneSignal:guistats.DoneSignal, NumPartitions:guistats.NumPartitions,
								NumVertices:guistats.NumVertices, NodeVertCostFn:guistats.NodeVertCostFn, PartitionList: guistats.PartitionList[graph.CurrentIteration], 
								CurrentIteration:guistats.CurrentIteration, NumActiveNodes:len(guistats.ActiveNodes), 
								ActiveNodesVert:guistats.ActiveNodesVert, TotalAliveTime:guistats.TotalAliveTime, InactiveNodes: guistats.InactiveNodes,
								AvgTiming:guistats.AvgTiming}


			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
		    
		    json.NewEncoder(w).Encode(data)

		    guistats.Mux.Unlock()

		    log.Printf("Wrote data")

		}

	delay()
}


// handler for running the server
func (guistats *serverStats) runServer(server string) {
	http.HandleFunc("/guiserver", guistats.sendGraphStats)
	http.ListenAndServe(fmt.Sprint(":",server), nil)
	log.Printf("GUI Server running from port %s", server)
}

// initialiser that will be called from runGUI inside the gui folder
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
