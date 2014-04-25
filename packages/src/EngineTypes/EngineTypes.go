package EngineTypes

import (
	"DatabaseModule"
	"container/list"
	"errors"
)

type IModule interface {
	GenerateTask(Message)
	InitModule(func(Message, Task, bool) (Message, bool), *DatabaseModule.DatabaseModule)
}

type StateMessage int

const (
	START StateMessage = iota
	END
	PAUSE
)

const (
	MODULE = iota
	CLIENT_PROCESSOR
	MAP
)

type MapField struct {
	Code string
	Desc map[string]string
}

type Message struct {
	MessageId int
	Priority  int
	Sender    int
	Request   bool
	Action    string
	Data      map[string]interface{}
}

type MessageQueue struct {
	items  []*Message
	head   int
	length int
	max    int
	WakeUp chan bool
	Empty  bool
}

type Task struct {
	Kill  chan bool
	Input chan Message
	Run   func()
}

type TaskContainer struct {
	waitingTasks map[int]Task
	readyTasks   list.List
	WakeUp       chan bool
	Empty        bool
}

func (t *TaskContainer) PushTask(task Task, sleep bool, id int) int {
	if t.waitingTasks == nil {
		t.waitingTasks = make(map[int]Task)
	}

	if sleep {
		t.waitingTasks[id] = task
	} else {
		t.readyTasks.PushBack(task)
	}
	return len(t.waitingTasks)
}

func (t *TaskContainer) GetTask(id int) (Task, error) {
	task, ok := t.waitingTasks[id]
	if len(t.waitingTasks) == 0 {
		return task, errors.New("Task container is empty!")
	}
	if !ok {
		return task, errors.New("Missing task")
	}
	delete(t.waitingTasks, id)
	return task, nil
}

func (q *MessageQueue) Init(n int) {
	q.items = make([]*Message, n)
	q.max = n
	q.length = 0
	q.head = 0
	q.Empty = true
	q.WakeUp = make(chan bool)
}

func (q *MessageQueue) Pop() *Message {

	q.length--
	if q.length == 0 {
		q.Empty = true
		select {
		case <-q.WakeUp:
		default:
		}
	}
	old := q.items[q.head]
	q.items[q.head] = nil
	q.head++
	if q.head == q.max {
		q.head = 0
	}
	return old
}

func (q *MessageQueue) Push(msg *Message) error {
	if q.length == q.max {
		return errors.New("MessageQueue is full!")
	}
	q.items[(q.head+q.length)%q.max] = msg
	q.length++
	if q.Empty {
		q.Empty = false
		select {
		case q.WakeUp <- true:
		default:
		}
	}
	return nil
}

func (q MessageQueue) GetLength() int {
	return q.length
}
