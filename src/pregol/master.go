package pregol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const checkpointPATH = "checkpoint.json"

type masterState int

const (
	SUPERSTEP masterState = iota
	DEAD
	DONE
	SAVESTATE
)

// Master ...
type Master struct {
	numPartitions    int             // Number of partitions
	checkpoint       int             // Number of supersteps before reaching a checkpoint
	nodeAdrs         map[string]bool //
	activeNodes      []activeNode
	graphsToNodes    []graphReader
	graphFile        string //
	client           *http.Client
	currentIteration int
	port             string
	primaryAddress   string
}

// NewMaster Constructor for Master struct
func NewMaster(numPartitions, checkpoint int, ipFile, graphFile string, port string, primaryAddress string) *Master {
	m := Master{}
	m.numPartitions = numPartitions
	m.checkpoint = checkpoint
	m.activeNodes = make([]activeNode, 0)
	m.graphFile = graphFile
	m.client = &http.Client{
		Timeout: time.Second * 10,
	}
	m.currentIteration = 0
	m.port = port
	m.primaryAddress = primaryAddress
	dat, err := ioutil.ReadFile(ipFile)
	if err != nil {
		panic(err)
	}
	m.nodeAdrs = make(map[string]bool)
	for _, ip := range strings.Split(string(dat), "\n") {
		m.nodeAdrs[strings.TrimSpace(ip)] = false
	}

	return &m
}

// InitConnections Initializes connections with all machines via ip addresses found in nodeAdrs []string.
// Generates a list of activeNodes which are the machines that Master will request to start Superstep.
func (m *Master) InitConnections() {

	var wg sync.WaitGroup
	activeNodeChan := make(chan activeNode, len(m.nodeAdrs))
	fmt.Println("Initiating connection with:")
	for ip := range m.nodeAdrs {
		wg.Add(1)

		// Initialize connection with all nodes, and check if they are active using GET request
		go func(ip string, wg *sync.WaitGroup, activeNodeChan chan activeNode) {
			defer wg.Done()
			fmt.Println(getURL(ip, "initConnection"))
			resp, err := m.client.Get(getURL(ip, "initConnection"))
			if err != nil {
				return
			}

			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Println("Machine", ip, "connected")
				activeNodeChan <- activeNode{ip, make([]int, 0)}
			}
		}(ip, &wg, activeNodeChan)

	}
	wg.Wait()
	close(activeNodeChan)

	m.activeNodes = make([]activeNode, 0)
	for ip := range m.nodeAdrs {
		m.nodeAdrs[ip] = false
	}
	for elem := range activeNodeChan {
		m.activeNodes = append(m.activeNodes, elem)
		m.nodeAdrs[elem.IP] = true
	}

	if len(m.activeNodes) == 0 {
		panic(errors.New("no active nodes"))
	}
}

func (m *Master) savePartStatus(){
	currentCheckpoint := m.currentIteration - (m.currentIteration % m.checkpoint)
	data := guiSave{CurrentIteration:currentCheckpoint, GraphsToNodes:m.graphsToNodes, NodeAdrs:m.nodeAdrs}
	file, _ := json.MarshalIndent(data, "", " ")
	ioutil.WriteFile("../gui/guiSave.json", file, 0644)
	fmt.Println("Gui Save written to file.")
}

