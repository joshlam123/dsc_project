package main

import (
	"fmt"
	"pregol"
)

func main() {
	pregol.SetUdf(func(vertex *Vertex, superstep int) (bool, map[int]float64) {
		// Do smth

		return (false, make(map[int]float64))
	})
	pregol.Run()
	select {}
}
