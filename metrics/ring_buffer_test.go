package metrics

import "testing"

func TestRingBuffer(t *testing.T) {
	rb := newRingBuffer[int]()
	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	assertRingBufferValues(t, rb, []int{1, 2, 3})
	if rb.PopFirst() != 1 {
		t.Fatal("invalid first value")
	}
	assertRingBufferValues(t, rb, []int{2, 3})
	if rb.PopFirst() != 2 {
		t.Fatal("invalid second value")
	}
	assertRingBufferValues(t, rb, []int{3})
	if rb.PopFirst() != 3 {
		t.Fatal("invalid third value")
	}
	assertRingBufferValues(t, rb, []int{})
	rb.Push(4)
	assertRingBufferValues(t, rb, []int{4})
	rb.Push(5)
	assertRingBufferValues(t, rb, []int{4, 5})
	rb.Push(6)
	assertRingBufferValues(t, rb, []int{4, 5, 6})
	if rb.PopFirst() != 4 {
		t.Fatal("invalid value")
	}
	assertRingBufferValues(t, rb, []int{5, 6})
	rb.Push(7)
	assertRingBufferValues(t, rb, []int{5, 6, 7})
}

func assertRingBufferValues(t *testing.T, rb *ringBuffer[int], values []int) {
	if rb.Len() != len(values) {
		t.Fatal("invalid length")
	}
	n := 0
	for i, x := range rb.Iter {
		if i != n {
			t.Fatal("unexpected index")
		}
		n += 1
		if values[i] != x {
			t.Fatal("unexpected values")
		}
	}
	if n != len(values) {
		t.Fatal("unexpected iter length")
	}
}
