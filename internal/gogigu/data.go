package gogigu

import "container/heap"

var _ heap.Interface = (*NodePriorityQueue)(nil)

type NodePriorityQueue []*Node

func NewQueue() *NodePriorityQueue {
	q := &NodePriorityQueue{}
	heap.Init(q)
	return q
}

func (q *NodePriorityQueue) Enqueue(n *Node) {
	heap.Push(q, n)
}

func (q *NodePriorityQueue) Dequeue() *Node {
	return heap.Pop(q).(*Node)
}

func (q NodePriorityQueue) Len() int {
	return len(q)
}

func (q NodePriorityQueue) Less(i, j int) bool {
	return q[i].committedAt().Before(q[j].committedAt())
}

func (q NodePriorityQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *NodePriorityQueue) Push(x interface{}) {
	*q = append(*q, x.(*Node))
}

func (q *NodePriorityQueue) Pop() interface{} {
	old := *q
	n := len(old)
	x := old[n-1]
	old[n-1] = nil
	*q = old[0 : n-1]
	return x
}
