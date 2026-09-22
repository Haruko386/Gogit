package terminal

import (
	"bytes"
	"unicode/utf8"
)

var keySequences = []struct {
	sequence []byte
	key      KeyType
}{
	{sequence: []byte("\x1b[A"), key: KeyUp},
	{sequence: []byte("\x1b[B"), key: KeyDown},
	{sequence: []byte("\x1b[C"), key: KeyRight},
	{sequence: []byte("\x1b[D"), key: KeyLeft},
	{sequence: []byte("\x1b[H"), key: KeyHome},
	{sequence: []byte("\x1b[F"), key: KeyEnd},
	{sequence: []byte("\x1b[3~"), key: KeyDelete},
}

// Decoder turns arbitrarily chunked terminal bytes into logical key events.
type Decoder struct {
	pending []byte
	skipLF  bool
}

// Feed accepts the next bytes read from the terminal.
func (d *Decoder) Feed(data []byte) []Key {
	d.pending = append(d.pending, data...)
	keys := make([]Key, 0, len(data))

	for len(d.pending) > 0 {
		if d.pending[0] == '\x1b' {
			key, size, complete := decodeEscape(d.pending)
			if !complete {
				break
			}
			d.pending = d.pending[size:]
			keys = append(keys, Key{Type: key})
			d.skipLF = false
			continue
		}

		value := d.pending[0]
		switch value {
		case '\r':
			keys = append(keys, Key{Type: KeyEnter})
			d.pending = d.pending[1:]
			d.skipLF = true
			continue
		case '\n':
			d.pending = d.pending[1:]
			if d.skipLF {
				d.skipLF = false
				continue
			}
			keys = append(keys, Key{Type: KeyEnter})
			continue
		case '\x08', '\x7f':
			keys = append(keys, Key{Type: KeyBackspace})
		case '\t':
			keys = append(keys, Key{Type: KeyTab})
		case '\x01':
			keys = append(keys, Key{Type: KeyCtrlA})
		case '\x03':
			keys = append(keys, Key{Type: KeyCtrlC})
		case '\x04':
			keys = append(keys, Key{Type: KeyCtrlD})
		case '\x05':
			keys = append(keys, Key{Type: KeyCtrlE})
		case '\x0b':
			keys = append(keys, Key{Type: KeyCtrlK})
		case '\x0c':
			keys = append(keys, Key{Type: KeyCtrlL})
		case '\x15':
			keys = append(keys, Key{Type: KeyCtrlU})
		case '\x17':
			keys = append(keys, Key{Type: KeyCtrlW})
		default:
			if !utf8.FullRune(d.pending) {
				return keys
			}
			decoded, size := utf8.DecodeRune(d.pending)
			keys = append(keys, Key{Type: KeyRune, Rune: decoded})
			d.pending = d.pending[size:]
			d.skipLF = false
			continue
		}

		d.pending = d.pending[1:]
		d.skipLF = false
	}

	return keys
}

// Flush emits a pending standalone Escape key. It is intended for an input
// loop that uses a short timeout to distinguish Escape from an ANSI sequence.
func (d *Decoder) Flush() []Key {
	if len(d.pending) == 0 {
		return nil
	}

	keys := make([]Key, 0, len(d.pending))
	if d.pending[0] == '\x1b' {
		keys = append(keys, Key{Type: KeyEscape})
		d.pending = d.pending[1:]
	}
	remaining := append([]byte(nil), d.pending...)
	d.pending = nil
	return append(keys, d.Feed(remaining)...)
}

func decodeEscape(data []byte) (KeyType, int, bool) {
	for _, candidate := range keySequences {
		if len(data) >= len(candidate.sequence) && bytes.Equal(data[:len(candidate.sequence)], candidate.sequence) {
			return candidate.key, len(candidate.sequence), true
		}
		if bytes.HasPrefix(candidate.sequence, data) {
			return 0, 0, false
		}
	}

	if len(data) == 1 {
		return 0, 0, false
	}
	if data[1] != '[' {
		return KeyEscape, 1, true
	}

	for index := 2; index < len(data); index++ {
		if data[index] >= 0x40 && data[index] <= 0x7e {
			return KeyEscape, index + 1, true
		}
	}
	return 0, 0, false
}
