// Hero project Hero.go
package Hero

import (
	"DatabaseModule"
	"EngineTypes"
	"fmt"
)

type HeroModule struct {
	databaseHandler *DatabaseModule.DatabaseModule                                                // Handler to database ORM ( at this moment we use gorp and mysql database )
	sendMessage     func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool) // function to send messages to engine

	heroList map[int]Hero

	// min and max attribValue constants
	minAttribValue map[string]int
	maxAttribValue map[string]int
}

type Hero struct {
	attributes map[string]int
}

func (hero *Hero) InitHero(m *HeroModule) {
	// for now, all atributes are given a constant value
	// eventually players might be able to give their attribute points distribution
	for attribName := range m.minAttribValue {
		hero.attributes[attribName] = (m.minAttribValue[attribName] + m.maxAttribValue[attribName]) / 2
	}
}

func (m *HeroModule) InitModule(sender func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool), h *DatabaseModule.DatabaseModule) {
	m.databaseHandler = h  // Seting database handler
	m.sendMessage = sender //Seting sendMessage function

	// set min, max attribute values
	m.minAttribValue = map[string]int{
		"strength": 1,
	}
	m.maxAttribValue = map[string]int{
		"strength": 10,
	}
}

/* Function binds request to proper function */
func (m *HeroModule) GenerateTask(req EngineTypes.Message) {
	switch req.Action {
	case "changeAttribute":
		m.changeAttribute(req)
	case "getAttribute":
		m.getAttribute(req)
	default:
		fmt.Println("HeroModule debug: Wrong action!") // if there is no proper action, report an error
	}
}

/****************************ACTION FUNCTIONS*********************************/

/* changeAttribute action handler
   Data variable has a format Data[attribName] = changeValue, where attribName is a string
   and a changeValuealue is an int value said attribute has to be modified by. Resulting
   attribute value must lay between min and max values defined in minAttribValue,
   maxAttribValue respectively.
   No response is send.
*/
func (m *HeroModule) changeAttribute(msg EngineTypes.Message) {
	task := EngineTypes.Task{} // create task structure
	task.Run = func() {        // create functions that just replays the msg
		defer close(task.Kill) // always remember to free the resources
		defer close(task.Input)
		msg.Request = false             // becouse msg is a request, change it to response
		m.sendMessage(msg, task, false) // send message to engine with no wait

		// changeAttribute(s) value(s)
		// assuming Data field in received has
		//     Data[attribName]valueModifier
		//     Data["hero"]heroId
		heroPtr := m.heroList[msg.Data["hero"].(int)]
		for attribName, attribValChange := range msg.Data {
			if attribName != "hero" {
				heroPtr.attributes[attribName] += attribValChange.(int)
				if m.maxAttribValue[attribName] < heroPtr.attributes[attribName] {
					heroPtr.attributes[attribName] = m.maxAttribValue[attribName]
				}
				if m.minAttribValue[attribName] > heroPtr.attributes[attribName] {
					heroPtr.attributes[attribName] = m.minAttribValue[attribName]
				}
			}
		}
	}
	go task.Run() // run task in background
}

func (m *HeroModule) getAttribute(req EngineTypes.Message) {
	task := EngineTypes.Task{}
	task.Run = func() {
		var msg EngineTypes.Message
		msg.Request = false
		msg.MessageId = req.MessageId
		msg.Data = make(map[string]interface{})

		// similar to changeAttribute, Data field in req has
		// 	Data[attribName]none - maybe multiple attributes
		// 	Data["hero"]heroId

		heroPtr := m.heroList[req.Data["hero"].(int)]
		msg.Data["hero"] = req.Data["hero"]
		for attribName, _ := range req.Data {
			if attribName != "hero" {
				msg.Data[attribName] = heroPtr.attributes[attribName]
			}
		}

		m.sendMessage(msg, task, false)
	}
	go task.Run()
}
