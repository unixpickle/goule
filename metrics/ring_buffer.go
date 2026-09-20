package metrics

type ringBuffer[T any] struct {
	start  int
	len    int
	buffer []T
}

func newRingBuffer[T any]() *ringBuffer[T] {
	return &ringBuffer[T]{}
}

func (r *ringBuffer[T]) Len() int {
	return r.len
}

func (r *ringBuffer[T]) At(idx int) T {
	if idx < 0 || idx > r.len {
		panic("index out of bounds")
	}
	return r.buffer[(idx+r.start)%len(r.buffer)]
}

func (r *ringBuffer[T]) Set(idx int, value T) {
	if idx < 0 || idx > r.len {
		panic("index out of bounds")
	}
	r.buffer[(idx+r.start)%len(r.buffer)] = value
}

func (r *ringBuffer[T]) First() T {
	if r.len == 0 {
		panic("ring buffer is empty")
	}
	return r.At(0)
}

func (r *ringBuffer[T]) Last() T {
	if r.len == 0 {
		panic("ring buffer is empty")
	}
	return r.At(r.len - 1)
}

func (r *ringBuffer[T]) PopFirst() T {
	if r.len == 0 {
		panic("ring buffer is empty")
	}
	value := r.buffer[r.start%len(r.buffer)]
	var zero T
	r.buffer[r.start] = zero
	r.start = (r.start + 1) % len(r.buffer)
	r.len -= 1
	return value
}

func (r *ringBuffer[T]) Push(x T) {
	if r.len == len(r.buffer) {
		newBuffer := make([]T, len(r.buffer)*2+1)
		for i := 0; i < r.len; i++ {
			newBuffer[i] = r.At(i)
		}
		r.buffer = newBuffer
		r.start = 0
	}
	r.buffer[(r.start+r.len)%len(r.buffer)] = x
	r.len += 1
}

func (r *ringBuffer[T]) Iter(yield func(idx int, v T) bool) {
	for i := 0; i < r.Len(); i++ {
		if !yield(i, r.At(i)) {
			return
		}
	}
}
