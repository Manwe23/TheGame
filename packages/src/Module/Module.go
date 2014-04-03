// Module project Module.go
package Module

import (
	"EngineTypes"
	"fmt"
	"github.com/coopernurse/gorp"
	"math/rand"
	"time"
)

/* Module type structure. Functions below require all of this.
Feel free to modify this as you want, just keep it working properly. */
type Module struct {
	inbox           chan EngineTypes.Message  // inbox channel. Place where all messages come.
	outbox          chan EngineTypes.Message  // outbox channel. Put on this channel messages which need to be send to engine.
	errors          chan error                // error channel. If you want to send this error to engine just put it on this channel.
	control         chan int                  // control channel. Methods use it to start, pause or end module work
	pause           chan int                  // additional channel to pause module
	state           int                       // Variable to represent module state. For more see EngineTypes module (START, END, PAUSE)
	lastMessageId   int                       // id of last send message
	taskContainer   EngineTypes.TaskContainer // Container with sleeping tasks
	databaseHandler *gorp.DbMap               // Handler to database ORM ( at this moment we use gorp and mysql database )
}

/* Main module loop with control. End of this functions means shutdown of module */
func (m *Module) run() {

	rand.Seed(time.Now().Unix()) // Initialize random with Unix time as a seed

	for {
		select {
		case c := <-m.control: // check is there any control message to deal with
			switch c {
			case EngineTypes.START: // use it to restart module after pause and kocham kocinke
				m.state = EngineTypes.START
			case EngineTypes.END: // use it to terminate module
				m.state = EngineTypes.END
				return
			case EngineTypes.PAUSE:
				m.state = EngineTypes.PAUSE // use it to pause module
				c = <-m.control             // wait for next control change
				m.control <- c              //  and send it back to then channel for next check
				continue
			default:
			}
		default:
		}
		timeout := make(chan bool, 1) // create timeout channel
		go func() {
			time.Sleep(1 * time.Second) // set timeout duration
			timeout <- true
		}()
		var msg EngineTypes.Message // Declare a new message
		select {                    // check if there is any waiting message in inbox
		case msg = <-m.inbox: // if there is any waiting message, get it
			if msg.Request { // if message is request
				m.generateTask(msg) // generate new task to handle this request
			} else {
				task, err := m.taskContainer.GetTask(msg.MessageId) // get sleeping task, which waits for this response
				if err != nil {                                     // check if there is an error
					fmt.Println("Module debug:", err) // if so, report it; better way to do this is to send it to engine
					continue                          //abort any further actions
				}
				task.Input <- msg // wake up task by sending him response that he waiting for
			}
		case <-timeout:
			// here you can put some default behavior that module should do while there is no request to serve
		}
	}
	// here you can put some staff to be done just before end of module
}

/* Function binds request to proper function */
func (m *Module) generateTask(req EngineTypes.Message) {
	switch req.Action {
	case "Action1":
		m.action1(req) // sample action1 bind
	case "Action2":
		m.action2(req) // sample action2 bind
	default:
		fmt.Println("Module debug: Wrong action!") // if there is no proper action, report an error
	}
}

/* Functions that stats the module */
func (m *Module) Start() bool {

	m.inbox = make(chan EngineTypes.Message, 100) // create inbox channel of size 100 (feel free to change it)
	m.control = make(chan int, 10)                // create control channel of size 10 (feel free to change it)
	m.pause = make(chan int)                      // create pause channel
	m.state = EngineTypes.START                   // set module state to START
	m.lastMessageId = 0                           // init id of message to 0
	go m.run()                                    // run module
	return true                                   // returns true when all goes ok, false otherwise

}

/* Function that end the main loop of the module and terminate it */
func (m Module) End() {
	m.control <- EngineTypes.END
}

/* Functions that pause or unpause the main loop of the module */
func (m Module) Pause(on bool) {
	if on {
		m.control <- EngineTypes.PAUSE // pause
	} else {
		m.control <- EngineTypes.START // unpouse
	}
}

/* Funtion that in future will open module on clients screed. Suspended at this moment */
func (m Module) Open() {

}

/* Funtion that in future will close module on clients screed. Suspended at this moment */
func (m Module) Close() {

}

/* Function that returns current state of the module */
func (m Module) GetState() int {
	return 0
}

/* Function that returns pointer to inbox of this module */
func (m Module) GetInbox() *chan EngineTypes.Message {
	return &m.inbox
}

/* Function that allows engine to set inbox of the engine as outbox channel of module */
func (m *Module) SetOutbox(c *chan EngineTypes.Message) {
	m.outbox = *c
}

/* Function that allows engine to set error channel of the engine as error channel of module */
func (m Module) SetErrorChan(c *chan error) {
	m.errors = *c
}

/* Function that allows engine to set databaseHandler of this module */
func (m *Module) SetDatabaseHandler(h *gorp.DbMap) {
	m.databaseHandler = h
}

/****************************ACTION FUNCTIONS********************************
Here are the definitions of action functions that runs as tasks.
Each task contains:
  Kill channel - to terminate task
  Input channel - channel where messages are sent. It also wake ups the task.
  Run - function that do the proper staff to answer the request

Below are some simple examples to present the idea.
*/

/* Here is simple task that just replays msg that was send.
   Aware that this simple actions doesn't need to be pushed to task container
   becouse it doesn't needs to get response from engine
*/
func (m *Module) action1(msg EngineTypes.Message) {
	task := EngineTypes.Task{} // create task structure
	task.Run = func() {        // create functions that just replays the msg
		msg.Request = false // becouse msg is a request, change it to response
		m.outbox <- msg     // send message to engine
	}
	go task.Run() // run task in background
}

/* Here is an example of taks that need more data from engine to finish the job */
func (m *Module) action2(req EngineTypes.Message) { // functions gets request as agrument to get request data
	task := EngineTypes.Task{}                     // create task structure
	task.Kill = make(chan bool)                    // initialize Kill channel
	task.Input = make(chan EngineTypes.Message, 3) // initialize Input channel. Menage space well!
	task.Run = func() {                            //define Run function.
		var msg EngineTypes.Message                         // create new request message
		msg.Action = "Some action"                          // set request action
		msg.MessageId = m.lastMessageId                     // set request id. Remember that the response will come back with the same id.
		msg.Sender = "Module"                               // set sender.
		msg.Request = true                                  // set that this message is request
		msg.Data = make(map[string]EngineTypes.DataTypes)   // initialize data map
		attr1 := EngineTypes.DataTypes{}                    //
		attr1.Type = "string"                               // Set some data
		attr1.String = "string value1"                      // to the request message
		msg.Data["atrr1"] = attr1                           //
		attr2 := EngineTypes.DataTypes{}                    // In future there will be no need to
		attr2.Type = "int"                                  // specify type of data.
		attr2.Int = 1                                       //
		msg.Data["atrr2"] = attr2                           //
		m.lastMessageId++                                   // increment module last message id
		m.taskContainer.PushTask(task, true, msg.MessageId) // push task to task container
		m.outbox <- msg                                     // send request
		for {                                               // now the task will wait until response will come
			select {
			case <-task.Kill: // if there will occuer kill message, task will be terminated without waiting for response
				return
			case msg = <-task.Input: // get response
				msg.Request = false           // here is again just echo
				msg.MessageId = req.MessageId // but you can even send another request to engine.
				m.outbox <- msg               // just remember to push it into task container
				return
			}
		}
		defer close(task.Kill) // always remember to free the resources
		defer close(task.Input)
	}
	go task.Run() // run task in background
}
