package graphlib

import (
	"math"
)

// -----------------------------
// internal helpers
// -----------------------------

// Queue is a simple FIFO queue.
type Queue[T any] interface {
	pop() T    // remove and return the front element (panics on empty)
	enqueue(T) // add an element to the back
	empty() bool
	peek() T // look at the front element without removing it (panics on empty)
}

// QueueImpl is a circular‑buffer implementation of Queue.
type QueueImpl[T any] struct {
	buf        []T // backing storage
	head, tail int // index of first element, index *after* last element
	size       int // number of valid elements
	// note: we could pass a heuristic function here to calculate the growth
}

// NewQueue creates a queue with at least the given initial capacity.
func NewQueue[T any](capacity int) *QueueImpl[T] {
	if capacity < 1 {
		capacity = 1
	}
	return &QueueImpl[T]{buf: make([]T, capacity)}
}

// empty reports whether the queue has no elements.
func (q *QueueImpl[T]) empty() bool { return q.size == 0 }

// enqueue adds an element to the back of the queue.
func (q *QueueImpl[T]) enqueue(x T) {
	if q.size == len(q.buf) { // full — grow (×2)
		q.grow()
	}
	q.buf[q.tail] = x
	q.tail = (q.tail + 1) % len(q.buf)
	q.size++
}

// pop removes and returns the element at the front of the queue.
// It panics if the queue is empty.
func (q *QueueImpl[T]) pop() T {
	if q.empty() {
		panic("queue is empty")
	}
	v := q.buf[q.head]
	// Clear the slot so the value can be GC’d quickly (optional).
	var zero T
	q.buf[q.head] = zero

	q.head = (q.head + 1) % len(q.buf)
	q.size--
	return v
}

// peek returns (without removing) the element at the front of the queue.
// It panics if the queue is empty.
func (q *QueueImpl[T]) peek() T {
	if q.empty() {
		panic("queue is empty")
	}
	return q.buf[q.head]
}

// grow doubles the capacity while keeping the logical order intact.
func (q *QueueImpl[T]) grow() {
	// appending to the existing slice (and moving the contents) would be the same
	// as doing it manually. But in this implementation we still have the chance to control
	// the newCap value.

	newCap := int(math.Ceil(float64(len(q.buf)) * 1.5)) // todo chance to add another heuristic
	if newCap == 0 {
		newCap = 1
	}
	newBuf := make([]T, newCap)

	// Copy elements in logical order: [head..end) + [0..tail)
	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}
	q.buf = newBuf
	q.head = 0
	q.tail = q.size
}
