// httpServer project httpServer.go
package HttpServer

import (
	"Config"
	"EngineTypes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type User struct {
	id           int
	login        string
	auth         string
	logged       bool
	connection   *websocket.Conn
	connectionId string
}

type HttpServer struct {
	users        map[string]User
	usersIds     map[int]string
	engineInbox  chan EngineTypes.Message
	engineOutbox chan EngineTypes.Message
}

func (s *HttpServer) Init(in chan EngineTypes.Message, out chan EngineTypes.Message) {
	s.users = make(map[string]User)
	s.usersIds = make(map[int]string)
	s.engineInbox = in
	s.engineOutbox = out
	rand.Seed(time.Now().Unix())
	s.LogIn()
}

func send(msg []byte, conn *websocket.Conn) {

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		conn.Close()

	}

}

func (s *HttpServer) sendToUser(userId int, msg EngineTypes.Message) {
	user := s.users[s.usersIds[userId]]

	send(encodeMsg(msg), user.connection)
	fmt.Println(time.Now().UnixNano())
}

func (s *HttpServer) sendToEngine(msg EngineTypes.Message) {
	s.engineInbox <- msg
	fmt.Println(time.Now().UnixNano())
}

func encodeMsg(msg EngineTypes.Message) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("error:", err)
	}
	return []byte(b)
}

func decodeMsg(rq []byte) (EngineTypes.Message, error) {
	var eMsg EngineTypes.Message
	err := json.Unmarshal(rq, &eMsg)
	if err != nil {
		fmt.Println("error:", err)
	}

	return eMsg, err
}

func setNewWebSocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, string, error) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return conn, "", errors.New("Not a websocket handshake")
	} else if err != nil {
		log.Println(err)
		return conn, "", err
	}
	log.Println("Succesfully upgraded connection")
	return conn, fmt.Sprintf("%d", rand.Int()), nil
}

func (s *HttpServer) LogIn() {
	var user User
	user.login = "Manwe"
	user.logged = true
	user.id = 23
	user.connectionId = ""
	s.users["manwe"] = user
	s.usersIds[23] = "manwe"
}

func check(user User, userID string, connectionID string) bool {
	return true
}

func (s *HttpServer) reciveMessagesFromEngine() {
	var err error
	for {
		select {
		case msg := <-s.engineOutbox:
			s.sendToUser(msg.Sender, msg)
		}
		if err != nil {
			fmt.Println("Engine error:", err)
		}
	}
}

func (s *HttpServer) RequestsMenager(rq []byte, user int) {
	eMsg, err := decodeMsg(rq)
	if err != nil {

		return
	}
	eMsg.Request = true
	eMsg.Sender = user
	s.sendToEngine(eMsg)
}

func (s *HttpServer) gameHandler(w http.ResponseWriter, req *http.Request) {
	// Taken from gorilla's website

	userID := req.FormValue("userID")
	connectionID := req.FormValue("connectionID")

	fmt.Println("connected:", userID)
	user := s.users[userID]
	if !check(user, userID, connectionID) {
		fmt.Println("Acces denied!\nLog out!")
		return
	}

	conn, newConnectionID, err := setNewWebSocketConnection(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}
	user.connection = conn
	user.connectionId = newConnectionID
	s.users[userID] = user

	for {
		// Blocks until a message is read
		_, rq, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return
		}
		s.RequestsMenager(rq, user.id)
	}
}

func (s *HttpServer) Start(in chan EngineTypes.Message, out chan EngineTypes.Message) {
	// command line flags
	port := Config.HttpServerPort
	dir := Config.HttpServerDir
	fmt.Println("http server dir: ", dir)

	s.Init(in, out)
	go s.reciveMessagesFromEngine()

	// handle all requests by serving a file of the same name
	fs := http.Dir(dir)
	fileHandler := http.FileServer(fs)
	http.Handle("/", fileHandler)
	http.HandleFunc("/ws", s.gameHandler)

	log.Printf("Running on port %d\n", port)

	addr := fmt.Sprintf("127.0.0.1:%d", port) //todo: put it into config file
	// this call blocks -- the progam runs here forever

	err := http.ListenAndServe(addr, nil)
	fmt.Println(err.Error())
}
