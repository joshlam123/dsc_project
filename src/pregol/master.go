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
)

// Master ...
type Master struct {
	numPartitions    int
	checkpoint       int
	nodeAdrs         map[string]bool
	activeNodes      []activeNode
	graphsToNodes    []graphReader
	graphFile        string
	client           *http.Client
	currentIteration int
}

type guiSend struct {
	master Master
	iter   int
}

// NewMaster Constructor for Master struct
func NewMaster(numPartitions, checkpoint int, ipFile, graphFile string) *Master {
	m := Master{}
	m.numPartitions = numPartitions
	m.checkpoint = checkpoint
	m.activeNodes = make([]activeNode, 0)
	m.graphFile = graphFile
	m.client = &http.Client{
		Timeout: time.Second * 5,
	}
	m.currentIteration = 0
	dat, err := ioutil.ReadFile(ipFile)
	if err != nil {
		panic(err)
	}
	m.nodeAdrs = make(map[string]bool)
	for _, ip := range strings.Split(string(dat), "\n") {
		m.nodeAdrs[ip] = false
	}

	return &m
}

// InitConnections Initializes connections with all machines via ip addresses found in nodeAdrs []string.
// Generates a list of activeNodes which are the machines that Master will request to start superstep.
func (m *Master) InitConnections() {

	var wg sync.WaitGroup
	activeNodeChan := make(chan activeNode, len(m.nodeAdrs))
	for ip := range m.nodeAdrs {
		wg.Add(1)

		// Initialize connection with all nodes, and check if they are active using GET request
		go func(ip string, wg *sync.WaitGroup, activeNodeChan chan activeNode) {
			defer wg.Done()

			resp, err := m.client.Get(getURL(ip, "initConnection"))
			if err != nil {
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("Machine", ip, "connected.")
				activeNodeChan <- activeNode{ip, make([]int, 0)}

				// bodyBytes, err := ioutil.ReadAll(resp.Body)
				// if err != nil {
				// 	log.Fatal(err)
				// }
				// bodyString := string(bodyBytes)
				// fmt.Println(bodyString)
			}
		}(ip, &wg, activeNodeChan)

	}
	wg.Wait()
	close(activeNodeChan)

	m.activeNodes = make([]activeNode, 0)
	m.nodeAdrs = make(map[string]bool)
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

// AssignPartitions Assign partToVert to active nodes
func (m *Master) AssignPartitions(graphFile string) {
	g := getGraphFromFile(graphFile)
	g.PartitionToNode = make(map[int]int)
	m.graphsToNodes = make([]graphReader, len(m.activeNodes))

	for i := 0; i < m.numPartitions; i++ {
		cNode := i % len(m.activeNodes)
		m.activeNodes[cNode].PartitionList = append(m.activeNodes[cNode].PartitionList, i)
		g.PartitionToNode[i] = cNode
	}

	for i := range m.graphsToNodes {
		m.graphsToNodes[i] = newGraphReader()
		m.graphsToNodes[i].Info = g.Info
		m.graphsToNodes[i].Info.NodeID = i
		m.graphsToNodes[i].Info.NumPartitions = m.numPartitions
		m.graphsToNodes[i].ActiveNodes = m.activeNodes
		m.graphsToNodes[i].PartitionToNode = g.PartitionToNode
		m.graphsToNodes[i].superstep = m.currentIteration
	}

	for k, v := range g.Vertices {
		partitionIdx := getPartition(k, m.numPartitions)
		cNode := partitionIdx % len(m.activeNodes)
		m.graphsToNodes[cNode].Vertices[k] = v
		m.graphsToNodes[cNode].Edges[k] = g.Edges[k]
		if _, ok := g.outQueue[k]; ok {
			m.graphsToNodes[cNode].outQueue[k] = g.outQueue[k]
		}
	}

	for aV := range g.ActiveVerts {
		partitionIdx := getPartition(aV, m.numPartitions)
		cNode := partitionIdx % len(m.activeNodes)
		m.graphsToNodes[cNode].ActiveVerts = append(m.graphsToNodes[cNode].ActiveVerts, aV)
	}
}

// DisseminateGraph ...
func (m *Master) DisseminateGraph() {
	var wg sync.WaitGroup

	for idx, aNode := range m.activeNodes {
		wg.Add(1)

		// Send graph to all active nodes through POST request
		go func(ip string, wg *sync.WaitGroup, graphToSend graphReader) {
			defer wg.Done()

			c := &http.Client{}
			req, err := http.NewRequest("POST", getURL(ip, "disseminateGraph"), bytes.NewBuffer(getJSONByteFromGraph(graphToSend)))

			if err != nil {
				panic(err)
			}

			// resp, err2 := m.client.Do(req)
			resp, err2 := c.Do(req)

			if err2 != nil {
				panic(err2)
			}

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("Machine", ip, "received graph.")

				// bodyBytes, err := ioutil.ReadAll(resp.Body)
				// if err != nil {
				// 	log.Fatal(err)
				// }
				// bodyString := string(bodyBytes)
				// fmt.Println(bodyString)
			}
		}(aNode.IP, &wg, m.graphsToNodes[idx])
	}

	wg.Wait()
}

