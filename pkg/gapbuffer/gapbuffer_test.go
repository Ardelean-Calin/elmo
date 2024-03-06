package gapbuffer

import (
	"slices"
	"testing"
)

func TestBasic(t *testing.T) {
	content := []rune("Hello")
	buf := NewGapBuffer[rune]()
	buf.SetContent(content)
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

	got := string(b.Collect())
	if got != want {
		t.Fatalf("Cursor test failed. expected: %s, got %s", want, got)
	}

}

func TestCursorBounds(t *testing.T) {
	b := NewGapBuffer[rune]()
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
	b := NewGapBuffer[rune]()
	if b.GetAbs(0) != 'T' {
		t.FailNow()
	}
	if b.GetAbs(1) != 'e' {
		t.FailNow()
	}
	if b.GetAbs(2) != 's' {
		t.FailNow()
	}
	if b.GetAbs(3) != 't' {
		t.FailNow()
	}
}

func TestDelete(t *testing.T) { //  se
	b := NewGapBuffer[int]() // {1*, 2, 3, 4}    se
	b.CursorRight()          // now at 2  => {1, 2*, 3, 4}
	if b.GapStart != 1 || b.GapEnd != 1 {
		t.Fatalf("Unexpected Gaps - GapStart: %d\tGapEnd: %d", b.GapStart, b.GapEnd)
	}
	//                             s  e
	b.Delete() // deleted 2 => {1, _, 3*, 4}
	if b.GapStart != 1 || b.GapEnd != 2 {
		t.Fatalf("Unexpected Gaps - GapStart: %d\tGapEnd: %d", b.GapStart, b.GapEnd)
	}
	//                             s     e
	b.Backspace() // deleted 1 => {_, _, 3*, 4}
	if b.GapStart != 0 || b.GapEnd != 2 {
		t.Fatalf("Unexpected Gaps - GapStart: %d\tGapEnd: %d", b.GapStart, b.GapEnd)
	}

	want := []int{3, 4}
	got := b.Collect()

	if !slices.Equal(got, want) {
		t.Fatalf("Backspace delete failed. Expected: %v - Got: %v", want, got)
	}

}
