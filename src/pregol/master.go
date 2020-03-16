package pregol

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Master ...
type Master struct {
	numPartitions int
	checkpoint    int
	nodeAdrs      map[string]bool
	activeNodes   []activeNode
	graphsToNodes []graphReader
	graphFile     string
	client        *http.Client
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

			resp, err := m.client.Get(getURL(ip, "3000", "initConnection"))
			if err != nil {
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("Machine", ip, "connected.")
				activeNodeChan <- activeNode{ip, make([]int, 0)}

				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				bodyString := string(bodyBytes)
				fmt.Println(bodyString)
			}
		}(ip, &wg, activeNodeChan)

	}
	wg.Wait()
	close(activeNodeChan)

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
		m.graphsToNodes[i].ActiveNodes = m.activeNodes
	}

	for k, v := range g.Vertices {
		partitionIdx := getPartition(k, m.numPartitions)
		cNode := partitionIdx % len(m.activeNodes)
		m.graphsToNodes[cNode].Vertices[k] = v
		m.graphsToNodes[cNode].Edges[k] = g.Edges[k]
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

			req, err := http.NewRequest("POST", getURL(ip, "3000", "disseminateGraph"), bytes.NewBuffer(getJSONByteFromGraph(graphToSend)))

			if err != nil {
				panic(err)
			}

			resp, err2 := m.client.Do(req)

			if err2 != nil {
				panic(err2)
			}

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("Machine", ip, "received graph.")

				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				bodyString := string(bodyBytes)
				fmt.Println(bodyString)
			}
		}(aNode.IP, &wg, m.graphsToNodes[idx])
	}

	wg.Wait()
}

func getURL(ip, port, path string) string {
	return "http://" + strings.TrimSpace(ip) + ":" + port + "/" + path
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

// Run ...
func (m *Master) Run() {
	currentIter := 0
	nodeDied := true
	nodeRevived := false
	checkpointFile := "checkpoint.json"

	for {
		if nodeDied {
			if currentIter < m.checkpoint {
				m.rollback(m.graphFile)
				currentIter = 0
			} else {
				m.rollback(checkpointFile)
				// TODO: Load messages
				currentIter -= (currentIter % m.checkpoint)
			}
			nodeDied = false
			continue
		}

		if currentIter%m.checkpoint == 0 {
			// TODO: Save worker states

			if nodeRevived {
				m.rollback(checkpointFile)
				// TODO: Load messages
				nodeRevived = false
			}
		}

		nodeDiedChan := make(chan bool, len(m.activeNodes))

		var wg sync.WaitGroup
		for ip, active := range m.nodeAdrs {
			if active {
				wg.Add(1)
				go func(ip string, nodeDiedChan chan bool, wg *sync.WaitGroup) {
					// TODO: Start superstep and ping
					defer wg.Done()

					resp, err := m.client.Get(getURL(ip, "3000", "startSuperstep"))
					if err != nil {
						nodeDiedChan <- true
						return
					}

					if resp.StatusCode != http.StatusOK {
						nodeDiedChan <- true
						return
					}

					for {
						pingResp, err2 := m.client.Get(getURL(ip, "3000", "ping"))
						if err2 != nil {
							nodeDiedChan <- true
							return
						}

						if pingResp.StatusCode != http.StatusOK {
							nodeDiedChan <- true
							return
						}

						bodyBytes, _ := ioutil.ReadAll(resp.Body)
						result := string(bodyBytes)
						if result == "done" {
							return
						}
						time.Sleep(time.Second * 10)
					}

				}(ip, nodeDiedChan, &wg)
			}
		}

		wg.Wait()
		close(nodeDiedChan)
		for ifNodeDied := range nodeDiedChan {
			nodeDied = ifNodeDied || nodeDied
		}

		// Check nodeRevived
		nodeRevivedChan := make(chan bool, len(m.nodeAdrs)-len(m.activeNodes))
		for ip, active := range m.nodeAdrs {
			if !active {
				wg.Add(1)
				go func(ip string, wg *sync.WaitGroup) {
					defer wg.Done()
					resp, err := m.client.Get(getURL(ip, "3000", "ping"))
					if err != nil {
						return
					}
					nodeRevivedChan <- true
				}(ip, &wg)
			}
		}
		wg.Wait()
		close(nodeRevivedChan)
		for ifNodeRevived := range nodeRevivedChan {
			nodeRevived = ifNodeRevived || nodeRevived
		}

		// TODO: Check end condition
		currentIter++

		// TODO: JOSH send the master condition to GUI
		//guiMsg := guiSend{master: m, iter: currentIter}
		//req, err := http.NewRequest("POST", getURL(ip, "3000", "guiserver"), bytes.NewBuffer(guiMsg), currentIter)
	}
}
