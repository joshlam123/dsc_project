package main

import (
	"pregol"
)

func main() {
	m := pregol.NewMaster(6, 1, "ip_add.txt", "../examples/data/unweighted/prob/rand100.json")
	m.Run()
}
