package utils

import (
	"cmp"
	"container/heap"
)

type pqItem[P cmp.Ordered, T any] struct {
	priority P
	value    T
}

type pq[P cmp.Ordered, T any] struct {
	q    []*pqItem[P, T]
	less func(a, b P) bool
}

func (pq *pq[P, T]) Len() int { return len(pq.q) }
func (pq *pq[P, T]) Less(i, j int) bool {
	return pq.less(pq.q[i].priority, pq.q[j].priority)
}

func (pq *pq[P, T]) Swap(i, j int) {
	pq.q[i], pq.q[j] = pq.q[j], pq.q[i]
}

func (pq *pq[P, T]) Push(x any) {
	pq.q = append(pq.q, x.(*pqItem[P, T]))
}

func (pq *pq[P, T]) Pop() any {
	last := len(pq.q) - 1
	item := pq.q[last]
	pq.q = pq.q[:last]
	return item
}

type PriorityQueue[P cmp.Ordered, T any] struct {
	pq *pq[P, T]
}

func NewPriorityQueue[P cmp.Ordered, T any](less func(a, b P) bool) *PriorityQueue[P, T] {
	if less == nil {
		less = func(a, b P) bool {
			return a < b
		}
	}
	pq := &pq[P, T]{
		q:    []*pqItem[P, T]{},
		less: less,
	}

	heap.Init(pq)
	return &PriorityQueue[P, T]{
		pq: pq,
	}
}

func (pq *PriorityQueue[P, T]) Len() int {
	return pq.pq.Len()
}

func (pq *PriorityQueue[P, T]) Push(priority P, value T) {
	heap.Push(pq.pq, &pqItem[P, T]{priority, value})
}

func (pq *PriorityQueue[P, T]) Pop() (T, P) {
	val := heap.Pop(pq.pq).(*pqItem[P, T])
	return val.value, val.priority
}

func (pq *PriorityQueue[P, T]) Empty() bool {
	return pq.Len() == 0
}
