package shaker

import (
	zmq "github.com/pebbe/zmq4"
)

type Heartbeat struct {
	socket zmq.Socket
}

func (heartbeat *Heartbeat) startHeartbeat(endpoint string) {
	if heartbeat.socket == (zmq.Socket{}) {
		//context, _ := zmq.Context zmq.NewContext() //@todo check error
		//heartbeat.socket = context.NewSocket(zmq.PUSH)
		//heartbeat.socket.Connect(endpoint)
	}
}
