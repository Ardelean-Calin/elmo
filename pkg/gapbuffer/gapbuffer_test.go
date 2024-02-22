package gapbuffer

import "testing"

func TestBasic(t *testing.T) {
	content := "Hello"
	buf := New().WithContent(content)
	if len(buf.buffer) != 5 {
		t.Fatalf("Gap buffer not initialized correctly.")
	}
	want := "Hello"
	if buf.String() != want {
		t.Fatalf("buf.String() failed. expected: %s, got: %s", want, buf.String())
	}

	buf.InsertString("Boo! ")
	want = "Boo! Hello"
	if buf.String() != want {
		t.Fatalf("InsertString failed. expected: %s, got: %s", want, buf.String())
	}
	if len(buf.buffer) != 10 {
		t.Fatalf("Gap buffer length - expected: %d, got %d", 10, len(buf.buffer))
	}
}

func TestCursor(t *testing.T) {
	content := "Hello World!"
	b := New().WithContent(content)
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
	b.InsertString(" My name is Calin.")

	got := b.String()
	if got != want {
		t.Fatalf("Cursor test failed. expected: %s, got %s", want, got)
	}

}

func TestCursorBounds(t *testing.T) {
	b := New()
	b.CursorLeft()
	b.InsertString("Oi!")

	want := "Oi!"
	got := b.String()
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
	b.InsertString("1) ")

	want = "1) Oi!"
	got = b.String()
	if got != want {
		t.Fatalf("Cursor Bounds Test failed. expected: %s, got %s", want, got)
	}
}

func TestRuneAt(t *testing.T) {
	b := New().WithContent("Test")
	if b.RuneAt(0) != 'T' {
		t.FailNow()
	}
	if b.RuneAt(1) != 'e' {
		t.FailNow()
	}
	if b.RuneAt(2) != 's' {
		t.FailNow()
	}
	if b.RuneAt(3) != 't' {
		t.FailNow()
	}
}
