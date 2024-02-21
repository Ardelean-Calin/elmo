package buffer

import (
	"io"
	"moe/ui/components/cursor"
	"os"
	"path"
	"strings"
)

// Buffer represents an opened file.
type Buffer struct {
	parentNode *BufferNode

	Path     string       // Absolute path on disk.
	fd       *os.File     // File descriptor.
	val      *[][]rune    // Actual raw text data. TODO: Piece Chain.
	modified bool         // Content was modified and not saved to disk
	cursor   cursor.Model // Cursor position inside this buffer.
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
	s := string(bytes)
	var content [][]rune
	for _, line := range strings.Split(s, "\n") {
		content = append(content, []rune(line))
	}

	return &Buffer{
		parentNode: nil,
		Path:       path,
		fd:         fd,
		val:        &content,
		modified:   false,
		cursor:     cursor.New(),
	}, nil

}

// String returns the string contained in this buffer
func (b *Buffer) String() string {
	content := b.val
	if content == nil {
		return ""
	}

	var sb strings.Builder
	for _, r := range *content {
		for _, v := range r {
			sb.WriteRune(v)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// Name returns the title of the buffer window to display
func (b Buffer) Name() string {
	_, name := path.Split(b.Path)
	return name
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

func (l *LinkedList) Iter() func(func(*BufferNode) bool) {
	return func(yield func(*BufferNode) bool) {
		for node := l.head.Next; node.Next != nil; node = node.Next {
			if !yield(node) {
				return
			}
		}
	}
}

func (l *LinkedList) AddNode(n *BufferNode) {
	n.Prev = l.tail.Prev
	n.Next = l.tail
	l.tail.Prev.Next = n
	l.tail.Prev = n
}
