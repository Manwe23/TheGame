// TheMap project TheMap.go
package TheMap

import (
	"EngineTypes"
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/ziutek/mymysql/godrv"
	"log"
	"math/rand"
	"time"
)

type Mapa struct {
	inbox           chan EngineTypes.Message
	outbox          chan EngineTypes.Message
	errors          chan error
	control         chan int
	pause           chan int
	state           int
	lastMessageId   int
	taskContainer   EngineTypes.TaskContainer
	databaseHandler *gorp.DbMap
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

func (m *Mapa) run() {

	rand.Seed(time.Now().Unix())
	for {
		select {
		case c := <-m.control:
			switch c {
			case EngineTypes.START:
				m.state = EngineTypes.START
			case EngineTypes.END:
				m.state = EngineTypes.END
				return
			case EngineTypes.PAUSE:
				m.state = EngineTypes.PAUSE
				c = <-m.control
				m.control <- c
				continue
			default:
			}
		default:
		}
		var msg EngineTypes.Message
		select {
		case msg = <-m.inbox:
			if msg.Request {
				m.generateTask(msg)
			} else {
				task, err := m.taskContainer.GetTask(msg.MessageId)
				if err != nil {
					fmt.Println("Map debug:", err)
					continue
				}
				fmt.Println("Map debug: task found")
				task.Input <- msg
			}
			fmt.Println("Map recive msg:", msg)
		default:
			random := rand.Intn(5000)
			time.Sleep(time.Duration(random) * time.Millisecond)
			m.getCords()
		}
	}
	fmt.Println("Map done.")
}

func (m *Mapa) generateTask(req EngineTypes.Message) {
	switch req.Action {
	case "GetCords":
		fmt.Println("Map debug: get cordc action")
		m.getResources(req)
	case "Echo":
		fmt.Println("Map debug: echo action")
		m.echo(req)
	default:
		fmt.Println("Map debug: Wrong action!")
	}
}

func (m *Mapa) Start() bool {
	fmt.Println("Map start.")
	m.inbox = make(chan EngineTypes.Message, 10)
	m.control = make(chan int, 10)
	m.pause = make(chan int)
	m.state = EngineTypes.START
	m.lastMessageId = 0
	go m.run()
	return true

}

func (m Mapa) End() {
	m.control <- EngineTypes.END
}

func (m Mapa) Pause(on bool) {
	if on {
		m.control <- EngineTypes.PAUSE
	} else {
		m.control <- EngineTypes.START
	}
}

func (m Mapa) Open() {

}

func (m Mapa) Close() {

}

func (m Mapa) GetState() int {
	return 0
}

func (m Mapa) GetInbox() *chan EngineTypes.Message {
	return &m.inbox
}

func (m *Mapa) SetOutbox(c *chan EngineTypes.Message) {
	m.outbox = *c
}

func (m *Mapa) SetErrorChan(c *chan error) {
	m.errors = *c
}

func (m *Mapa) SetDatabaseHandler(h *gorp.DbMap) {
	m.databaseHandler = h
}

/////////// ACTION FUNCTIONS /////////////

func (m *Mapa) getResources(req EngineTypes.Message) {
	task := EngineTypes.Task{}
	task.Kill = make(chan bool)
	task.Input = make(chan EngineTypes.Message, 3)
	task.Run = func() {
		var msg EngineTypes.Message
		msg.Action = "GetResources"
		msg.MessageId = m.lastMessageId
		msg.Sender = "Mapa"
		msg.Request = true
		m.lastMessageId++
		fmt.Println("Map debug: push msg:", msg)
		m.taskContainer.PushTask(task, true, msg.MessageId)
		m.outbox <- msg
		for {
			select {
			case <-task.Kill:
				fmt.Println("Task killed!")
				return
			case msg = <-task.Input:
				fmt.Println("Map debug: Task get msg:", msg)
				msg.Request = false
				msg.Action = "Response"
				msg.MessageId = req.MessageId
				m.outbox <- msg
				return
			}
		}
		defer close(task.Kill)
		defer close(task.Input)
	}
	go task.Run() // tutaj przydałaby się jakaś kontorla nad tym ile aktualnie mamy tasków w tle
}

func (m *Mapa) getCords() {
	task := EngineTypes.Task{}
	task.Kill = make(chan bool)
	task.Input = make(chan EngineTypes.Message, 3)
	task.Run = func() {

		var msg EngineTypes.Message
		msg.Action = "GetCords"
		msg.MessageId = m.lastMessageId
		msg.Sender = "Mapa"
		msg.Request = true
		m.taskContainer.PushTask(task, true, msg.MessageId)
		m.lastMessageId++

		fmt.Println("Map debug: push msg:", msg)
		m.taskContainer.PushTask(task, true, msg.MessageId)
		m.outbox <- msg
		for {
			select {
			case <-task.Kill:
				fmt.Println("Task killed!")
				return
			case msg = <-task.Input:
				fmt.Println("Map debug: getCords finished with msg ", msg)
				return
			}
		}
		defer close(task.Kill)
		defer close(task.Input)
	}
	go task.Run() // tutaj przydałaby się jakaś kontorla nad tym ile aktualnie mamy tasków w tle
}

func (m *Mapa) echo(msg EngineTypes.Message) {

	msg.Request = false
	msg.Action = "Response"
	m.outbox <- msg

}