func getURL(address, path string) string {
	return "http://" + strings.TrimSpace(address) + "/" + path
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
	nodeDiedChan := make(chan bool, len(m.activeNodes))
	inactiveChan := make(chan bool, len(m.activeNodes))

	var wg sync.WaitGroup
	for ip, active := range m.nodeAdrs {
		if active {
			wg.Add(1)
			go func(ip string, nodeDiedChan, inactiveChan chan bool, wg *sync.WaitGroup) {
				// Start Superstep
				defer wg.Done()

				resp, err := m.client.Get(getURL(ip, "startSuperstep"))
				if err != nil {
					nodeDiedChan <- true
					return
				}

				if resp.StatusCode != http.StatusOK {
					nodeDiedChan <- true
					return
				}

				// Start pinging
				for {
					// pingResp, err2 := m.client.Get(getURL(ip, "3000", "ping"))
					fmt.Println("Pinging", ip)
					req, _ := http.NewRequest("POST", getURL(ip, "ping"), bytes.NewBuffer([]byte("Completed Superstep?")))
					pingResp, err2 := m.client.Do(req)
					if err2 != nil {
						nodeDiedChan <- true
						return
					}

					if pingResp.StatusCode != http.StatusOK {
						nodeDiedChan <- true
						return
					}
					defer pingResp.Body.Close()
					bodyBytes, _ := ioutil.ReadAll(pingResp.Body)
					fmt.Println(bodyBytes)
					result := string(bodyBytes)
					fmt.Println(result)
					if result != "still not done" {
						var activeVert []int
						json.Unmarshal(bodyBytes, &activeVert)
						fmt.Println(activeVert)
						if len(activeVert) == 0 {
							fmt.Println("No active workers")
							inactiveChan <- true
						} else {
							inactiveChan <- false
						}
						return
					} else {
						fmt.Println(ip, "still busy")
					}
					time.Sleep(time.Second * 5)
				}
			}(ip, nodeDiedChan, inactiveChan, &wg)
		}
	}
	fmt.Println("Waiting")
	wg.Wait()
	fmt.Println("Superstep completed")

	// Checking for dead workers
	nodeDied := false
	fmt.Println("Checking for dead workers")
	close(nodeDiedChan)
	for ifNodeDied := range nodeDiedChan {
		nodeDied = ifNodeDied || nodeDied
	}

	// Checking for revived nodes
	// nodeRevivedChan := make(chan bool, len(m.nodeAdrs)-len(m.activeNodes))
	// for ip, active := range m.nodeAdrs {
	// 	if !active {
	// 		wg.Add(1)
	// 		go func(ip string, wg *sync.WaitGroup) {
	// 			defer wg.Done()
	// 			_, err := m.client.Get(getURL(ip, "3000", "ping"))
	// 			if err != nil {
	// 				return
	// 			}
	// 			nodeRevivedChan <- true
	// 		}(ip, &wg)
	// 	}
	// }
	// wg.Wait()
	// close(nodeRevivedChan)
	// for ifNodeRevived := range nodeRevivedChan {
	// 	nodeRevived = ifNodeRevived || nodeRevived
	// }

	// Checking for active workers
	fmt.Println("Checking for active workers")
	close(inactiveChan)
	allInactive := true
	for ifAllInactive := range inactiveChan {
		allInactive = allInactive && ifAllInactive
	}
	return nodeDied, allInactive
}

