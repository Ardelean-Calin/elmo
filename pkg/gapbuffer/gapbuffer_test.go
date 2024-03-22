package gapbuffer

import (
	"testing"

	"github.com/matryer/is"
)

func TestBasic(t *testing.T) {
	is := is.New(t)

	content := []byte("Hello")
	buf := NewGapBuffer[byte]()
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
	b := NewGapBuffer[byte]()
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

	b := NewGapBuffer[rune]()
	b.CursorLeft()
	b.InsertSlice([]rune("Oi!"))

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
	b.InsertSlice([]rune("1) "))

	want = "1) Oi!"
	is.Equal(b.String(), want)
}

func TestRuneAt(t *testing.T) {
	is := is.New(t)

	b := NewGapBuffer[rune]()
	b.SetContent([]rune("Test"))

	is.Equal(b.GetAbs(0), 'T')
	is.Equal(b.GetAbs(1), 'e')
	is.Equal(b.GetAbs(2), 's')
	is.Equal(b.GetAbs(3), 't')
}

func TestDelete(t *testing.T) {
	is := is.New(t)

	b := NewGapBuffer[int]() // {1*, 2, 3, 4}    se
	b.SetContent([]int{1, 2, 3, 4})
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

	want := []int{3, 4}
	got := b.Collect()
	is.Equal(got, want)
}
