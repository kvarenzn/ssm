// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package optional

import "fmt"

type Optional[T any] []T

func None[T any]() Optional[T] {
	return nil
}

func Some[T any](value T) Optional[T] {
	return []T{value}
}

func (v Optional[T]) Unwrap() T {
	return v[0]
}

func (v Optional[T]) UnwrapPtr() *T {
	return &v[0]
}

func (v Optional[T]) IsNone() bool {
	return v == nil
}

func (v Optional[T]) IsSome() bool {
	return v != nil
}

func (v Optional[T]) String() string {
	if v == nil {
		return "None"
	}

	var val any = v[0]
	switch val := val.(type) {
	case string:
		return fmt.Sprintf("Some(%q)", val)
	case fmt.Stringer:
		return fmt.Sprintf("Some(%s)", val.String())
	default:
		return fmt.Sprintf("Some(%v)", val)
	}
}
