// Engine project Engine.go
package Engine

import (
	"EngineTypes"
	"Module"
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/ziutek/mymysql/godrv"
	"time"
)

type Engine struct {
	modules         map[string]IModule
	requestsQueue   EngineTypes.MessageQueue
	responsesQueue  EngineTypes.MessageQueue
	inbox           chan EngineTypes.Message
	taskContainer   EngineTypes.TaskContainer
	lastMessageId   int
	databaseHandler *gorp.DbMap
}

type IModule interface {
	Start() bool
	End()
	Pause(bool)
	Open()
	Close()
	GetState() int
	GetInbox() *chan EngineTypes.Message
	SetOutbox(*chan EngineTypes.Message)
	SetErrorChan(*chan error)
	SetDatabaseHandler(*gorp.DbMap)
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
	e.modules = make(map[string]IModule)
	e.modules["Module"] = &Module.Module{}
	e.modules["Module"].SetOutbox(&e.inbox)
	e.modules["Module"].SetDatabaseHandler(e.databaseHandler)
	if !e.modules["Module"].Start() {
		return false
	}
	time.Sleep(1000 * time.Millisecond)
	return true
}

func (e *Engine) reciveMessages() {

	for {
		select {
		case msg := <-e.inbox:
			if msg.Request {
				e.requestsQueue.Push(&msg)
			} else {
				e.responsesQueue.Push(&msg)
			}

		}
	}
}

func (e *Engine) sendMessage(m IModule, msg EngineTypes.Message) {

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
		case "Echo":
			e.echo(*req)
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
	db, err := sql.Open("mymysql", "tcp:localhost:3306*mydb/myuser/mypassword")
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

func (e *Engine) echo(msg EngineTypes.Message) {
	task := EngineTypes.Task{} // create task structure
	task.Run = func() {        // create functions that just replays the msg
		msg.Request = false                       // becouse msg is a request, change it to response
		e.sendMessage(e.modules[msg.Sender], msg) // send message to engine
	}
	go task.Run() // run task in background
}
