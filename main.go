package main

import (
	"Engine"
)

func main() {

	var engine Engine.Engine
	engine.Start()

	//var httpServer HttpServer.HttpServer
	//httpServer.Start()

	wait := make(chan bool)
	select {
	case <-wait:
	}
}
