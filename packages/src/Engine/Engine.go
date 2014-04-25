// Engine project Engine.go
package Engine

import (
	"BaseModule"
	"ClientProcessor"
	"Config"
	"DatabaseModule"
	"EngineTypes"
	"Module"
	"TheMap"
	"fmt"
	"time"
)

type Engine struct {
	modules         map[int]BaseModule.BaseModule
	requestsQueue   EngineTypes.MessageQueue
	responsesQueue  EngineTypes.MessageQueue
	inbox           chan EngineTypes.Message
	clientInbox     chan EngineTypes.Message
	clientOutbox    chan EngineTypes.Message
	taskContainer   EngineTypes.TaskContainer
	lastMessageId   int
	databaseManager DatabaseModule.DatabaseManager
	clientProcessor ClientProcessor.ClientProcessor
}

func (e *Engine) run() {

	go e.reciveMessages()
	go e.taskGenerator()
	go e.taskManager()

	for {
		time.Sleep(1000 * time.Millisecond)
		// here will be some debug information
	}
}

func (e *Engine) loadModules() bool {
	e.modules = make(map[int]BaseModule.BaseModule)
	var module BaseModule.BaseModule
	module.InitB(&Module.Module{}, e.databaseManager.Create(EngineTypes.MODULE))
	module.SetOutbox(&e.inbox)

	if !module.Start() {

		return false
	}
	e.modules[EngineTypes.MODULE] = module
	fmt.Println("Module Module loaded.")

	time.Sleep(1000 * time.Millisecond)

	var mapa BaseModule.BaseModule
	mapa.InitB(&TheMap.Mapa{}, e.databaseManager.Create(EngineTypes.MAP))
	mapa.SetOutbox(&e.inbox)

	if !mapa.Start() {
		return false
	}
	e.modules[EngineTypes.MAP] = mapa
	fmt.Println("Module Mapa loaded.")

	time.Sleep(1000 * time.Millisecond)

	return true
}

func (e *Engine) reciveMessages() {
	var err error
	for {
		select {
		case msg := <-e.inbox:

			if msg.Request {
				err = e.requestsQueue.Push(&msg)
			} else {
				err = e.responsesQueue.Push(&msg)
			}

		}
		if err != nil {
			fmt.Println("Engine error:", err)
		}

	}
}

func (e *Engine) sendMessage(m BaseModule.BaseModule, msg EngineTypes.Message) {

	outbox := *m.GetInbox()
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	select {
	case outbox <- msg:
	case <-timeout:
		fmt.Println("Message not send. Timeout!.")
	}

}

func (e *Engine) taskGenerator() {

	for {
		if e.requestsQueue.Empty {
			<-e.requestsQueue.WakeUp
		}

		req := e.requestsQueue.Pop()

		switch req.Action {
		case "getCords":
			e.getCords(*req)
		case "getArea":
			e.getArea(*req)
		default:
			fmt.Println("Engine debug:Wrong action!")

		}
	}

}

func (e *Engine) taskManager() {

	for {
		if e.responsesQueue.Empty {
			<-e.responsesQueue.WakeUp
		}
		req := e.responsesQueue.Pop()
		task, err := e.taskContainer.GetTask(req.MessageId)

		if err != nil {
			fmt.Println("Engine debug:", err)
			continue
		}
		task.Input <- *req
	}

}

func (e *Engine) Init(in chan EngineTypes.Message, out chan EngineTypes.Message) {
	e.inbox = make(chan EngineTypes.Message, Config.EngineInboxSize)

	e.responsesQueue = EngineTypes.MessageQueue{}
	e.responsesQueue.Init(Config.EngineResponsesQueueSize)

	e.requestsQueue = EngineTypes.MessageQueue{}
	e.requestsQueue.Init(Config.EngineRequestsQueueSize)

	e.lastMessageId = 0

	e.clientProcessor = ClientProcessor.ClientProcessor{}
	e.clientProcessor.Init(in, out, &e.inbox)

	e.databaseManager = DatabaseModule.DatabaseManager{}
	e.databaseManager.Init()
}

func (e Engine) Start(in chan EngineTypes.Message, out chan EngineTypes.Message) {
	fmt.Println("Engine:Starting...")
	e.Init(in, out)
	if !e.loadModules() {
		fmt.Println("Engine:Failed to load modules!\nBye.")
		return
	}
	e.clientProcessor.Start()
	go e.run()

}

func (e *Engine) sendMessageToClient(msg EngineTypes.Message) {
	outbox := e.clientProcessor.Inbox
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(Config.EngineSendMessageTimeout * time.Second)
		timeout <- true
	}()

	select {
	case outbox <- msg:
	case <-timeout:
		fmt.Println("Message not send. Timeout!.")
	}
}

/////////// ACTION FUNCTIONS /////////////

func (e *Engine) getCords(req EngineTypes.Message) {
	task := EngineTypes.Task{}                     // create task structure
	task.Input = make(chan EngineTypes.Message, 3) //todo: put it into config file
	task.Run = func() {                            // create functions that just replays the msg
		req.Request = false // becouse msg is a request, change it to response
		var msg EngineTypes.Message
		msg.Action = "getCords"
		msg.Request = true
		e.lastMessageId++
		if e.lastMessageId == 2147483647 {
			e.lastMessageId = -2147483647
		}
		msg.MessageId = e.lastMessageId
		e.sendMessage(e.modules[EngineTypes.MAP], msg)
		e.taskContainer.PushTask(task, true, msg.MessageId)
		res := <-task.Input
		res.MessageId = req.MessageId
		if req.Sender == EngineTypes.CLIENT_PROCESSOR {
			e.sendMessageToClient(res)
		} else {
			e.sendMessage(e.modules[req.Sender], res)
		}

	}
	go task.Run() // run task in background
}

func (e *Engine) getArea(req EngineTypes.Message) {
	task := EngineTypes.Task{} // create task structure
	task.Input = make(chan EngineTypes.Message, 3)
	task.Run = func() { // create functions that just replays the msg

		reqId := req.MessageId
		e.lastMessageId++
		if e.lastMessageId == 2147483647 {
			e.lastMessageId = -2147483647
		}
		req.MessageId = e.lastMessageId
		e.sendMessage(e.modules[EngineTypes.MAP], req)
		e.taskContainer.PushTask(task, true, req.MessageId)
		res := <-task.Input

		res.MessageId = reqId
		if req.Sender == EngineTypes.CLIENT_PROCESSOR {
			e.sendMessageToClient(res)
		} else {
			e.sendMessage(e.modules[req.Sender], res)
		}

	}
	go task.Run() // run task in background
}
