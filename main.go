package main

import (
	"Engine"
	"EngineTypes"
	"HttpServer"
	_ "github.com/ziutek/mymysql/godrv"
)

func main() {

	engineClientInbox := make(chan EngineTypes.Message, 1000) //todo: put it into config file
	clientEngineInbox := make(chan EngineTypes.Message, 1000) //todo: put it into config file

	var engine Engine.Engine
	engine.Start(engineClientInbox, clientEngineInbox)

	var httpServer HttpServer.HttpServer
	httpServer.Start(clientEngineInbox, engineClientInbox)

	//newLoad()
	wait := make(chan bool)
	select {
	case <-wait:
	}
}
