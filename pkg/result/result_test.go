package result

import "testing"

func TestOffsetToLineCol(t *testing.T) {
	data := []byte("abc\ndef\nghi")
	cases := []struct {
		offset    int
		line, col int
	}{
		{0, 1, 1},  // start
		{2, 1, 3},  // 'c'
		{4, 2, 1},  // start of line 2 (just after first '\n')
		{6, 2, 3},  // 'f'
		{8, 3, 1},  // start of line 3
		{100, 3, 4}, // clamped past end
	}
	for _, c := range cases {
		line, col := OffsetToLineCol(data, c.offset)
		if line != c.line || col != c.col {
			t.Errorf("offset %d: got (%d,%d), want (%d,%d)", c.offset, line, col, c.line, c.col)
		}
	}
}

func TestOffsetToLineColMultibyte(t *testing.T) {
	// "café" — 'é' is 2 bytes. The 'x' after it is at byte offset 5 but column 5.
	data := []byte("café x")
	line, col := OffsetToLineCol(data, 5) // byte offset of the space
	if line != 1 || col != 5 {
		t.Errorf("got (%d,%d), want (1,5)", line, col)
	}
}
