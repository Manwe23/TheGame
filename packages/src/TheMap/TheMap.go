// TheMap project TheMap.go
package TheMap

import (
	"DatabaseModule"
	"EngineTypes"
	"fmt"
	"math/rand"
	"time"
)

type Mapa struct {
	databaseHandler   *DatabaseModule.DatabaseModule // Handler to database ORM ( at this moment we use gorp and mysql database )
	sendMessage       func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool)
	configurationData map[string]ConfigTable
}

type Group struct {
	TableName string
	Code      string
	Name      string
}

type Row struct {
	Code string
	Name string
}

type ConfigTable struct {
	Name  string
	Code  string
	Desc  string
	Table []Row
}

func (m *Mapa) load() {

	var sqlQuery string

	conn := m.databaseHandler.NewConnection("mapconfigurationdata", "engine", "enginepassword") //todo: put it into config file

	var groups []Group
	_, err := conn.Select(&groups, "SELECT tableName,code,name FROM groups LEFT JOIN description ON groups.idDescription=description.idDescription")
	if err != nil {
		fmt.Println("Select failed with err:", err)
	}

	m.configurationData = make(map[string]ConfigTable)
	for x, p := range groups {
		var configTable ConfigTable
		configTable.Code = p.Code
		configTable.Name = p.TableName
		configTable.Desc = p.Name

		sqlQuery = fmt.Sprintf("SELECT code,name FROM %s LEFT JOIN description ON %s.idDescription=description.idDescription", p.TableName, p.TableName)
		_, err = conn.Select(&configTable.Table, sqlQuery)
		m.configurationData[p.TableName] = configTable
		fmt.Printf("    %d: %v\n", x, configTable)
	}

}

func (m Mapa) Render() {
	fmt.Println("Rendering...")
}

func (m Mapa) Get(long int, lat int) {
	//long, lat = m.GetRandomCords()
	//fmt.Printf("[%d,%d]=%s\n", long, lat, m.GetRandomFieldType())
}

func (m Mapa) Set(long int, lat int, attr_name string, attr_value string) {
	fmt.Printf("[%d,%d].%s = %s\n", long, lat, attr_name, attr_value)
}

func (m Mapa) GetRandomFieldType() (result EngineTypes.MapField) {
	result.Code = ""
	result.Desc = make(map[string]string)
	for _, value := range m.configurationData {
		i := rand.Intn(len(value.Table))
		result.Code += value.Code + value.Table[i].Code
		result.Desc[value.Desc] += value.Table[i].Name
	}
	return
}

func (m Mapa) GetRandomArea(width int, height int) (result [][]EngineTypes.MapField) {
	result = make([][]EngineTypes.MapField, width)

	for i := range result {
		result[i] = make([]EngineTypes.MapField, height)
		for j := range result[i] {
			result[i][j] = m.GetRandomFieldType()
		}
	}
	return
}

func (m Mapa) GetRandomCords() (long int, lat int) {

	long = 180 - rand.Intn(360)
	lat = 180 - rand.Intn(360)
	return
}

func (m *Mapa) InitModule(sender func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool), h *DatabaseModule.DatabaseModule) {
	m.databaseHandler = h
	m.sendMessage = sender
	rand.Seed(time.Now().UnixNano())
	m.load()
}

/* Function binds request to proper function */
func (m *Mapa) GenerateTask(req EngineTypes.Message) {
	switch req.Action {
	case "getCords":
		m.getCords(req) // sample action1 bind
	case "getArea":
		m.getArea(req) // sample action1 bind
	default:
		fmt.Println("Map debug: Wrong action!") // if there is no proper action, report an error
	}
}

/***********MAP ACTIONS************/

func (m *Mapa) getCords(req EngineTypes.Message) {
	task := EngineTypes.Task{} // create task structure
	task.Run = func() {
		var msg EngineTypes.Message // create functions that just replays the msg
		msg.Request = false         // becouse msg is a request, change it to response
		msg.MessageId = req.MessageId
		msg.Data = make(map[string]EngineTypes.DataTypes) // initialize data map
		attr1 := EngineTypes.DataTypes{}                  //
		attr1.Type = "string"
		long, lat := m.GetRandomCords()
		attr1.String = fmt.Sprintf("[%d,%d]", long, lat)
		msg.Data["Cords"] = attr1

		m.sendMessage(msg, task, false) // send message to engine with no wait
	}
	go task.Run() // run task in background
}

func (m *Mapa) getArea(req EngineTypes.Message) {
	task := EngineTypes.Task{}
	task.Run = func() {
		var msg EngineTypes.Message
		msg.Request = false
		msg.MessageId = req.MessageId
		msg.Data = make(map[string]EngineTypes.DataTypes)

		area := m.GetRandomArea(req.Data["width"].Int, req.Data["height"].Int)
		result := EngineTypes.DataTypes{}
		result.Type = "MapArea"
		result.MapArea = area

		msg.Data["Area"] = result
		m.sendMessage(msg, task, false)
	}
	go task.Run()
}
