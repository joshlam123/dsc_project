package main

import ("math/rand"
		"net/http"
		"encoding/json"
		"fmt")


// the types of data needed by the master to disseminate into the graph are:

// 1. histogram of the number of outdegrees of each graph
// 2. histogram of the number of active vertices being processed (in each worker)
// 3. progress of computation 
// 4. current computed cost function (value of user-defined function)
// 5. # of active nodes being processed
// 6. total size of the graph


type T struct {
	// maybe need to add dead nodes?? 
	numNodes		map[int]int
	numPartitions	int
	activeVertices 	[]activeNode
	computeProgress float64
	currentSuperStep int
	costFn 			float64
	activeNodes 	float64
	graphSize		float64
}

// might not use later
// func processMasterData(mRecv guiSend) (T) {
// 	// activeVertices is a property of master
// }

func sendGraphStats (w http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	var mRecv Master
	err := decoder.Decode(&mRecv)

	if err != nil {
		// this means that it received nothing from the post request
		// there must be a better way to handle this
        panic(err)
    }

	// response := processMasterData(mRecv)
	response := T{numNodes:len(mRecv.m.nodeAdrs), numPartitions:mRecv.m.numPartitions, activeVertices:m.activeVertices, 
			computeProgress: ?????, currentSuperStep: mRecv.iter, costFn: ?????, activeNodes: mRecv.m.activeNodes,
			 graphSize:}
	// fmt.Println(response)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        panic(err)
    }
}


func main() {
	http.HandleFunc("/guiserver", sendGraphStats)
	http.ListenAndServe(":3000", nil)
}
