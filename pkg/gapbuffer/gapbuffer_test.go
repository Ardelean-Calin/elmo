package gapbuffer

import (
	"testing"

	"github.com/matryer/is"
)

func TestBasic(t *testing.T) {
	is := is.New(t)

	content := []rune("Hello")
	buf := NewGapBuffer[rune]()
	buf.SetContent(content)

	is.Equal(string(buf.Collect()), "Hello")
	is.Equal(buf.TotalLen(), 5)

	buf.InsertElements([]rune("Boo! "))
	is.Equal(string(buf.Collect()), "Boo! Hello")
	is.Equal(buf.TotalLen(), 10)
}

func TestCursor(t *testing.T) {
	is := is.New(t)

	content := []rune("HelloWorld!")
	b := NewGapBuffer[rune]()
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
	b.InsertElements([]rune(" My name is Calin."))

	is.Equal(string(b.Collect()), want)
}

func TestCursorBounds(t *testing.T) {
	is := is.New(t)

	b := NewGapBuffer[rune]()
	b.CursorLeft()
	b.InsertElements([]rune("Oi!"))

	want := "Oi!"
	is.Equal(string(b.Collect()), want)

	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.CursorLeft()
	b.InsertElements([]rune("1) "))

	want = "1) Oi!"
	is.Equal(string(b.Collect()), want)
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
