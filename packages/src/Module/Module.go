// Module project Module.go
package Module

import (
	"DatabaseModule"
	"EngineTypes"
	"fmt"
)

type Module struct {
	databaseHandler *DatabaseModule.DatabaseModule                                                // Handler to database ORM ( at this moment we use gorp and mysql database )
	sendMessage     func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool) // function to send messages to engine
}

func (m *Module) InitModule(sender func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool), h *DatabaseModule.DatabaseModule) {
	m.databaseHandler = h  // Seting database handler
	m.sendMessage = sender //Seting sendMessage function
}

/* Function binds request to proper function */
func (m *Module) GenerateTask(req EngineTypes.Message) {
	switch req.Action {
	case "Action1":
		m.action1(req) // sample action1 bind
	case "Action2":
		m.action2(req) // sample action2 bind
	default:
		fmt.Println("Module debug: Wrong action!") // if there is no proper action, report an error
	}
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
		defer close(task.Kill) // always remember to free the resources
		defer close(task.Input)
		msg.Request = false             // becouse msg is a request, change it to response
		m.sendMessage(msg, task, false) // send message to engine with no wait
	}
	go task.Run() // run task in background
}

/* Here is in example of request to get map cords */
func (m *Module) getCords() {
	task := EngineTypes.Task{} // create task structure
	task.Run = func() {        // create functions that just replays the msg
		defer close(task.Kill) // always remember to free the resources
		defer close(task.Input)
		var msg EngineTypes.Message             // create new request message
		msg.Action = "getCords"                 // set request action
		msg.Sender = EngineTypes.MODULE         // set sender.
		msg.Request = true                      // becouse msg is a request, change it to response
		_, ok := m.sendMessage(msg, task, true) // send message
		if !ok {
			return
		}

	}
	go task.Run() // run task in background
}

/* Here is an example of taks that need more data from engine to finish the job */
func (m *Module) action2(req EngineTypes.Message) { // functions gets request as agrument to get request data
	task := EngineTypes.Task{} // create task structure
	task.Run = func() {        //define Run function.
		defer close(task.Kill) // always remember to free the resources
		defer close(task.Input)
		var msg EngineTypes.Message             // create new request message
		msg.Action = "Some action"              // set request action
		msg.Sender = EngineTypes.MODULE         // set sender.
		msg.Request = true                      // set that this message is request
		msg.Data = make(map[string]interface{}) // initialize data map
		msg.Data["atrr1"] = "string value1"     //
		msg.Data["atrr2"] = 1                   //
		res, ok := m.sendMessage(msg, task, true)
		if !ok {
			return
		}
		res.Request = false // here is again just echo
		res.MessageId = req.MessageId
		m.sendMessage(res, task, false)
	}
	go task.Run() // run task in background
}
