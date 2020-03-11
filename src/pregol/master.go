package pregol

import (
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
	NodeAdrss     []string
	numPartitions int
	ActiveNodes   []ActiveNode
}

type ActiveNode struct {
	ip            string
	partitionList []int
}

// NewMaster ...
func NewMaster(ipFile string, numPartitions int) *Master {
	m := Master{}
	m.numPartitions = numPartitions
	m.ActiveNodes = make([]ActiveNode, 0)

	dat, err := ioutil.ReadFile(ipFile)
	if err != nil {
		panic(err)
	}
	m.NodeAdrss = strings.Split(string(dat), "\n")
	return &m
}

// InitConnections ...
func (m *Master) InitConnections() {
	var wg sync.WaitGroup
	activeNodeChan := make(chan ActiveNode, len(m.NodeAdrss))
	for i := range m.NodeAdrss {
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
		}(m.NodeAdrss[i], &wg, activeNodeChan)

	}
	wg.Wait()
	close(activeNodeChan)
	for elem := range activeNodeChan {
		m.ActiveNodes = append(m.ActiveNodes, elem)
	}
}

// AssignPartitions Assign partitions to active nodes
func (m *Master) AssignPartitions() {
	for i := 0; i < m.numPartitions; i++ {
		cNode := i % len(m.ActiveNodes)
		m.ActiveNodes[cNode].partitionList = append(m.ActiveNodes[cNode].partitionList, i)
	}
}

func (m *Master) DisseminateGraph(graphFile string) {
	g := getGraphFromFile(graphFile)
	for idx, aNode := range m.ActiveNodes {
		go func() {
			// Do something
		}()
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
