package optional

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
