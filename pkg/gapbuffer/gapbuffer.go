package gapbuffer

import (
	"flag"
)

// Gap Buffer implementation. See: https://routley.io/posts/gap-buffer
type GapBuffer struct {
	buffer   []rune
	gapStart int
	gapEnd   int
}

func New() GapBuffer {
	return GapBuffer{
		buffer:   []rune{},
		gapStart: 0,
		gapEnd:   0,
	}
}

func (b GapBuffer) WithContent(content string) GapBuffer {
	b.buffer = []rune(content)
	// No gap for now
	b.gapStart = 0 // Index of the first character *in* the gap
	b.gapEnd = 0   // Index of the first character *after* the gap

	return b
}

func (b *GapBuffer) String() string {
	return string(b.buffer[:b.gapStart]) + string(b.buffer[b.gapEnd:])
}

func (b *GapBuffer) gapSize() int {
	return b.gapEnd - b.gapStart
}

func (b *GapBuffer) growGap() {
	// Gap size should be up to 5% of the total buffer size, but at least 64 bytes.
	// I've taken this from https://shorturl.at/FKOUZ
	gapSize := max(len(b.buffer)/20, 64)
	// In case of testing we use a smaller size to make it easier.
	if flag.Lookup("test.v") != nil {
		gapSize = 5
	}
	newBuffer := make([]rune, len(b.buffer)+gapSize)

	copy(newBuffer, b.buffer[:b.gapStart])
	copy(newBuffer[b.gapEnd+gapSize:], b.buffer[b.gapEnd:])

	b.buffer = newBuffer
	b.gapEnd = b.gapStart + gapSize
}

// CursorGoto moves the cursor to the given position
func (b *GapBuffer) CursorGoto(pos int) {
	panic("Unimplemented")
}

// CursorRight moves the cursor left one character.
func (b *GapBuffer) CursorLeft() {
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
func (b *GapBuffer) CursorRight() {
	// We are already at the end!
	if b.gapEnd == len(b.buffer) {
		return
	}

	// [ab_____c] becomes [abc_____]
	b.buffer[b.gapStart] = b.buffer[b.gapEnd]
	b.gapStart++
	b.gapEnd++
}

// InsertString inserts a string at the current cursor position.
func (b *GapBuffer) InsertString(s string) {
	runes := []rune(s)
	for _, r := range runes {
		b.Insert(r)
	}
}

// Insert inserts a single character at the current cursor position.
func (b *GapBuffer) Insert(c rune) {
	if b.gapSize() == 0 {
		b.growGap()
	}

	b.buffer[b.gapStart] = c
	b.gapStart++
}

// Delete deletes the character at the current cursor position.
func (b *GapBuffer) Delete() {
	if b.gapStart == 0 {
		return
	}

	b.gapStart--
}
