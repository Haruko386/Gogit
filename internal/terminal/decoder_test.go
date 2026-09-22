package terminal

import (
	"reflect"
	"testing"
)

func TestDecoderHandlesChunkedArrowSequence(t *testing.T) {
	var decoder Decoder
	if got := decoder.Feed([]byte("\x1b[")); len(got) != 0 {
		t.Fatalf("partial sequence produced keys: %#v", got)
	}
	got := decoder.Feed([]byte("A"))
	want := []Key{{Type: KeyUp}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Feed() = %#v, want %#v", got, want)
	}
}

func TestDecoderHandlesSplitUTF8Rune(t *testing.T) {
	var decoder Decoder
	encoded := []byte("分")
	if got := decoder.Feed(encoded[:1]); len(got) != 0 {
		t.Fatalf("partial rune produced keys: %#v", got)
	}
	got := decoder.Feed(encoded[1:])
	want := []Key{{Type: KeyRune, Rune: '分'}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Feed() = %#v, want %#v", got, want)
	}
}

func TestDecoderTreatsCRLFAsOneEnter(t *testing.T) {
	var decoder Decoder
	got := decoder.Feed([]byte("a\r\nb"))
	want := []Key{
		{Type: KeyRune, Rune: 'a'},
		{Type: KeyEnter},
		{Type: KeyRune, Rune: 'b'},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Feed() = %#v, want %#v", got, want)
	}
}

func TestDecoderRecognizesEditingControlKeys(t *testing.T) {
	var decoder Decoder
	got := decoder.Feed([]byte{'\x01', '\x05', '\x0b', '\x0c', '\x15', '\x17'})
	want := []Key{
		{Type: KeyCtrlA},
		{Type: KeyCtrlE},
		{Type: KeyCtrlK},
		{Type: KeyCtrlL},
		{Type: KeyCtrlU},
		{Type: KeyCtrlW},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Feed() = %#v, want %#v", got, want)
	}
}

func TestDecoderFlushesStandaloneEscape(t *testing.T) {
	var decoder Decoder
	if got := decoder.Feed([]byte("\x1b")); len(got) != 0 {
		t.Fatalf("pending Escape produced keys: %#v", got)
	}
	got := decoder.Flush()
	want := []Key{{Type: KeyEscape}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Flush() = %#v, want %#v", got, want)
	}
}
