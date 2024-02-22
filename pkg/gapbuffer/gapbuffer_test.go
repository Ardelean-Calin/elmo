package gapbuffer

import "testing"

func TestBasic(t *testing.T) {
	content := []rune("Hello")
	buf := New(content)
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
	content := []rune("Hello World!")
	b := New(content)
	want := "Hello, World. My name is Calin."

	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	b.CursorRight()
	b.Insert(',')
	b.CursorRight()
	b.CursorRight()
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
	b := New([]rune{})
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
	b := New([]rune("Test"))
	if b.ElementAt(0) != 'T' {
		t.FailNow()
	}
	if b.ElementAt(1) != 'e' {
		t.FailNow()
	}
	if b.ElementAt(2) != 's' {
		t.FailNow()
	}
	if b.ElementAt(3) != 't' {
		t.FailNow()
	}
}
