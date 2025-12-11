package utils_test

import (
	"testing"

	"github.com/kvarenzn/ssm/utils"
)

func assertEqual[T comparable](t *testing.T, val, expected T) {
	if val != expected {
		t.Errorf("Expected %v, but got %v", expected, val)
	}
}

func TestBasic01(t *testing.T) {
	pq := utils.NewPriorityQueue[float64, int](nil)
	pq.Push(0, 1)
	pq.Push(3, 23)
	pq.Push(2, 34)
	pq.Push(1, 55)
	assertEqual(t, pq.Pop(), 1)
	assertEqual(t, pq.Pop(), 55)
	assertEqual(t, pq.Pop(), 34)
	assertEqual(t, pq.Pop(), 23)
}
