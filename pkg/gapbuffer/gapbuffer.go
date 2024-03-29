package gapbuffer

import (
	"bytes"
	"flag"
	"io"
	"strings"
)

// Gap Buffer implementation. See: https://routley.io/posts/gap-buffer
type GapBuffer struct {
	Buffer   []byte // NOTE: In the future I might want this to be a pointer. I avoided this for now as it looked ugly
	GapStart int    // Index of the first character *in* the gap
	GapEnd   int    // Index of the first character *after* the gap
}

// NewGapBuffer creates a new empty gapbuffer
func NewGapBuffer() GapBuffer {
	return GapBuffer{
		Buffer:   []byte{},
		GapStart: 0,
		GapEnd:   0,
	}
}

// SetContent sets the gapbuffer content
func (gb *GapBuffer) SetContent(content []byte) {
	gb.Buffer = content
}

func (gb *GapBuffer) Reader() io.Reader {
	return bytes.NewBuffer(gb.Bytes())
}

// Bytes returns a slice of the content of the gap buffer, without gap
func (gb *GapBuffer) Bytes() []byte {
	var dest []byte
	dest = append(dest, gb.Buffer[:gb.GapStart]...)
	dest = append(dest, gb.Buffer[gb.GapEnd:]...)
	return dest
}

func (gb *GapBuffer) Split(sep byte) [][]byte {
	var splits [][]byte = make([][]byte, len(gb.FindAll(sep)))
	index := 0
	for _, r := range gb.Bytes() {
		if r == sep {
			index++
		} else {
			splits[index] = append(splits[index], r)
		}
	}

	return splits
}

// Pos returns the current position inside the Gap Buffer
func (gb *GapBuffer) Pos() int {
	return gb.GapStart
}

// Find finds the first occurence of the given value inside the gapbuffer
// and returns its index.
// Note: Find returns the absolute index, ignoring gap
func (gb *GapBuffer) Find(val byte) (i int, ok bool) {
	for i, v := range gb.Bytes() {
		if v == val {
			ok = true
			return i, ok
		}
	}

	return -1, false
}

// FindAll returns a slice with the indices of all found items inside the gap buffer
// Note: FindAll returns the absolute index, ignoring gap
func (gb *GapBuffer) FindAll(val byte) []int {
	var results []int
	for index, v := range gb.Bytes() {
		if v == val {
			results = append(results, index)
		}
	}

	return results
}

// gapSize returns the gap size
func (gb *GapBuffer) gapSize() int {
	return gb.GapEnd - gb.GapStart
}

// growGap grows the gap by up to 5% of the total buffer size
func (gb *GapBuffer) growGap() {
	// Gap size should be up to 5% of the total buffer size, but at least 64 bytes.
	// I've taken this from https://shorturl.at/FKOUZ
	gapSize := max(len(gb.Buffer)/20, 64)
	// In case of testing we use a smaller size to make it easier.
	if flag.Lookup("test.v") != nil {
		gapSize = 5
	}
	newBuffer := make([]byte, len(gb.Buffer)+gapSize)

	copy(newBuffer, gb.Buffer[:gb.GapStart])
	copy(newBuffer[gb.GapEnd+gapSize:], gb.Buffer[gb.GapEnd:])

	gb.Buffer = newBuffer
	gb.GapEnd = gb.GapStart + gapSize
}

// Len returns the total length of the gap buffer, excluding gap
func (gb *GapBuffer) Len() int {
	return len(gb.Buffer) - gb.gapSize()
}

// TotalLen returns the total length of the gap buffer (including gaps)
func (gb *GapBuffer) TotalLen() int {
	return len(gb.Buffer)
}

// Cursor returns the current cursor position
func (gb *GapBuffer) Cursor() int {
	return gb.GapStart
}

// Current returns the element at the current cursor position
func (gb *GapBuffer) Current() byte {
	return gb.Buffer[gb.GapStart]
}

// Next returns the next element after the cursor
func (gb *GapBuffer) Next() byte {
	if gb.GapEnd+1 >= gb.Len() {
		return gb.Current()
	}

	return gb.Buffer[gb.GapEnd+1]
}

// GetAbs returns the element at the given position. Ignores gap and treats buffer
// as a linear array
func (gb *GapBuffer) GetAbs(pos int) byte {
	if pos > gb.GapStart {
		pos += gb.gapSize()
	}

	pos = clamp(pos, 0, gb.Len())

	return gb.Buffer[pos]
}

