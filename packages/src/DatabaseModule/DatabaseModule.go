// DatabaseModule project DatabaseModule.go
package DatabaseModule

import (
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
)

type DatabaseMenager struct {
	// monitor todo: Podpiąć monitor silnika do kontrolowania zapytan do bazy danych

	clientsConnections  map[int]*DatabaseModule    //aliases to connection by client name
	databaseConnections map[string]*DatabaseModule //aliases to connection by database name
}

func (menager *DatabaseMenager) Init() {
	menager.clientsConnections = make(map[int]*DatabaseModule)
	menager.databaseConnections = make(map[string]*DatabaseModule)
}

func (menager *DatabaseMenager) Create(client int) *DatabaseModule {
	var d DatabaseModule
	d.init(client, menager)
	return &d
}

func (menager *DatabaseMenager) addConnection(client int, database string, d *DatabaseModule) {
	menager.clientsConnections[client] = d
	menager.databaseConnections[database] = d
}

/*DatabaseModule*/

type DatabaseModule struct {
	clientModule int
	menager      *DatabaseMenager
	connections  map[string]*Connection
}

func (d *DatabaseModule) init(client int, menager *DatabaseMenager) {
	d.clientModule = client
	d.menager = menager
	d.connections = make(map[string]*Connection)
}

func (d *DatabaseModule) NewConnection(database string, user string, pwd string) *Connection {
	db, err := sql.Open("mymysql", fmt.Sprintf("tcp:localhost:3306*%s/%s/%s", database, user, pwd))
	if err != nil {
		fmt.Println(err)
	}
	var conn = Connection{}
	conn.init(&gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}})
	d.menager.addConnection(d.clientModule, database, d)
	d.connections[database] = &conn
	return &conn
}

func (d *DatabaseModule) Close() {
	for _, conn := range d.connections {
		conn.Close()
	}
}

/*Connection*/

type Connection struct {
	connectionHandler *gorp.DbMap
}

func (c *Connection) init(conn *gorp.DbMap) {
	c.connectionHandler = conn
}

func (c *Connection) Select(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	return c.connectionHandler.Select(i, query, args...)
}

func (c *Connection) Insert(list ...interface{}) error {
	return c.connectionHandler.Insert(list...)
}

func (c *Connection) Update(list ...interface{}) (int64, error) {
	return c.connectionHandler.Update(list...)
}

func (c *Connection) Close() {
	c.connectionHandler.Db.Close()
}
