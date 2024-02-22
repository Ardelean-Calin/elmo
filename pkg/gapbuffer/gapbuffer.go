package gapbuffer

import (
	"flag"
)

// Gap Buffer implementation. See: https://routley.io/posts/gap-buffer
type GapBuffer[T any] struct {
	buffer   []T // NOTE: In the future I might want this to be a pointer. I avoided this for now as it looked ugly
	gapStart int // Index of the first character *in* the gap
	gapEnd   int // Index of the first character *after* the gap
}

// New creates a new gapbuffer and populates it with content
func New[T any](content []T) GapBuffer[T] {
	return GapBuffer[T]{
		buffer:   content,
		gapStart: 0,
		gapEnd:   0,
	}
}

// Collect returns a slice of the content of the gap buffer, without gap
func (b *GapBuffer[T]) Collect() []T {
	var dest []T
	dest = append(dest, b.buffer[:b.gapStart]...)
	dest = append(dest, b.buffer[b.gapEnd:]...)
	return dest
}

// gapSize returns the gap size
func (b *GapBuffer[T]) gapSize() int {
	return b.gapEnd - b.gapStart
}

func (b *GapBuffer[T]) growGap() {
	// Gap size should be up to 5% of the total buffer size, but at least 64 bytes.
	// I've taken this from https://shorturl.at/FKOUZ
	gapSize := max(len(b.buffer)/20, 64)
	// In case of testing we use a smaller size to make it easier.
	if flag.Lookup("test.v") != nil {
		gapSize = 5
	}
	newBuffer := make([]T, len(b.buffer)+gapSize)

	copy(newBuffer, b.buffer[:b.gapStart])
	copy(newBuffer[b.gapEnd+gapSize:], b.buffer[b.gapEnd:])

	b.buffer = newBuffer
	b.gapEnd = b.gapStart + gapSize
}

// Len returns the size of the raw text
func (b *GapBuffer[T]) Len() int {
	return len(b.buffer) - b.gapSize()
}

// TotalLen returns the total length of the gap buffer (including gaps)
func (b *GapBuffer[T]) TotalLen() int {
	return len(b.buffer)
}

// ElementAt returns the element at the given position.
func (b *GapBuffer[T]) ElementAt(pos int) T {
	if pos > b.gapStart && b.gapSize() != 0 {
		pos = (b.gapStart - pos) + b.gapEnd
	}

	return b.buffer[pos]
}

// CursorGoto moves the cursor to the given position
func (b *GapBuffer[T]) CursorGoto(pos int) {
	panic("Unimplemented")
}

// CursorRight moves the cursor left one character.
func (b *GapBuffer[T]) CursorLeft() {
	// We are already at the start!
	if b.gapStart == 0 {
		return
	}

	// The first element before the start of the gap gets copied after the gap
	// [abc_____] becomes [ab_____c]
	b.buffer[b.gapEnd-1] = b.buffer[b.gapStart-1]
	b.gapStart--
	b.gapEnd--
}

// CursorRight moves the cursor right one character.
func (b *GapBuffer[T]) CursorRight() {
	// We are already at the end!
	if b.gapEnd == len(b.buffer) {
		return
	}

	// [ab_____c] becomes [abc_____]
	b.buffer[b.gapStart] = b.buffer[b.gapEnd]
	b.gapStart++
	b.gapEnd++
}

// InsertElements inserts a slice of T at the current cursor position.
func (b *GapBuffer[T]) InsertElements(elements []T) {
	for _, el := range elements {
		b.Insert(el)
	}
}

// Insert inserts a single element at the current cursor position.
func (b *GapBuffer[T]) Insert(el T) {
	if b.gapSize() == 0 {
		b.growGap()
	}

	b.buffer[b.gapStart] = el
	b.gapStart++
}

// Delete deletes the character at the current cursor position.
func (b *GapBuffer[T]) Delete() {
	if b.gapStart == 0 {
		return
	}

	b.gapStart--
}
