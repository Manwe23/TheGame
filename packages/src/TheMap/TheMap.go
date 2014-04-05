// TheMap project TheMap.go
package TheMap

import (
	"EngineTypes"
	"database/sql"
	"fmt"
	_ "github.com/ziutek/mymysql/godrv"
	"log"
	"math/rand"
	"time"
)

type Mapa struct {
	databaseHandler EngineTypes.DataBaseHandler // Handler to database ORM ( at this moment we use gorp and mysql database )
	sendMessage     func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool)
}

type ConfigTable struct {
	Name   string
	Code   string
	DescId int
	Desc   string
	Table  map[int]Field
}

type Field struct {
	Name      string
	Code      string
	DescId    int
	Desc      string
	Exeptions []Field
}

var configurationData map[string]ConfigTable

func (m Mapa) Load() {

	var field Field
	var code string
	var id int
	var name string
	var idDescription sql.NullInt64
	var sqlQuery string
	var configTable ConfigTable
	fmt.Println("preparing...")
	con, err := sql.Open("mymysql", "mapconfigurationdata/TheGameMap/TheGameMap")
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()
	fmt.Println("passed")
	rows, err := con.Query("SELECT * FROM groups")
	if err != nil {
		log.Fatal(err)
	}

	configurationData = make(map[string]ConfigTable)

	for rows.Next() {
		if err := rows.Scan(&id, &name, &code, &idDescription); err != nil {
			log.Fatal(err)
		}
		configTable.Table = make(map[int]Field)
		configTable.Code = code
		configTable.Name = name
		if idDescription.Valid {
			configTable.DescId = int(idDescription.Int64)
		}
		configurationData[name] = configTable
	}
	if err != nil {
		log.Fatal(err)
	}
	var counter int
	for key, value := range configurationData {
		sqlQuery = fmt.Sprintf("SELECT * FROM %s\n", value.Name)
		rows, err := con.Query(sqlQuery)
		if err != nil {
			log.Fatal(err)
		}
		counter = 0
		for rows.Next() {
			if err := rows.Scan(&id, &code, &idDescription); err != nil {
				log.Fatal(err)
			}
			field.Name = name
			field.Code = code
			if idDescription.Valid {
				field.DescId = int(idDescription.Int64)
			}
			configurationData[key].Table[counter] = field
			counter++
		}
	}
}

func (m Mapa) Render() {
	fmt.Println("Rendering...")
}

func (m Mapa) Get(long int, lat int) {
	long, lat = m.GetRandomCords()
	fmt.Printf("[%d,%d]=%s\n", long, lat, m.GetRandomFieldType())
}

func (m Mapa) Set(long int, lat int, attr_name string, attr_value string) {
	fmt.Printf("[%d,%d].%s = %s\n", long, lat, attr_name, attr_value)
}

func (m Mapa) GetRandomFieldType() string {
	rand.Seed(time.Now().Unix())
	result := ""
	for _, value := range configurationData {
		result += value.Code + value.Table[rand.Intn(len(value.Table))].Code
	}
	return result
}

func (m Mapa) GetRandomCords() (long int, lat int) {
	rand.Seed(time.Now().Unix())
	long = 180 - rand.Intn(360)
	lat = 180 - rand.Intn(360)
	return
}

func (m *Mapa) InitModule(sender func(EngineTypes.Message, EngineTypes.Task, bool) (EngineTypes.Message, bool), h EngineTypes.DataBaseHandler) {
	m.databaseHandler = h
	m.sendMessage = sender
}

/* Function binds request to proper function */
func (m *Mapa) GenerateTask(req EngineTypes.Message) {
	switch req.Action {
	case "getCords":
		m.getCords(req) // sample action1 bind
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
		msg.Data["Response"] = attr1
		fmt.Println("Map debug: Send msg", msg)
		m.sendMessage(msg, task, false) // send message to engine with no wait
	}
	go task.Run() // run task in background
}
