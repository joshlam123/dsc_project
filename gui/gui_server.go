package main

import ("pregol"
		"websocket")

type T struct {
	Msg string
	Count int
}

// receive JSON type T
var data T
websocket.JSON.Receive(ws, &data)

// send JSON type T
websocket.JSON.Send(ws, data)

var Message = Codec{marshal, unmarshal}
