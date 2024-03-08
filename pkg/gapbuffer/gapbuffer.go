package gapbuffer

import (
	"flag"
	"strings"
)

// Comparable is a custom interface that both int and rune satisfy.
type Comparable interface {
	~int | ~rune
}

// Gap Buffer implementation. See: https://routley.io/posts/gap-buffer
type GapBuffer[T Comparable] struct {
	Buffer   []T // NOTE: In the future I might want this to be a pointer. I avoided this for now as it looked ugly
	GapStart int // Index of the first character *in* the gap
	GapEnd   int // Index of the first character *after* the gap
}

// NewGapBuffer creates a new empty gapbuffer
func NewGapBuffer[T Comparable]() GapBuffer[T] {
	return GapBuffer[T]{
		Buffer:   []T{},
		GapStart: 0,
		GapEnd:   0,
	}
}

// SetContent sets the gapbuffer content
func (b *GapBuffer[T]) SetContent(content []T) {
	b.Buffer = content
}

// Collect returns a slice of the content of the gap buffer, without gap
func (b *GapBuffer[T]) Collect() []T {
	var dest []T
	dest = append(dest, b.Buffer[:b.GapStart]...)
	dest = append(dest, b.Buffer[b.GapEnd:]...)
	return dest
}

// Pos returns the current position inside the Gap Buffer
func (b *GapBuffer[T]) Pos() int {
	return b.GapStart
}

// FindAll returns a slice with the indices of all found items inside the gap buffer
func (b *GapBuffer[T]) FindAll(val T) []int {
	var results []int
	for index, v := range b.Buffer {
		if v == val {
			results = append(results, index)
		}
	}

	return results
}

// gapSize returns the gap size
func (b *GapBuffer[T]) gapSize() int {
	return b.GapEnd - b.GapStart
}

// growGap grows the gap by up to 5% of the total buffer size
func (b *GapBuffer[T]) growGap() {
	// Gap size should be up to 5% of the total buffer size, but at least 64 bytes.
	// I've taken this from https://shorturl.at/FKOUZ
	gapSize := max(len(b.Buffer)/20, 64)
	// In case of testing we use a smaller size to make it easier.
	if flag.Lookup("test.v") != nil {
		gapSize = 5
	}
	newBuffer := make([]T, len(b.Buffer)+gapSize)

	copy(newBuffer, b.Buffer[:b.GapStart])
	copy(newBuffer[b.GapEnd+gapSize:], b.Buffer[b.GapEnd:])

	b.Buffer = newBuffer
	b.GapEnd = b.GapStart + gapSize
}

// Count returns the size of the raw text, excliding the gap
func (b *GapBuffer[T]) Count() int {
	return len(b.Buffer) - b.gapSize()
}

// Len returns the total length of the gap buffer, including gap
func (b *GapBuffer[T]) Len() int {
	return len(b.Buffer)
}

// TotalLen returns the total length of the gap buffer (including gaps)
func (b *GapBuffer[T]) TotalLen() int {
	return len(b.Buffer)
}

// GetAbs returns the element at the given position. Ignores gap and treats buffer
// as a linear array
func (b *GapBuffer[T]) GetAbs(pos int) T {
	if pos > b.GapStart {
		pos += b.gapSize()
	}
	pos = min(b.Len()-1, pos)

	return b.Buffer[pos]
}

// CursorGoto moves the cursor to the given (absolute) position
func (b *GapBuffer[T]) CursorGoto(pos int) (actualPos int) {
	if pos+b.gapSize() > b.Len() {
		pos = b.Len() - b.gapSize()
	}

	if pos > b.GapStart {
		// [a b _ _ _ c d e f] becomes [a b c d e _ _ _ f]
		diff := pos - b.GapStart
		copy(b.Buffer[b.GapStart:], b.Buffer[b.GapEnd:b.GapEnd+diff])
	} else if pos < b.GapStart {
		// [a b c d e _ _ _ f] becomes [a b _ _ _ c d e f]
		//      ^     s     e               s     e
		//                                  ^
		// Step 0: New empty buffer [_ _ _ _ _ _ _ _ _]
		newBuf := make([]T, b.TotalLen())
		i := 0

		// Step 1: Copy part until new position to buffer
		// [a b _ _ _ _ _ _ _]
		i += copy(newBuf[i:pos], b.Buffer[i:pos])
		// Step 2: Leave a gap
		i += b.gapSize()
		// Step 3: Copy part from position to gapStart
		// [a b _ _ _ c d e _]
		i += copy(newBuf[i:], b.Buffer[pos:b.GapStart])
		// Finally: Copy all the bytes from gapEnd to end
		copy(newBuf[i:], b.Buffer[b.GapEnd:])

		b.Buffer = newBuf
	}
	b.GapEnd = pos + b.gapSize()
	b.GapStart = pos
	actualPos = pos
	return
}

// CursorRight moves the cursor left one character.
// NOTE: The cursor is always the start of the gap
func (b *GapBuffer[T]) CursorLeft() {
	// We are already at the start!
	if b.GapStart == 0 {
		return
	}

	// The first element before the start of the gap gets copied after the gap
	// [abc_____] becomes [ab_____c]
	b.Buffer[b.GapEnd-1] = b.Buffer[b.GapStart-1]
	b.GapStart--
	b.GapEnd--
}

// CursorRight moves the cursor right one character.
// NOTE: The cursor is always the start of the gap
func (b *GapBuffer[T]) CursorRight() {
	// We are already at the end!
	if b.GapEnd == len(b.Buffer) {
		return
	}

	// [ab_____c] becomes [abc_____]
	b.Buffer[b.GapStart] = b.Buffer[b.GapEnd]
	b.GapStart++
	b.GapEnd++
}

// InsertSlice inserts a slice of T at the current cursor position.
func (b *GapBuffer[T]) InsertSlice(slice []T) {
	for _, el := range slice {
		b.Insert(el)
	}
}

// Insert inserts a single element at the current cursor position.
func (b *GapBuffer[T]) Insert(el T) {
	if b.gapSize() == 0 {
		b.growGap()
	}

	b.Buffer[b.GapStart] = el
	b.GapStart++
}

// Delete deletes the character at the current cursor position.
func (b *GapBuffer[T]) Delete() {
	if b.GapEnd == len(b.Buffer) {
		return
	}

	// Grow the gap towards right
	b.GapEnd++
}

// Backspace deletes the character before the current position
func (b *GapBuffer[T]) Backspace() {
	if b.GapStart == 0 {
		return
	}

	// Grow the gap towards left
	b.GapStart--
}

// String returns a string repesentation of this Gap Buffer.
// NOTE: Will return gibberish for non-rune type Gap Buffers.
func (b *GapBuffer[T]) String() string {
	var sb strings.Builder
	for _, x := range b.Collect() {
		sb.WriteRune(rune(x))
	}
	return sb.String()
}

/* Provide an iterator interface for the GapBuffer.
   This iterator jumps over gaps.
*/

func (b *GapBuffer[T]) Iter() GapBufferIterator[T] {
	return GapBufferIterator[T]{
		gb:    b,
		index: 0,
	}
}

type GapBufferIterator[T Comparable] struct {
	index int
	gb    *GapBuffer[T]
}

func (gi *GapBufferIterator[T]) HasNext() bool {
	return gi.index < gi.gb.Count()
}

func (gi *GapBufferIterator[T]) Next() (int, T) {
	index := gi.index
	val := gi.gb.GetAbs(index)
	gi.index++
	return index, val
}
