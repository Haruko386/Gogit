package protocol

import "testing"

func TestScannerFindsSplitMarker(t *testing.T) {
	marker := "__READY__"
	frame := promptFrame(marker, "base", `E:\repo`)
	split := len(frame) / 2
	scanner := NewScanner(marker)

	visible, prompts := scanner.Push(
		[]byte("command output\r\n" + frame[:split]),
	)
	if string(visible) != "command output\r\n" {
		t.Fatalf("first visible output = %q", visible)
	}
	if len(prompts) != 0 {
		t.Fatalf("first prompts = %#v", prompts)
	}

	visible, prompts = scanner.Push([]byte(frame[split:]))
	if len(visible) != 0 {
		t.Fatalf("second visible output = %q", visible)
	}
	if len(prompts) != 1 {
		t.Fatalf("second prompts = %#v", prompts)
	}
	if prompts[0].Environment != "base" {
		t.Fatalf("environment = %q", prompts[0].Environment)
	}
	if prompts[0].Directory != `E:\repo` {
		t.Fatalf("directory = %q", prompts[0].Directory)
	}
}

func TestScannerFindsMultipleMarkers(t *testing.T) {
	marker := "__READY__"
	scanner := NewScanner(marker)
	data := promptFrame(marker, "base", `E:\one`) +
		"one" +
		promptFrame(marker, "dl", `E:\two`) +
		"two"

	visible, prompts := scanner.Push([]byte(data))
	if string(visible) != "onetwo" {
		t.Fatalf("visible output = %q", visible)
	}
	if len(prompts) != 2 {
		t.Fatalf("prompts = %#v", prompts)
	}
	if prompts[1].Environment != "dl" || prompts[1].Directory != `E:\two` {
		t.Fatalf("second prompt = %#v", prompts[1])
	}
}

func TestScannerDoesNotDiscardFalsePrefix(t *testing.T) {
	marker := "__READY__"
	scanner := NewScanner(marker)
	input := "text __READY__PROMPT_BEGX"

	visible, prompts := scanner.Push([]byte(input))
	if string(visible) != input {
		t.Fatalf("visible output = %q", visible)
	}
	if len(prompts) != 0 {
		t.Fatalf("prompts = %#v", prompts)
	}
}

func TestScannerFlushesPendingBytes(t *testing.T) {
	marker := "__READY__"
	scanner := NewScanner(marker)
	begin := BeginMarker(marker)
	partial := begin[:len(begin)-2]

	visible, prompts := scanner.Push([]byte("text " + partial))
	if string(visible) != "text " {
		t.Fatalf("visible output = %q", visible)
	}
	if len(prompts) != 0 {
		t.Fatalf("prompts = %#v", prompts)
	}

	remaining := scanner.Flush()
	if string(remaining) != partial {
		t.Fatalf("remaining output = %q", remaining)
	}
}

func TestNewMarkerCreatesDifferentMarkers(t *testing.T) {
	first, err := NewMarker()
	if err != nil {
		t.Fatal(err)
	}

	second, err := NewMarker()
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Fatalf("markers should be different: %q", first)
	}
}

func promptFrame(marker, environment, directory string) string {
	return BeginMarker(marker) +
		environment +
		string(rune(fieldSeparator)) +
		directory +
		EndMarker(marker)
}
