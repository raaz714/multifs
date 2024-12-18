package pathiterator

import "fmt"

func (q *Queue) Enqueue(value *StrTreePair) {
	*q = append(*q, *value)
}

func (q *Queue) EnqueueArray(values []StrTreePair) {
	for _, value := range values {
		*q = append(*q, value)
	}
}

func (q *Queue) Size() int {
	return len(*q)
}

func (q *Queue) Empty() bool {
	return q.Size() == 0
}

func (q *Queue) Peek() (*StrTreePair, error) {
	if q.Size() == 0 {
		return nil, fmt.Errorf("Queue is empty")
	}
	return &(*q)[0], nil
}

func (q *Queue) Dequeue() (*StrTreePair, error) {
	if q.Size() == 0 {
		return nil, fmt.Errorf("Queue is empty")
	}
	value := &(*q)[0]
	*q = (*q)[1:]
	return value, nil
}