// AssignPartitions Assign partToVert to active nodes
func (m *Master) AssignPartitions(graphFile string) {
	gOriginal := getGraphFromFile(m.graphFile)                // Original file
	g := getGraphFromFile(graphFile)                          // Read graphfile (could be original or checkpoint)
	g.PartitionToNode = make(map[int]int)                     // partition id to node id (to be inserted in each graphToNodes)
	m.graphsToNodes = make([]graphReader, len(m.activeNodes)) // 1 graph to send to each node

	// Loop through each partition
	// Assign partitions to nodes
	for i := 0; i < m.numPartitions; i++ {
		cNode := i % len(m.activeNodes)
		m.activeNodes[cNode].PartitionList = append(m.activeNodes[cNode].PartitionList, i)
		g.PartitionToNode[i] = cNode
	}

	// Loop through all nodes
	// Assign data to graphsToNodes
	for i := range m.graphsToNodes {
		m.graphsToNodes[i] = newGraphReader()
		m.graphsToNodes[i].Info = gOriginal.Info
		m.graphsToNodes[i].Info.NodeID = i
		m.graphsToNodes[i].Info.NumPartitions = m.numPartitions
		m.graphsToNodes[i].ActiveNodes = m.activeNodes
		m.graphsToNodes[i].PartitionToNode = g.PartitionToNode
		m.graphsToNodes[i].Superstep = m.currentIteration
	}

	// Loop through all vertices
	// Assign data to vertices
	for k, v := range g.Vertices {
		partitionIdx := getPartition(k, m.numPartitions)
		cNode := partitionIdx % len(m.activeNodes)
		m.graphsToNodes[cNode].Vertices[k] = v
		m.graphsToNodes[cNode].Edges[k] = g.Edges[k]
		if _, ok := g.OutQueue[k]; ok {
			m.graphsToNodes[cNode].OutQueue[k] = g.OutQueue[k]
		}
	}

	// Loop through all active vertices
	// Activate active vertices flag on graphsToNodes
	for aV := range g.ActiveVerts {
		partitionIdx := getPartition(aV, m.numPartitions)
		cNode := partitionIdx % len(m.activeNodes)
		m.graphsToNodes[cNode].ActiveVerts = append(m.graphsToNodes[cNode].ActiveVerts, aV)
	}
	go m.savePartStatus()
}

// DisseminateGraph ...
func (m *Master) DisseminateGraph() {
	var wg sync.WaitGroup

	for idx, aNode := range m.activeNodes {
		wg.Add(1)

		// Send graph to all active nodes through POST request
		go func(ip string, wg *sync.WaitGroup, graphToSend graphReader) {
			defer wg.Done()

			req, err := http.NewRequest("POST", getURL(ip, "disseminateGraph"), bytes.NewBuffer(getJSONByteFromGraph(graphToSend)))

			if err != nil {
				panic(err)
			}

			// resp, err2 := m.client.Do(req)
			resp, err2 := m.client.Do(req)

			if err2 != nil {
				panic(err2)
			}

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("Machine", ip, "received graph.")
			}
		}(aNode.IP, &wg, m.graphsToNodes[idx])
	}

	wg.Wait()
}

func getURL(address, path string) string {
	return "http://" + strings.TrimSpace(address) + "/" + path + "/" + strings.TrimSpace(strings.Split(address, ":")[1])
}

// getPartition Get the partition which a vertex belongs to.
func getPartition(vertexID int, numPartitions int) int {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(string(vertexID)))
	return int(algorithm.Sum32() % uint32(numPartitions))
}

func (m *Master) rollback(graphFile string) {
	m.InitConnections()
	m.AssignPartitions(graphFile)
	m.DisseminateGraph()
}

// ------------- State machine stuffs ---------------------

