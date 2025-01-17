package appender

type MapAppender[K comparable, V any] struct {
	src    map[K]V
	getKey func(v V) K
}

func NewMapAppender[K comparable, V any](
	size int,
	getKey func(v V) K,
) *MapAppender[K, V] {
	if size == 0 {
		return &MapAppender[K, V]{
			src:    make(map[K]V),
			getKey: getKey,
		}
	}

	return &MapAppender[K, V]{
		src:    make(map[K]V, size),
		getKey: getKey,
	}
}

func (m *MapAppender[K, V]) Append(v V) {
	m.src[m.getKey(v)] = v
}

func (m *MapAppender[K, V]) HasKey(key K) (ok bool) {
	_, ok = m.src[key]
	return
}

func (m *MapAppender[K, V]) Len() int {
	return len(m.src)
}

func (m *MapAppender[K, V]) Map() map[K]V {
	return m.src
}

type SliceAppender[T any] struct {
	src []T
}

func NewSliceAppender[T any](len, cap int) *SliceAppender[T] {
	return &SliceAppender[T]{
		src: make([]T, len, cap),
	}
}

func (s *SliceAppender[T]) Append(v T) {
	s.src = append(s.src, v)
}

func (s *SliceAppender[T]) Len() int {
	return len(s.src)
}

func (s *SliceAppender[T]) Slice() *[]T {
	return &s.src
}
