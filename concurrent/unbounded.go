package concurrent

import (
	"sync"
)

/**** YOU CANNOT MODIFY ANY OF THE FOLLOWING INTERFACES/TYPES ********/
type Task interface{}

type DEQueue interface {
	PushBottom(task Task)
	IsEmpty() bool //returns whether the queue is empty
	PopTop() Task
	PopBottom() Task
	QLen() int
}

//NOTE: I added a QLen func to the interface as per advice in Slack.

/******** DO NOT MODIFY ANY OF THE ABOVE INTERFACES/TYPES *********************/

// NewUnBoundedDEQueue returns an empty UnBoundedDEQueue
func NewUnBoundedDEQueue() DEQueue {
	sentinel := &Node{}
	dequeue := &Dequeue{head: sentinel, tail: sentinel}
	dequeue.head.next = dequeue.tail
	dequeue.tail.prev = dequeue.head
	return dequeue
}

type Node struct {
	task interface{}
	prev *Node
	next *Node
}

type Dequeue struct {
	mu   sync.Mutex
	head *Node
	tail *Node
	len  int
}

func (q *Dequeue) PushBottom(task Task) {

	q.mu.Lock()
	defer q.mu.Unlock()

	taskNode := &Node{task: task}

	originalHead := q.head.next
	originalHead.prev = taskNode
	taskNode.prev = q.head
	q.head.next = taskNode
	taskNode.next = originalHead
	q.len += 1
}

func (q *Dequeue) PopBottom() Task {

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.head.next == q.tail {
		return nil
	}

	taskNode := q.head.next
	q.head.next = taskNode.next
	taskNode.next.prev = q.head
	q.len -= 1

	return taskNode.task
}

func (q *Dequeue) PopTop() Task {

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.tail.prev == q.head {
		return nil
	}

	taskNode := q.tail.prev
	q.tail.prev = taskNode.prev
	taskNode.prev.next = q.tail

	q.len -= 1
	return taskNode.task
}

func (q *Dequeue) QLen() int {

	q.mu.Lock()
	defer q.mu.Unlock()

	len := q.len
	return len
}

func (q *Dequeue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.len == 0 {
		return true
	}
	return false

}
