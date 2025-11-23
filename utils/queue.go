package utils

import (
	"errors"
)

var ErrEmptyQueue = errors.New("Queue is empty")

type Queue[T any] struct {
	q          []T
	head, tail int
}

func NewQueue[T any](size int) *Queue[T] {
	return &Queue[T]{
		q: make([]T, size+1), // keep a empty slot
	}
}

func (q *Queue[T]) Empty() bool {
	return q.head == q.tail
}

func (q *Queue[T]) Cap() int {
	return len(q.q) - 1
}

func (q *Queue[T]) Available() int {
	return q.Cap() - q.Len()
}

func (q *Queue[T]) Len() int {
	size := len(q.q)
	return (q.tail + size - q.head) % size
}

func (q *Queue[T]) next(i int) int {
	return (i + 1) % len(q.q)
}

func (q *Queue[T]) full() bool {
	return q.next(q.tail) == q.head
}

func (q *Queue[T]) expand(size int) {
	oldSize := len(q.q)
	newQ := make([]T, oldSize+size)
	if q.head <= q.tail {
		copy(newQ, q.q[q.head:q.tail])
	} else {
		n := copy(newQ, q.q[q.head:])
		copy(newQ[n:], q.q[:q.tail])
	}

	q.q = newQ
	q.head = 0
	q.tail = q.Len()
}

func (q *Queue[T]) Push(v T) {
	if q.full() {
		expandSize := max(10, len(q.q)/2)
		q.expand(expandSize)
	}

	q.q[q.tail] = v
	q.tail = q.next(q.tail)
}

func (q *Queue[T]) Peek() (T, error) {
	var zero T
	if q.Empty() {
		return zero, ErrEmptyQueue
	}

	return q.q[q.head], nil
}

func (q *Queue[T]) Pop() (T, error) {
	var zero T
	if q.Empty() {
		return zero, ErrEmptyQueue
	}

	result := q.q[q.head]
	q.q[q.head] = zero
	q.head = q.next(q.head)
	return result, nil
}
