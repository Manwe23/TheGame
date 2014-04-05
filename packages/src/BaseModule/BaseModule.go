// BaseModule project BaseModule.go
package BaseModule

import (
	"EngineTypes"
	"fmt"
	"math/rand"
	"time"
)

/* BaseModule type structure. Functions below require all of this.
Feel free to modify this as you want, just keep it working properly. */
type BaseModule struct {
	Inbox         chan EngineTypes.Message      // inbox channel. Place where all messages come.
	Outbox        chan EngineTypes.Message      // outbox channel. Put on this channel messages which need to be send to engine.
	Errors        chan error                    // error channel. If you want to send this error to engine just put it on this channel.
	control       chan EngineTypes.StateMessage // control channel. Methods use it to start, pause or end module work
	pause         chan int                      // additional channel to pause module
	state         EngineTypes.StateMessage      // Variable to represent module state. For more see EngineTypes module (START, END, PAUSE)
	lastMessageId int                           // id of last send message
	taskContainer EngineTypes.TaskContainer     // Container with sleeping tasks
	extension     EngineTypes.IModule           // Module extension specifies the behavior of the module
}

/* Main module loop with control. End of this functions means shutdown of module */
func (m *BaseModule) run() {

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
		case msg = <-m.Inbox: // if there is any waiting message, get it
			if msg.Request { // if message is request
				m.extension.GenerateTask(msg) // generate new task to handle this request
			} else {
				task, err := m.taskContainer.GetTask(msg.MessageId) // get sleeping task, which waits for this response
				if err != nil {                                     // check if there is an error
					fmt.Println("BaseModule debug:", err) // if so, report it; better way to do this is to send it to engine
					continue                              //abort any further actions
				}
				task.Input <- msg // wake up task by sending him response that he waiting for
			}
		case <-timeout:
			// here you can put some default behavior that module should do while there is no request to serve
		}
	}
	// here you can put some staff to be done just before end of module
}

func (m *BaseModule) InitB(im EngineTypes.IModule, h EngineTypes.DataBaseHandler) {
	m.Inbox = make(chan EngineTypes.Message, 100)       // create inbox channel of size 100 (feel free to change it)
	m.control = make(chan EngineTypes.StateMessage, 10) // create control channel of size 10 (feel free to change it)
	m.pause = make(chan int)                            // create pause channel
	m.state = EngineTypes.START                         // set module state to START
	m.lastMessageId = 0                                 // init id of message to 0
	m.extension = im                                    // specifies the behavior
	m.extension.InitModule(m.SendMessage, h)            // initialize extension module
}

/* Functions that stats the module */
func (m *BaseModule) Start() bool {

	go m.run()  // run module
	return true // returns true when all goes ok, false otherwise

}

/* Function that end the main loop of the module and terminate it */
func (m BaseModule) End() {
	m.control <- EngineTypes.END
}

/* Functions that pause or unpause the main loop of the module */
func (m BaseModule) Pause(on bool) {
	if on {
		m.control <- EngineTypes.PAUSE // pause
	} else {
		m.control <- EngineTypes.START // unpouse
	}
}

/* Funtion that in future will open module on clients screed. Suspended at this moment */
func (m BaseModule) Open() {

}

/* Funtion that in future will close module on clients screed. Suspended at this moment */
func (m BaseModule) Close() {

}

/* Function that returns current state of the module */
func (m BaseModule) GetState() int {
	return 0
}

/* Function that returns pointer to inbox of this module */
func (m BaseModule) GetInbox() *chan EngineTypes.Message {
	return &m.Inbox
}

/* Function that allows engine to set inbox of the engine as outbox channel of module */
func (m *BaseModule) SetOutbox(c *chan EngineTypes.Message) {
	m.Outbox = *c
}

/* Function that allows engine to set error channel of the engine as error channel of module */
func (m BaseModule) SetErrorChan(c *chan error) {
	m.Errors = *c
}

func (m *BaseModule) SendMessage(msg EngineTypes.Message, task EngineTypes.Task, wait bool) (EngineTypes.Message, bool) {
	task.Kill = make(chan bool)                    // initialize Kill channel
	task.Input = make(chan EngineTypes.Message, 3) // initialize Input channel
	if msg.Request {
		m.lastMessageId++               // increment message id.
		msg.MessageId = m.lastMessageId // assign unique messageId to the request
	}

	if wait || msg.Request {
		m.taskContainer.PushTask(task, true, msg.MessageId)
	}

	m.Outbox <- msg // send message to the engine

	for { // now the task will wait until response will come
		select {
		case <-task.Kill: // if there will occuer kill message, task will be terminated without waiting for response
			return msg, false
		case msg = <-task.Input: // get response
			return msg, true
		}
	}
}