func (m *Master) superstep() (bool, bool) {
	fmt.Println("Starting Superstep", m.currentIteration)
	diedChan := make(chan bool, len(m.activeNodes))
	inactiveChan := make(chan bool, len(m.activeNodes))

	var wg sync.WaitGroup
	for ip, active := range m.nodeAdrs {
		if active {
			wg.Add(1)
			go func(ip string, diedChan, inactiveChan chan bool, wg *sync.WaitGroup) {
				// Start Superstep
				defer wg.Done()

				resp, err := m.client.Get(getURL(ip, "startSuperstep"))
				if err != nil {
					diedChan <- true
					return
				}

				if resp.StatusCode != http.StatusOK {
					diedChan <- true
					return
				}

				// Start pinging
				for {
					fmt.Println("Pinging", ip)
					req, _ := http.NewRequest("POST", getURL(ip, "ping"), bytes.NewBuffer([]byte("Completed Superstep?")))
					pingResp, err2 := m.client.Do(req)
					if err2 != nil || pingResp.StatusCode != http.StatusOK {
						diedChan <- true
						return
					}
					defer pingResp.Body.Close()
					bodyBytes, _ := ioutil.ReadAll(pingResp.Body)
					result := string(bodyBytes)
					if result != "still not done" {
						var activeVert []int
						json.Unmarshal(bodyBytes, &activeVert)
						fmt.Println("Active vertices from", ip, ":", activeVert)
						if len(activeVert) == 0 {
							inactiveChan <- true
						} else {
							inactiveChan <- false
						}
						return
					}
					fmt.Println(ip, "is still busy")
					time.Sleep(time.Second * 5)
				}
			}(ip, diedChan, inactiveChan, &wg)
		}
	}
	fmt.Println("Waiting for Superstep to end")
	wg.Wait()
	fmt.Println("Superstep", m.currentIteration, "end")

	// Checking for dead workers
	nodeDied := false
	close(diedChan)
	for ifNodeDied := range diedChan {
		nodeDied = ifNodeDied || nodeDied
	}

	// Checking for active workers
	close(inactiveChan)
	allInactive := true
	for ifAllInactive := range inactiveChan {
		allInactive = allInactive && ifAllInactive
	}

	fmt.Println("Dead nodes:", nodeDied, ", All nodes inactive:", allInactive)
	return nodeDied, allInactive
}

func (m *Master) saveState() bool {
	fmt.Println("Saving state")
	graphsChan := make(chan *graphReader, len(m.activeNodes))
	var wg sync.WaitGroup
	for ip, active := range m.nodeAdrs {
		if active {
			wg.Add(1)
			go func(ip string, wg *sync.WaitGroup, graphsChan chan *graphReader) {
				defer wg.Done()

				resp, err := m.client.Get(getURL(ip, "saveState"))
				if err != nil {
					return
				}

				if resp.StatusCode == http.StatusOK {
					bodyBytes, err := ioutil.ReadAll(resp.Body) // arr of bytes
					if err != nil {
						panic(err)
					}
					gr := getGraphFromJSONByte(bodyBytes)
					graphsChan <- gr
				}
			}(ip, &wg, graphsChan)
		}
	}
	wg.Wait()
	close(graphsChan)
	saveGraph := newGraphReader()

	for gr := range graphsChan {
		saveGraph.ActiveVerts = append(saveGraph.ActiveVerts, gr.ActiveVerts...)
		for id, vr := range gr.Vertices {
			saveGraph.Vertices[id] = vr
		}
		for id, erList := range gr.Edges {
			if _, ok := saveGraph.Edges[id]; !ok {
				saveGraph.Edges[id] = make([]edgeReader, 0)
			}
			saveGraph.Edges[id] = append(saveGraph.Edges[id], erList...)
		}
		for id, outQ := range gr.OutQueue {
			if _, ok := saveGraph.OutQueue[id]; !ok {
				saveGraph.OutQueue[id] = make([]float64, 0)
			}
			saveGraph.OutQueue[id] = append(saveGraph.OutQueue[id], outQ...)
		}
	}
	saveGraph.CurrentIteration = m.currentIteration
	saveFile := getJSONByteFromGraph(saveGraph)
	ioutil.WriteFile(checkpointPATH, saveFile, 0644)

	fmt.Println("Checking for revived nodes")
	// Checking for revived nodes
	reviveChan := make(chan bool, len(m.nodeAdrs)-len(m.activeNodes))
	for ip, active := range m.nodeAdrs {
		if !active {
			wg.Add(1)
			go func(ip string, reviveChan chan bool, wg *sync.WaitGroup) {
				defer wg.Done()
				_, err := m.client.Get(getURL(ip, "initConnection"))
				if err != nil {
					return
				}
				reviveChan <- true
			}(ip, reviveChan, &wg)
		}
	}
	wg.Wait()
	close(reviveChan)
	nodeRevived := false
	for ifNodeRevived := range reviveChan {
		nodeRevived = ifNodeRevived || nodeRevived
	}
	fmt.Println("Revived nodes:", nodeRevived)
	return nodeRevived
}

