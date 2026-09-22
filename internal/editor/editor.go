package editor

import "unicode"

// Editor owns the text and cursor state for one command line. Cursor offsets
// are rune indexes rather than byte indexes so non-ASCII input is not split.
type Editor struct {
	buffer []rune
	cursor int
}

// Line returns the complete command line.
func (e *Editor) Line() string {
	return string(e.buffer)
}

// Cursor returns the cursor position as a rune index.
func (e *Editor) Cursor() int {
	return e.cursor
}

// Insert adds a rune at the cursor and advances the cursor.
func (e *Editor) Insert(value rune) {
	e.buffer = append(e.buffer, 0)
	copy(e.buffer[e.cursor+1:], e.buffer[e.cursor:])
	e.buffer[e.cursor] = value
	e.cursor++
}

// Backspace removes the rune immediately before the cursor.
func (e *Editor) Backspace() bool {
	if e.cursor == 0 {
		return false
	}

	copy(e.buffer[e.cursor-1:], e.buffer[e.cursor:])
	e.buffer = e.buffer[:len(e.buffer)-1]
	e.cursor--
	return true
}

// MoveLeft moves the cursor one rune to the left.
func (e *Editor) MoveLeft() bool {
	if e.cursor == 0 {
		return false
	}
	e.cursor--
	return true
}

// MoveRight moves the cursor one rune to the right.
func (e *Editor) MoveRight() bool {
	if e.cursor == len(e.buffer) {
		return false
	}
	e.cursor++
	return true
}

// MoveHome moves the cursor to the beginning of the line.
func (e *Editor) MoveHome() bool {
	if e.cursor == 0 {
		return false
	}

	e.cursor = 0
	return true
}

// MoveEnd moves the cursor to the end of the line.
func (e *Editor) MoveEnd() bool {
	if e.cursor == len(e.buffer) {
		return false
	}

	e.cursor = len(e.buffer)
	return true
}

// Delete removes the rune under the cursor.
func (e *Editor) Delete() bool {
	if e.cursor >= len(e.buffer) {
		return false
	}

	e.buffer = append(e.buffer[:e.cursor], e.buffer[e.cursor+1:]...)
	return true
}

// DeletePreviousWord removes whitespace and the word immediately before the
// cursor, matching the common Ctrl+W shell editing behavior.
func (e *Editor) DeletePreviousWord() bool {
	if e.cursor == 0 {
		return false
	}

	start := e.cursor
	for start > 0 && unicode.IsSpace(e.buffer[start-1]) {
		start--
	}
	for start > 0 && !unicode.IsSpace(e.buffer[start-1]) {
		start--
	}

	e.buffer = append(e.buffer[:start], e.buffer[e.cursor:]...)
	e.cursor = start
	return true
}

// DeleteToStart removes everything before the cursor.
func (e *Editor) DeleteToStart() bool {
	if e.cursor == 0 {
		return false
	}

	e.buffer = append(e.buffer[:0], e.buffer[e.cursor:]...)
	e.cursor = 0
	return true
}

// DeleteToEnd removes everything from the cursor onward.
func (e *Editor) DeleteToEnd() bool {
	if e.cursor == len(e.buffer) {
		return false
	}

	e.buffer = e.buffer[:e.cursor]
	return true
}

// Replace replaces the rune range [start, end) and places the cursor after the
// replacement. It returns false when the range is invalid.
func (e *Editor) Replace(start, end int, value string) bool {
	if start < 0 || end < start || end > len(e.buffer) {
		return false
	}

	replacement := []rune(value)
	next := make([]rune, 0, len(e.buffer)-(end-start)+len(replacement))
	next = append(next, e.buffer[:start]...)
	next = append(next, replacement...)
	next = append(next, e.buffer[end:]...)

	e.buffer = next
	e.cursor = start + len(replacement)
	return true
}

// SetLine replaces the current command and moves the cursor to its end.
func (e *Editor) SetLine(line string) {
	e.buffer = []rune(line)
	e.cursor = len(e.buffer)
}

// Clear removes the current command.
func (e *Editor) Clear() {
	e.buffer = nil
	e.cursor = 0
}
