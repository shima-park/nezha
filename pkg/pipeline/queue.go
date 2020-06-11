package pipeline

import "container/list"

type Queue struct {
	l *list.List
}

func NewQueue() *Queue {
	return &Queue{l: list.New()}
}

func (q *Queue) Enqueue(v interface{}) {
	q.l.PushFront(v)
}

func (q *Queue) Dequeue() interface{} {
	e := q.l.Back()
	v := e.Value
	q.l.Remove(e)
	return v
}

func (q *Queue) Len() int {
	return q.l.Len()
}
