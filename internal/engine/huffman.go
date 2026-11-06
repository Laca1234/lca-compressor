package engine

import (
	"container/heap"
	"bytes"
	"encoding/binary"
	"io"
)

type Node struct {
	Byte        byte
	Freq        int
	Left, Right *Node
	IsLeaf      bool
}

type NodeHeap []*Node

func (h NodeHeap) Len() int           { return len(h) }
func (h NodeHeap) Less(i, j int) bool { return h[i].Freq < h[j].Freq }
func (h NodeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *NodeHeap) Push(x interface{}) {
	*h = append(*h, x.(*Node))
}

func (h *NodeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type Code struct {
	Val   uint32
	NBits uint8
}

func BuildHuffmanTree(data []byte) (*Node, map[byte]Code) {
	if len(data) == 0 {
		return nil, make(map[byte]Code)
	}

	freqs := make(map[byte]int)
	for _, b := range data {
		freqs[b]++
	}

	h := &NodeHeap{}
	heap.Init(h)
	for b, f := range freqs {
		heap.Push(h, &Node{Byte: b, Freq: f, IsLeaf: true})
	}

	if h.Len() == 1 {
		root := &Node{Left: heap.Pop(h).(*Node)}
		codes := make(map[byte]Code)
		codes[root.Left.Byte] = Code{Val: 0, NBits: 1}
		return root, codes
	}

	for h.Len() > 1 {
		left := heap.Pop(h).(*Node)
		right := heap.Pop(h).(*Node)
		parent := &Node{
			Freq:   left.Freq + right.Freq,
			Left:   left,
			Right:  right,
			IsLeaf: false,
		}
		heap.Push(h, parent)
	}

	root := heap.Pop(h).(*Node)
	codes := make(map[byte]Code)
	buildCodes(root, 0, 0, codes)
	return root, codes
}

func buildCodes(n *Node, currentVal uint32, depth uint8, codes map[byte]Code) {
	if n == nil {
		return
	}
	if n.IsLeaf {
		codes[n.Byte] = Code{Val: currentVal, NBits: depth}
		return
	}
	buildCodes(n.Left, (currentVal << 1), depth+1, codes)
	buildCodes(n.Right, (currentVal << 1) | 1, depth+1, codes)
}

// SerializeTree serializes the Huffman tree structure for block header embedding
func SerializeTree(root *Node, w io.Writer) error {
	if root == nil {
		return binary.Write(w, binary.BigEndian, uint8(0))
	}
	if root.IsLeaf {
		if err := binary.Write(w, binary.BigEndian, uint8(1)); err != nil {
			return err
		}
		return binary.Write(w, binary.BigEndian, root.Byte)
	}
	if err := binary.Write(w, binary.BigEndian, uint8(2)); err != nil {
		return err
	}
	if err := SerializeTree(root.Left, w); err != nil {
		return err
	}
	return SerializeTree(root.Right, w)
}

func DeserializeTree(r io.Reader) (*Node, error) {
	var nodeType uint8
	if err := binary.Read(r, binary.BigEndian, &nodeType); err != nil {
		return nil, err
	}
	if nodeType == 0 {
		return nil, nil
	}
	if nodeType == 1 {
		var b byte
		if err := binary.Read(r, binary.BigEndian, &b); err != nil {
			return nil, err
		}
		return &Node{Byte: b, IsLeaf: true}, nil
	}
	left, err := DeserializeTree(r)
	if err != nil {
		return nil, err
	}
	right, err := DeserializeTree(r)
	if err != nil {
		return nil, err
	}
	return &Node{Left: left, Right: right, IsLeaf: false}, nil
}
