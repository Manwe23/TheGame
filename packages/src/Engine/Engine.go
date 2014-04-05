// Engine project Engine.go
package Engine

import (
	"BaseModule"
	"EngineTypes"
	"Module"
	"TheMap"
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/ziutek/mymysql/godrv"
	"time"
)

type Engine struct {
	modules         map[string]BaseModule.BaseModule
	requestsQueue   EngineTypes.MessageQueue
	responsesQueue  EngineTypes.MessageQueue
	inbox           chan EngineTypes.Message
	taskContainer   EngineTypes.TaskContainer
	lastMessageId   int
	databaseHandler *gorp.DbMap
}

func (e *Engine) run() {

	go e.reciveMessages()
	go e.taskGenerator()
	go e.taskMenager()

	for {
		time.Sleep(1000 * time.Millisecond)
		// here will be some debug information
	}
}

func (e *Engine) loadModules() bool {
	e.modules = make(map[string]BaseModule.BaseModule)
	var module BaseModule.BaseModule
	module.InitB(&Module.Module{}, e.databaseHandler)
	module.SetOutbox(&e.inbox)

	if !module.Start() {

		return false
	}
	e.modules["Module"] = module
	fmt.Println("Module Module loaded.")

	time.Sleep(1000 * time.Millisecond)

	var mapa BaseModule.BaseModule
	mapa.InitB(&TheMap.Mapa{}, e.databaseHandler)
	mapa.SetOutbox(&e.inbox)

	if !mapa.Start() {
		return false
	}
	e.modules["Mapa"] = mapa
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
		default:
			fmt.Println("Engine debug:Wrong action!")

		}
	}

}

func (e *Engine) taskMenager() {

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

func (e *Engine) Init() {
	e.inbox = make(chan EngineTypes.Message, 100)

	e.responsesQueue = EngineTypes.MessageQueue{}
	e.responsesQueue.Init(100)

	e.requestsQueue = EngineTypes.MessageQueue{}
	e.requestsQueue.Init(100)

	e.lastMessageId = 0

	e.initDb()
}

func (e *Engine) initDb() {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("mymysql", "tcp:localhost:3306*mapconfigurationdata/engine/enginepassword")
	if err != nil {
		fmt.Println(err)
	}
	// construct a gorp DbMap
	e.databaseHandler = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

}

func (e Engine) Start() {
	fmt.Println("Engine:Starting...")
	e.Init()
	if !e.loadModules() {
		fmt.Println("Engine:Failed to load modules!\nBye.")
		return
	}
	go e.run()

}

/////////// ACTION FUNCTIONS /////////////

func (e *Engine) getCords(req EngineTypes.Message) {
	task := EngineTypes.Task{} // create task structure
	task.Input = make(chan EngineTypes.Message, 3)
	task.Run = func() { // create functions that just replays the msg
		req.Request = false // becouse msg is a request, change it to response
		var msg EngineTypes.Message
		msg.Action = "getCords"
		msg.Request = true
		e.lastMessageId++
		msg.MessageId = e.lastMessageId
		e.sendMessage(e.modules["Mapa"], msg)
		e.taskContainer.PushTask(task, true, msg.MessageId)
		fmt.Println("Debug1")
		res := <-task.Input
		res.MessageId = req.MessageId
		e.sendMessage(e.modules[req.Sender], res) // send message to engine
	}
	go task.Run() // run task in background
}
