package gapbuffer

import (
	"slices"
	"testing"
)

func TestBasic(t *testing.T) {
	content := []rune("Hello")
	buf := NewGapBuffer(content)
	if buf.TotalLen() != 5 {
		t.Fatalf("Gap buffer not initialized correctly.")
	}
	want := "Hello"
	if string(buf.Collect()) != want {
		t.Fatalf("buf.String() failed. expected: %s, got: %s", want, string(buf.Collect()))
	}

	buf.InsertElements([]rune("Boo! "))
	want = "Boo! Hello"
	if string(buf.Collect()) != want {
		t.Fatalf("InsertElements failed. expected: %s, got: %s", want, string(buf.Collect()))
	}
	if buf.TotalLen() != 10 {
		t.Fatalf("Gap buffer length - expected: %d, got %d", 10, buf.TotalLen())
	}
}

func TestCursor(t *testing.T) {
	content := []rune("HelloWorld!")
	b := NewGapBuffer(content)
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

	got := string(b.Collect())
	if got != want {
		t.Fatalf("Cursor test failed. expected: %s, got %s", want, got)
	}

}

func TestCursorBounds(t *testing.T) {
	b := NewGapBuffer([]rune{})
	b.CursorLeft()
	b.InsertElements([]rune("Oi!"))

	want := "Oi!"
	got := string(b.Collect())
	if got != want {
		t.Fatalf("Cursor Bounds Test failed. expected: %s, got %s", want, got)
	}

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
	got = string(b.Collect())
	if got != want {
		t.Fatalf("Cursor Bounds Test failed. expected: %s, got %s", want, got)
	}
}

func TestRuneAt(t *testing.T) {
	b := NewGapBuffer([]rune("Test"))
	if b.Get(0) != 'T' {
		t.FailNow()
	}
	if b.Get(1) != 'e' {
		t.FailNow()
	}
	if b.Get(2) != 's' {
		t.FailNow()
	}
	if b.Get(3) != 't' {
		t.FailNow()
	}
}

func TestDelete(t *testing.T) { //  se
	b := NewGapBuffer([]int{1, 2, 3, 4}) // {1*, 2, 3, 4}    se
	b.CursorRight()                      // now at 2  => {1, 2*, 3, 4}
	if b.gapStart != 1 || b.gapEnd != 1 {
		t.Fatalf("Unexpected gaps - gapStart: %d\tgapEnd: %d", b.gapStart, b.gapEnd)
	}
	//                             s  e
	b.Delete() // deleted 2 => {1, _, 3*, 4}
	if b.gapStart != 1 || b.gapEnd != 2 {
		t.Fatalf("Unexpected gaps - gapStart: %d\tgapEnd: %d", b.gapStart, b.gapEnd)
	}
	//                             s     e
	b.Backspace() // deleted 1 => {_, _, 3*, 4}
	if b.gapStart != 0 || b.gapEnd != 2 {
		t.Fatalf("Unexpected gaps - gapStart: %d\tgapEnd: %d", b.gapStart, b.gapEnd)
	}

	want := []int{3, 4}
	got := b.Collect()

	if !slices.Equal(got, want) {
		t.Fatalf("Backspace delete failed. Expected: %v - Got: %v", want, got)
	}

}
