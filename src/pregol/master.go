package pregol

import (
	"bytes"
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
	ActiveNodes   []ActiveNode
	graphsToNodes []graphReader
}

// NewMaster ...
func NewMaster(numPartitions int) *Master {
	m := Master{}
	m.numPartitions = numPartitions
	m.ActiveNodes = make([]ActiveNode, 0)

	return &m
}

// InitConnections ...
func (m *Master) InitConnections(ipFile string) {
	dat, err := ioutil.ReadFile(ipFile)
	if err != nil {
		panic(err)
	}
	nodeAdrss := strings.Split(string(dat), "\n")

	var wg sync.WaitGroup
	activeNodeChan := make(chan ActiveNode, len(nodeAdrss))
	for i := range nodeAdrss {
		wg.Add(1)

		go func(ip string, wg *sync.WaitGroup, activeNodeChan chan ActiveNode) {
			defer wg.Done()

			var client = &http.Client{
				Timeout: time.Second * 10,
			}

			resp, err := client.Get(getURL(ip, "3000"))
			if err != nil {
				panic(err)
			}

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("Machine", ip, "connected.")
				activeNodeChan <- ActiveNode{ip, make([]int, 0)}

				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				bodyString := string(bodyBytes)
				fmt.Println(bodyString)
			}
		}(nodeAdrss[i], &wg, activeNodeChan)

	}
	wg.Wait()
	close(activeNodeChan)
	for elem := range activeNodeChan {
		m.ActiveNodes = append(m.ActiveNodes, elem)
	}
}

// AssignPartitions Assign partitions to active nodes
func (m *Master) AssignPartitions(graphFile string) {
	g := getGraphFromFile(graphFile)
	m.graphsToNodes = make([]graphReader, len(m.ActiveNodes))

	for i := 0; i < m.numPartitions; i++ {
		cNode := i % len(m.ActiveNodes)
		m.ActiveNodes[cNode].partitionList = append(m.ActiveNodes[cNode].partitionList, i)
	}

	for k, v := range g.Vertices {
		partitionIdx := getPartition(k, m.numPartitions)
		cNode := partitionIdx % len(m.ActiveNodes)
		m.graphsToNodes[cNode].Vertices[k] = v
		m.graphsToNodes[cNode].Edges[k] = g.Edges[k]
	}

	for i := range m.graphsToNodes {
		m.graphsToNodes[i].Info = g.Info
		m.graphsToNodes[i].ActiveNodes = m.ActiveNodes
	}
}

// DisseminateGraph ...
func (m *Master) DisseminateGraph(graphFile string) {
	var wg sync.WaitGroup

	for idx, aNode := range m.ActiveNodes {
		go func(ip string, wg *sync.WaitGroup, graphToSend graphReader) {
			defer wg.Done()

			var client = &http.Client{
				Timeout: time.Second * 10,
			}
			req, err := http.NewRequest("POST", getURL(ip, "3000"), bytes.NewBuffer(getJSONByteFromGraph(graphToSend)))

			if err != nil {
				panic(err)
			}

			resp, err2 := client.Do(req)

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
		}(aNode.ip, &wg, m.graphsToNodes[idx])
		wg.Wait()
	}
}

func getURL(ip, port string) string {
	return "http://" + ip + ":" + port
}

// getPartition Get the partition which a vertex belongs to.
func getPartition(vertexID int, numPartitions int) int {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(string(vertexID)))
	return int(algorithm.Sum32() % uint32(numPartitions))
}
