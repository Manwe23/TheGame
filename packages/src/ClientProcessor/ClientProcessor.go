// ClientProcessor project ClientProcessor.go
package ClientProcessor

import (
	"BaseModule"
	"DatabaseModule"
	"EngineTypes"
	"fmt"
)

type ClientProcessor struct {
	requestsQueue    EngineTypes.MessageQueue
	setRequestsQueue EngineTypes.MessageQueue       // todo: change it to queue with long history
	cache            map[string]EngineTypes.Message // todo: upgrade it to groupcache
	clientInbox      chan EngineTypes.Message
	clientOutbox     chan EngineTypes.Message
	engineInbox      chan EngineTypes.Message
	Inbox            chan EngineTypes.Message
	engineConnector  BaseModule.BaseModule // module that i will use as connector to engine. It will be responsible for communication
	// between client processor and engine
}

type EngineConnector struct {
	sendMessage func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool)
}

func (e *EngineConnector) InitModule(sender func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool), h *DatabaseModule.DatabaseModule) {
	e.sendMessage = sender
}

/* Function binds request to proper function */
func (e *EngineConnector) GenerateTask(req EngineTypes.Message) {
	switch req.Action {
	case "getCords":
		//e.getCords(req) // sample action1 bind
	default:
		fmt.Println("Map debug: Wrong action!") // if there is no proper action, report an error
	}
}

func (c *ClientProcessor) Init(in chan EngineTypes.Message, out chan EngineTypes.Message, engineIn *chan EngineTypes.Message) {
	c.requestsQueue = EngineTypes.MessageQueue{}
	c.requestsQueue.Init(1000) //todo: put it into config file

	c.setRequestsQueue = EngineTypes.MessageQueue{}
	c.setRequestsQueue.Init(1000) //todo: put it into config file

	c.cache = make(map[string]EngineTypes.Message)
	c.Inbox = make(chan EngineTypes.Message, 1000) //todo: put it into config file
	c.clientInbox = in
	c.clientOutbox = out
	c.engineInbox = *engineIn

	//var module BaseModule.BaseModule
	var dh DatabaseModule.DatabaseModule
	c.engineConnector.InitB(&EngineConnector{}, &dh)
	c.engineConnector.SetOutbox(&c.engineInbox)

	if !c.engineConnector.Start() {
		return
	}
	//c.engineConnector = module
	c.Inbox = *c.engineConnector.GetInbox()
	fmt.Println("Module EngineConnector loaded.")

}

func (c *ClientProcessor) run() {

}

func (c *ClientProcessor) Start() {
	go c.reciveMessages()
	go c.manageRequestsQueue()
}

func (c *ClientProcessor) reciveMessages() {
	var err error
	for {
		select {
		case msg := <-c.clientOutbox:
			err = c.requestsQueue.Push(&msg)
		}
		if err != nil {
			fmt.Println("Engine error:", err)
		}
	}
}

func (c *ClientProcessor) manageRequestsQueue() {

	for {
		if c.requestsQueue.Empty {
			<-c.requestsQueue.WakeUp
		}

		req := c.requestsQueue.Pop()

		if req.Request {
			c.manageGetRequests(*req)
		} else {
			c.manageSetRequests(*req)
		}
	}
}

func (c *ClientProcessor) manageGetRequests(req EngineTypes.Message) {
	switch req.Action {
	case "getCords":
		c.getCords(req) // sample action1 bind
	case "getArea":
		c.getArea(req) // sample action1 bind
	default:
		fmt.Println("ClientProcessor debug: Wrong action!") // if there is no proper action, report an error
	}
}

func (c *ClientProcessor) manageSetRequests(req EngineTypes.Message) {

}

//ACTIONS

func (c *ClientProcessor) getCords(req EngineTypes.Message) {
	task := EngineTypes.Task{} // create task structure
	task.Input = make(chan EngineTypes.Message, 3)
	task.Run = func() { // create functions that just replays the msg
		req.Action = "Response from clinet processor"

		var msg EngineTypes.Message               // create new request message
		msg.Action = "getCords"                   // set request action
		msg.Sender = EngineTypes.CLIENT_PROCESSOR // set sender.
		msg.Request = true                        // becouse msg is a request, change it to response
		res, _ := c.engineConnector.SendMessage(msg, task, true)
		res.Sender = req.Sender
		c.clientInbox <- res
	}
	go task.Run() // run task in background
}

func (c *ClientProcessor) getArea(req EngineTypes.Message) {
	task := EngineTypes.Task{} // create task structure
	task.Input = make(chan EngineTypes.Message, 3)
	task.Run = func() { // create functions that just replays the msg

		reqSender := req.Sender
		req.Sender = EngineTypes.CLIENT_PROCESSOR
		req.Action = "getArea"
		res, _ := c.engineConnector.SendMessage(req, task, true)
		res.Sender = reqSender

		c.clientInbox <- res
	}
	go task.Run() // run task in background
}