func (m *Master) saveState() {
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
		saveGraph.Info = gr.Info
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
		for id, outQ := range gr.outQueue {
			if _, ok := saveGraph.outQueue[id]; !ok {
				saveGraph.outQueue[id] = make([]float64, 0)
			}
			saveGraph.outQueue[id] = append(saveGraph.outQueue[id], outQ...)
		}
	}
	saveFile := getJSONByteFromGraph(saveGraph)
	ioutil.WriteFile(checkpointPATH, saveFile, 0644)
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

	if _, err := os.Stat(checkpointPATH); err == nil {
		os.Remove(checkpointPATH)
	}
}

func (m *Master) Run() {
	currentState := SUPERSTEP
	m.rollback(m.graphFile)
	for {
		switch currentState {
		case SUPERSTEP:
			nodeDied, allInactive := m.superstep()
			m.currentIteration++
			if nodeDied {
				currentState = DEAD
			} else if allInactive {
				currentState = DONE
			} else {
				currentState = SUPERSTEP
				if m.currentIteration != 0 && m.currentIteration%m.checkpoint == 0 {
					m.saveState()
				}
			}
		case DEAD:
			m.dead()
			currentState = SUPERSTEP
		case DONE:
			m.done()
			return
		}
	}
}

// // Run ...
// func (m *Master) Run() {
// 	currentIter := 0
// 	nodeDied := true
// 	// nodeRevived := false
// 	// checkpointFile := "checkpoint.json"

// 	for {
// 		if nodeDied {
// 			if currentIter < m.checkpoint {
// 				m.rollback(m.graphFile)
// 				currentIter = 0
// 				// } else {
// 				// 	m.rollback(checkpointFile)
// 				// 	// TODO: Load messages
// 				// 	currentIter -= (currentIter % m.checkpoint)
// 				// }
// 				nodeDied = false
// 				//continue
// 			}

// 			if currentIter%m.checkpoint == 0 {
// 				var wg2 sync.WaitGroup
// 				for ip, active := range m.nodeAdrs {
// 					if active {
// 						wg2.Add(1)
// 						go func(ip string, wg *sync.WaitGroup) {
// 							// TODO: Start superstep
// 							defer wg.Done()

// 							resp, err := m.client.Get(getURL(ip, "3000", "saveState"))
// 							if err != nil {

// 								return
// 							}

// 							if resp.StatusCode != http.StatusOK {
// 								return
// 							}

// 						}(ip, &wg2)
// 					}
// 					// 	// TODO: Save worker states

// 					// 	if nodeRevived {
// 					// 		m.rollback(checkpointFile)
// 					// 		// TODO: Load messages
// 					// 		nodeRevived = false
// 					// 	}
// 					// }
// 				}

// 				nodeDiedChan := make(chan bool, len(m.activeNodes))
// 				inactiveChan := make(chan bool, len(m.activeNodes))

// 				var wg sync.WaitGroup
// 				for ip, active := range m.nodeAdrs {
// 					if active {
// 						wg.Add(1)
// 						go func(ip string, nodeDiedChan, inactiveChan chan bool, wg *sync.WaitGroup) {
// 							// TODO: Start superstep
// 							defer wg.Done()

