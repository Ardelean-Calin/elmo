package gapbuffer

import (
	"testing"

	"github.com/matryer/is"
)

func TestBasic(t *testing.T) {
	is := is.New(t)

	content := []byte("Hello")
	buf := NewGapBuffer()
	buf.SetContent(content)

	is.Equal(buf.String(), "Hello")
	is.Equal(buf.TotalLen(), 5)

	buf.InsertSlice([]byte("Boo! "))
	is.Equal(buf.String(), "Boo! Hello")
	is.Equal(buf.TotalLen(), 10)
}

func TestCursor(t *testing.T) {
	is := is.New(t)

	content := []byte("HelloWorld!")
	b := NewGapBuffer()
	b.SetContent(content)
	want := "Hello, World. My name is Calin."

	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	// b.InsertElements([]rune(", "))
	b.Insert(',')
	b.Insert(' ')
	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	b.Delete()
	b.Insert('.')
	b.InsertSlice([]byte(" My name is Calin."))

	is.Equal(b.String(), want)
}

func TestCursorBounds(t *testing.T) {
	is := is.New(t)

	b := NewGapBuffer()
	b.CursorLeft()
	b.InsertSlice([]byte("Oi!"))

	want := "Oi!"
	is.Equal(b.String(), want)

	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.InsertSlice([]byte("1) "))

	want = "1) Oi!"
	is.Equal(b.String(), want)
}

func TestByteAt(t *testing.T) {
	is := is.New(t)

	b := NewGapBuffer()
	b.SetContent([]byte("Test"))

	is.Equal(b.GetAbs(0), byte('T'))
	is.Equal(b.GetAbs(1), byte('e'))
	is.Equal(b.GetAbs(2), byte('s'))
	is.Equal(b.GetAbs(3), byte('t'))
}

func TestDelete(t *testing.T) {
	is := is.New(t)

	b := NewGapBuffer() // {1*, 2, 3, 4}    se
	b.SetContent([]byte{1, 2, 3, 4})
	is.Equal(b.GapStart, 0)
	is.Equal(b.GapEnd, 0)

	b.CursorRight() // now at 2  => {1, 2*, 3, 4}
	is.Equal(b.GapStart, 1)
	is.Equal(b.GapEnd, 1)
	//                             s  e
	b.Delete() // deleted 2 => {1, _, 3*, 4}
	is.Equal(b.GapStart, 1)
	is.Equal(b.GapEnd, 2)
	//                             s     e
	b.Backspace() // deleted 1 => {_, _, 3*, 4}
	is.Equal(b.GapStart, 0)
	is.Equal(b.GapEnd, 2)

	want := []byte{3, 4}
	got := b.Bytes()
	is.Equal(got, want)
}
