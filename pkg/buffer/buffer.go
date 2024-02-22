package buffer

import (
	"io"
	"moe/pkg/common"
	"moe/pkg/gapbuffer"
	"moe/ui/components/cursor"
	"os"
	"path"
)

// Buffer represents an opened file.
type Buffer struct {
	parentNode *BufferNode

	Path     string                     // Absolute path on disk.
	fd       *os.File                   // File descriptor.
	val      *gapbuffer.GapBuffer[rune] // Actual raw text data. Gap Buffer is a nice compromise between Piece Chain and buffer.
	modified bool                       // Content was modified and not saved to disk
	Cursor   cursor.Model               // Cursor position inside this buffer.
}

// NewBuffer constructs a new buffer from a path. If that file exists, it opens it for reading,
// otherwise it will just open a fake file in memory
func NewBuffer(path string) (*Buffer, error) {
	var bytes []byte

	fd, err := os.OpenFile(path, os.O_RDWR, 0664) // taken from helix
	if err == nil {
		// File exists
		bytes, err = io.ReadAll(fd)
		if err != nil {
			// Some weird error happened. Display it.
			return nil, err
		}
	}

	// Ok by this point I either have a fd with some bytes or a nil fd and nil bytes
	content := []rune(string(bytes))
	buf := gapbuffer.New(content)

	return &Buffer{
		parentNode: nil,
		Path:       path,
		fd:         fd,
		val:        &buf,
		modified:   false,
		Cursor:     cursor.New(),
	}, nil

}

// String returns the string contained in this buffer
func (b *Buffer) String() string {
	content := b.val
	if content == nil {
		return ""
	}

	return string(b.val.Collect())
}

// Name returns the title of the buffer window to display
func (b Buffer) Name() string {
	_, name := path.Split(b.Path)
	return name
}

func (b *Buffer) CursorRight() {
	pos := min(b.Cursor.Pos+1, b.val.Len())
	char := b.val.ElementAt(pos)
	b.Cursor.Char = string(char)
	b.Cursor.Pos = pos
}

func (b *Buffer) CursorLeft() {
	pos := max(b.Cursor.Pos-1, 0)
	char := b.val.ElementAt(pos)
	b.Cursor.Char = string(char)
	b.Cursor.Pos = pos
}

// The bufferline is composed of a linked-list
type BufferNode struct {
	Prev   *BufferNode
	Next   *BufferNode
	Buffer *Buffer
}

// InsertNode inserts node `n` before node `src`
func InsertNode(src *BufferNode, n *BufferNode) {
	n.Prev = src.Prev
	n.Next = src
	src.Prev.Next = n
	src.Prev = n
}

// ReplaceNode replaces node `old` with `new` in the Linked List
func ReplaceNode(old *BufferNode, new *BufferNode) {
	old.Prev.Next = new
	old.Next.Prev = new
	new.Next = old.Next
	new.Prev = old.Prev
}

// Node takes a *buffer and returns a *BufferNode
func Node(buf *Buffer) *BufferNode {
	node := BufferNode{
		Prev:   nil,
		Next:   nil,
		Buffer: buf,
	}
	buf.parentNode = &node
	return &node
}

type LinkedList struct {
	head *BufferNode
	tail *BufferNode
}

func NewList() LinkedList {
	head := &BufferNode{
		Prev:   nil,
		Next:   nil,
		Buffer: nil,
	}
	tail := &BufferNode{
		Prev:   nil,
		Next:   nil,
		Buffer: nil,
	}
	head.Next = tail
	tail.Prev = head

	return LinkedList{
		head: head,
		tail: tail,
	}
}

func (l *LinkedList) AddNode(n *BufferNode) {
	n.Prev = l.tail.Prev
	n.Next = l.tail
	l.tail.Prev.Next = n
	l.tail.Prev = n
}

// NodeIterator implements the Iterator interface for LinkedList
type NodeIterator struct {
	n *BufferNode
}

// HasNext tells us wether the iterator still has elements to consume
func (i *NodeIterator) HasNext() bool {
	next := i.n.Next
	// Only tail nodes have next equal to nil
	return next.Next != nil
}

// Next gets the next element in this iterator
func (i *NodeIterator) Next() *BufferNode {
	if i.HasNext() {
		node := i.n.Next
		i.n = node
		return i.n
	}
	return nil
}

// Creates an iterator over the LinkedList elements
func (l *LinkedList) Iter() common.Iterator[BufferNode] {
	return &NodeIterator{n: l.head}
}

/* Use this when GOEXPERIMENT=rangefunc is merged */
// func (l *LinkedList) Iter() func(func(*BufferNode) bool) {
// 	return func(yield func(*BufferNode) bool) {
// 		for node := l.head.Next; node.Next != nil; node = node.Next {
// 			if !yield(node) {
// 				return
// 			}
// 		}
// 	}
// }
