package main

import (
	"fmt"
	"pregol"
)

func main() {
	pregol.InitWorker()
	pregol.UdfChan <- func(vertex *Vertex, superstep int) (bool, map[int]float64) {
		// Do smth

		return (false, make(map[int]float64))
	}
	pregol.Run()
	select {}
}
