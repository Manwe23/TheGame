package main

import (
	"Config"
	"Engine"
	"EngineTypes"
	"HttpServer"
	_ "github.com/ziutek/mymysql/godrv"
)

func main() {

	engineClientInbox := make(chan EngineTypes.Message, Config.EngineClientInboxSize)
	clientEngineInbox := make(chan EngineTypes.Message, Config.ClientEngineInboxSize)

	var engine Engine.Engine
	engine.Start(engineClientInbox, clientEngineInbox)

	var httpServer HttpServer.HttpServer
	httpServer.Start(clientEngineInbox, engineClientInbox)
  fmt.Printf("Jest super!!!\n")
	//newLoad()
	wait := make(chan bool)
	select {
	case <-wait:
	}
}