// 							resp, err := m.client.Get(getURL(ip, "3000", "startSuperstep"))
// 							if err != nil {
// 								nodeDiedChan <- true
// 								return
// 							}

// 							if resp.StatusCode != http.StatusOK {
// 								nodeDiedChan <- true
// 								return
// 							}

// 							// Start pinging
// 							for {
// 								// pingResp, err2 := m.client.Get(getURL(ip, "3000", "ping"))
// 								fmt.Println("Pinging", ip)
// 								req, _ := http.NewRequest("POST", getURL(ip, "3000", "ping"), bytes.NewBuffer([]byte("Completed Superstep?")))
// 								pingResp, err2 := m.client.Do(req)
// 								if err2 != nil {
// 									nodeDiedChan <- true
// 									return
// 								}

// 								if pingResp.StatusCode != http.StatusOK {
// 									nodeDiedChan <- true
// 									return
// 								}
// 								defer pingResp.Body.Close()
// 								bodyBytes, _ := ioutil.ReadAll(pingResp.Body)
// 								fmt.Println(bodyBytes)
// 								result := string(bodyBytes)
// 								fmt.Println(result)
// 								if result != "still not done" {
// 									var activeVert []int
// 									json.Unmarshal(bodyBytes, &activeVert)
// 									fmt.Println(activeVert)
// 									if len(activeVert) == 0 {
// 										fmt.Println("No active workers")
// 										inactiveChan <- true
// 									} else {
// 										inactiveChan <- false
// 									}
// 									return
// 								} else {
// 									fmt.Println(ip, "still busy")
// 								}
// 								time.Sleep(time.Second * 5)
// 							}
// 						}(ip, nodeDiedChan, inactiveChan, &wg)
// 					}
// 				}
// 				fmt.Println("Waiting")
// 				wg.Wait()
// 				fmt.Println("Superstep completed")
// 				close(nodeDiedChan)
// 				fmt.Println("Checking for dead workers")
// 				for ifNodeDied := range nodeDiedChan {
// 					nodeDied = ifNodeDied || nodeDied
// 				}
// 				close(inactiveChan)
// 				fmt.Println("Checking for active workers")
// 				allInactive := true
// 				for ifAllInactive := range inactiveChan {
// 					allInactive = allInactive && ifAllInactive
// 				}
// 				if allInactive {
// 					fmt.Println("Computation has completed.")
// 					for ip, active := range m.nodeAdrs {
// 						if active {
// 							wg.Add(1)
// 							go func(ip string, wg *sync.WaitGroup) {
// 								defer wg.Done()
// 								m.client.Get(getURL(ip, "3000", "terminate"))
// 							}(ip, &wg)
// 						}
// 					}
// 					wg.Wait()
// 					break
// 				}

// 				// Check nodeRevived
// 				// nodeRevivedChan := make(chan bool, len(m.nodeAdrs)-len(m.activeNodes))
// 				// for ip, active := range m.nodeAdrs {
// 				// 	if !active {
// 				// 		wg.Add(1)
// 				// 		go func(ip string, wg *sync.WaitGroup) {
// 				// 			defer wg.Done()
// 				// 			_, err := m.client.Get(getURL(ip, "3000", "ping"))
// 				// 			if err != nil {
// 				// 				return
// 				// 			}
// 				// 			nodeRevivedChan <- true
// 				// 		}(ip, &wg)
// 				// 	}
// 				// }
// 				// wg.Wait()
// 				// close(nodeRevivedChan)
// 				// for ifNodeRevived := range nodeRevivedChan {
// 				// 	nodeRevived = ifNodeRevived || nodeRevived
// 				// }

// 				// TODO: Check end condition

// 				currentIter++

// 				// TODO: JOSH send the master condition to GUI
// 				//guiMsg := guiSend{master: m, iter: currentIter}
// 				//req, err := http.NewRequest("POST", getURL(ip, "3000", "guiserver"), bytes.NewBuffer(guiMsg), currentIter)

// 			}
// 		}
// 	}
// }
