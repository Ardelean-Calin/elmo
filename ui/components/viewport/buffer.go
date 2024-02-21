package viewport

// The bufferline is composed of a linked-list
type BufferNode struct {
	prev   *BufferNode
	next   *BufferNode
	buffer *buffer
}

// InsertNode inserts node `n` before node `src`
func InsertNode(src *BufferNode, n *BufferNode) {
	n.prev = src.prev
	n.next = src
	src.prev.next = n
	src.prev = n
}

// ReplaceNode replaces node `old` with `new` in the Linked List
func ReplaceNode(old *BufferNode, new *BufferNode) {
	old.prev.next = new
	old.next.prev = new
	new.next = old.next
	new.prev = old.prev
}

// Node takes a *buffer and returns a *BufferNode
func Node(buf *buffer) *BufferNode {
	node := BufferNode{
		prev:   nil,
		next:   nil,
		buffer: buf,
	}
	buf.parentNode = &node
	return &node
}
