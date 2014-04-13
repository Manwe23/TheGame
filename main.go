package main

import (
	"Engine"
	"EngineTypes"
	"HttpServer"
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/ziutek/mymysql/godrv"
)

func main() {

	engineClientInbox := make(chan EngineTypes.Message, 1000)
	clientEngineInbox := make(chan EngineTypes.Message, 1000)

	var engine Engine.Engine
	engine.Start(engineClientInbox, clientEngineInbox)

	var httpServer HttpServer.HttpServer
	httpServer.Start(clientEngineInbox, engineClientInbox)

	//newLoad()
	wait := make(chan bool)
	select {
	case <-wait:
	}
}

type Group struct {
	IdGroups      int64
	TableName     string
	Code          string
	IdDescription interface{}
	Table         []Row
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

var configurationData map[string]ConfigTable

func newLoad() {

	var sqlQuery string

	db, err := sql.Open("mymysql", "tcp:localhost:3306*mapconfigurationdata/engine/enginepassword")
	if err != nil {
		fmt.Println(err)
	}

	databaseHandler := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	defer databaseHandler.Db.Close()

	var groups []Group
	_, err = databaseHandler.Select(&groups, "select * from groups")
	if err != nil {
		fmt.Println("Select failed with err:", err)
	}

	configurationData = make(map[string]ConfigTable)
	for x, p := range groups {
		var configTable ConfigTable
		configTable.Code = p.Code
		configTable.Name = p.TableName

		sqlQuery = fmt.Sprintf("SELECT code,name FROM %s LEFT JOIN description ON %s.idDescription=description.idDescription", p.TableName, p.TableName)
		_, err = databaseHandler.Select(&configTable.Table, sqlQuery)
		configurationData[p.TableName] = configTable
		fmt.Printf("    %d: %v\n", x, configTable)
	}

}