// CursorGoto moves the cursor to the given (absolute) position
func (gb *GapBuffer) CursorGoto(pos int) (actualPos int) {
	if pos+gb.gapSize() > gb.Len() {
		pos = gb.Len() - gb.gapSize()
	}

	if pos > gb.GapStart {
		// [a b _ _ _ c d e f] becomes [a b c d e _ _ _ f]
		diff := pos - gb.GapStart
		copy(gb.Buffer[gb.GapStart:], gb.Buffer[gb.GapEnd:gb.GapEnd+diff])
	} else if pos < gb.GapStart {
		// [a b c d e _ _ _ f] becomes [a b _ _ _ c d e f]
		//      ^     s     e               s     e
		//                                  ^
		// Step 0: New empty buffer [_ _ _ _ _ _ _ _ _]
		newBuf := make([]byte, gb.TotalLen())
		i := 0

		// Step 1: Copy part until new position to buffer
		// [a b _ _ _ _ _ _ _]
		i += copy(newBuf[i:pos], gb.Buffer[i:pos])
		// Step 2: Leave a gap
		i += gb.gapSize()
		// Step 3: Copy part from position to gapStart
		// [a b _ _ _ c d e _]
		i += copy(newBuf[i:], gb.Buffer[pos:gb.GapStart])
		// Finally: Copy all the bytes from gapEnd to end
		copy(newBuf[i:], gb.Buffer[gb.GapEnd:])

		gb.Buffer = newBuf
	}
	gb.GapEnd = pos + gb.gapSize()
	gb.GapStart = pos
	actualPos = pos
	return
}

// CursorRight moves the cursor left one character.
// NOTE: The cursor is always the start of the gap
func (gb *GapBuffer) CursorLeft() {
	// We are already at the start!
	if gb.GapStart == 0 {
		return
	}

	// The first element before the start of the gap gets copied after the gap
	// [abc_____] becomes [ab_____c]
	gb.Buffer[gb.GapEnd-1] = gb.Buffer[gb.GapStart-1]
	gb.GapStart--
	gb.GapEnd--
}

// CursorRight moves the cursor right one character.
// NOTE: The cursor is always the start of the gap
func (gb *GapBuffer) CursorRight() {
	// We are already at the end!
	if gb.GapEnd == len(gb.Buffer)-1 {
		return
	}

	// [ab_____c] becomes [abc_____]
	gb.Buffer[gb.GapStart] = gb.Buffer[gb.GapEnd]
	gb.GapStart++
	gb.GapEnd++
}

// InsertSlice inserts a slice of T at the current cursor position.
func (gb *GapBuffer) InsertSlice(slice []byte) {
	for _, el := range slice {
		gb.Insert(el)
	}
}

// Insert inserts a single element at the current cursor position.
func (gb *GapBuffer) Insert(el byte) {
	if gb.gapSize() == 0 {
		gb.growGap()
	}

	gb.Buffer[gb.GapStart] = el
	gb.GapStart++
}

// Delete deletes the character at the current cursor position.
func (gb *GapBuffer) Delete() {
	if gb.GapEnd == len(gb.Buffer) {
		return
	}

	// Grow the gap towards right
	gb.GapEnd++
}

func (gb *GapBuffer) DeleteRange(length int) {
	if gb.GapEnd == len(gb.Buffer) {
		return
	}

	gb.GapEnd = clamp(gb.GapEnd+length, 0, len(gb.Buffer))
}

// Backspace deletes the character before the current position
func (gb *GapBuffer) Backspace() {
	if gb.GapStart == 0 {
		return
	}

	// Grow the gap towards left
	gb.GapStart--
}

// String returns a string repesentation of this Gap Buffer.
// NOTE: Will return gibberish for non-rune type Gap Buffers.
func (gb *GapBuffer) String() string {
	var sb strings.Builder
	for _, x := range gb.Bytes() {
		sb.WriteRune(rune(x))
	}
	return sb.String()
}

/* Provide an iterator interface for the GapBuffer.
   This iterator jumps over gaps.
*/

func (gb *GapBuffer) Iter() GapBufferIterator {
	return GapBufferIterator{
		gb:    gb,
		index: 0,
	}
}

type GapBufferIterator struct {
	index int
	gb    *GapBuffer
}

func (gi *GapBufferIterator) HasNext() bool {
	return gi.index < gi.gb.Len()
}

func (gi *GapBufferIterator) Next() (int, byte) {
	index := gi.index
	val := gi.gb.GetAbs(index)
	gi.index++
	return index, val
}

// clamp limits the value of val between [low, high)
func clamp(val, low, high int) int {
	return max(low, min(val, high-1))
}
