package utils

import "errors"

var ErrEmptySet = errors.New("Set is empty")

type Set[T comparable] struct {
	s map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		s: map[T]struct{}{},
	}
}

func (s *Set[T]) Add(v T) {
	s.s[v] = struct{}{}
}

func (s *Set[T]) Remove(v T) {
	delete(s.s, v)
}

func (s *Set[T]) Contains(v T) bool {
	_, ok := s.s[v]
	return ok
}

func (s *Set[T]) Len() int {
	return len(s.s)
}

func (s *Set[T]) Empty() bool {
	return len(s.s) == 0
}

func (s *Set[T]) Clear() {
	clear(s.s)
}

func (s *Set[T]) Pop() (T, error) {
	for k := range s.s {
		delete(s.s, k)
		return k, nil
	}

	var zero T
	return zero, ErrEmptySet
}