func (m *Master) dead() {
	if _, err := os.Stat(checkpointPATH); err == nil {
		m.rollback(checkpointPATH)
	} else {
		m.rollback(m.graphFile)
	}
	m.currentIteration -= m.currentIteration % m.checkpoint
}

func (m *Master) done() {
	var wg sync.WaitGroup
	fmt.Println("Computation has completed.")
	for ip, active := range m.nodeAdrs {
		if active {
			wg.Add(1)
			go func(ip string, wg *sync.WaitGroup) {
				defer wg.Done()
				m.client.Get(getURL(ip, "terminate"))
			}(ip, &wg)
		}
	}
	wg.Wait()

	// if _, err := os.Stat(checkpointPATH); err == nil {
	// 	os.Remove(checkpointPATH)
	// }
}

func (m *Master) pingMaster(rw http.ResponseWriter, r *http.Request) {
	// Check request
	defer r.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	currentCheckpoint := m.currentIteration - (m.currentIteration % m.checkpoint)

	// // Send CP file if not up to date
	if string(bodyBytes) != string(currentCheckpoint) && m.currentIteration > m.checkpoint {
		gr := getGraphFromFile(checkpointPATH)
		gr.CurrentIteration = currentCheckpoint
		sendCP := getJSONByteFromGraph(*gr)
		rw.Write(sendCP)
	}
}

func (m *Master) Run() {
	if len(m.primaryAddress) != 0 {
		for {
			time.Sleep(time.Second * 10)
			// Ping primary
			var sendStr = []byte(string(m.currentIteration))
			fmt.Println(getURL(m.primaryAddress, "pingMaster"))
			req, _ := http.NewRequest("GET", getURL(m.primaryAddress, "pingMaster"), bytes.NewBuffer(sendStr))
			pingResp, err := m.client.Do(req)
			if err != nil {
				// Primary not responding -> break!
				fmt.Println("Primary not responding! Taking over!")
				break
			}

			fmt.Println("Primary alive.")
			// Check for CP file
			defer pingResp.Body.Close()
			bodyBytes, _ := ioutil.ReadAll(pingResp.Body)
			if strings.TrimSpace(string(bodyBytes)) != "" {
				fmt.Println("Received checkpoint file, saving.")
				// Save CP file
				gr := getGraphFromJSONByte(bodyBytes)
				m.currentIteration = gr.CurrentIteration
				// saveFile := getJSONByteFromGraph(gr)
				ioutil.WriteFile(checkpointPATH, bodyBytes, 0644)
			}

		}

	}

	// Start replica-response server
	go func(m *Master) {
		http.HandleFunc(getPortPath("/pingMaster", m.port), m.pingMaster)
		http.ListenAndServe(fmt.Sprint(":", m.port), nil)
	}(m)

	currentState := SUPERSTEP
	if m.currentIteration != 0 {
		m.rollback(checkpointPATH)
	} else {
		m.rollback(m.graphFile)
	}

	for {
		switch currentState {
		case SUPERSTEP:
			nodeDied, allInactive := m.superstep()
			if nodeDied {
				currentState = DEAD
			} else if allInactive {
				currentState = DONE
			} else {
				if m.currentIteration != 0 && m.currentIteration%m.checkpoint == 0 {
					currentState = SAVESTATE
				} else {
					currentState = SUPERSTEP
				}
				m.currentIteration++
			}
		case SAVESTATE:
			nodeRevived := m.saveState()
			if nodeRevived {
				m.rollback(checkpointPATH)
			}
			currentState = SUPERSTEP
		case DEAD:
			m.dead()
			currentState = SUPERSTEP
		case DONE:
			m.done()
			return
		}

	}
}
