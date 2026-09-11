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

func TestScannerFlushDoesNotExposePartialMarker(t *testing.T) {
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
	if len(remaining) != 0 {
		t.Fatalf("partial protocol marker was exposed: %q", remaining)
	}
}

func TestScannerResynchronizesAtNestedBeginMarker(t *testing.T) {
	marker := "__READY__"
	scanner := NewScanner(marker)
	corrupt := BeginMarker(marker) + "damaged frame"
	valid := promptFrame(marker, "base", "/repo")

	visible, prompts := scanner.Push([]byte(corrupt + valid))
	if len(visible) != 0 {
		t.Fatalf("protocol data became visible: %q", visible)
	}
	if len(prompts) != 1 || prompts[0].Directory != "/repo" {
		t.Fatalf("scanner did not recover: %#v", prompts)
	}
}

func TestScannerResynchronizesAtSplitNestedBeginMarker(t *testing.T) {
	marker := "__READY__"
	scanner := NewScanner(marker)
	begin := BeginMarker(marker)
	split := len(begin) - 3

	visible, prompts := scanner.Push([]byte(
		begin + "damaged frame" + begin[:split],
	))
	if len(visible) != 0 || len(prompts) != 0 {
		t.Fatalf("first Push() = %q, %#v", visible, prompts)
	}

	validTail := begin[split:] + "base" + string(rune(fieldSeparator)) +
		"/repo" + EndMarker(marker)
	visible, prompts = scanner.Push([]byte(validTail))
	if len(visible) != 0 {
		t.Fatalf("protocol data became visible: %q", visible)
	}
	if len(prompts) != 1 || prompts[0].Directory != "/repo" {
		t.Fatalf("scanner did not recover: %#v", prompts)
	}
}

func TestScannerDiscardsOversizedFrameWhenEndMarkerIsPresent(t *testing.T) {
	marker := "__READY__"
	scanner := NewScanner(marker)
	oversized := BeginMarker(marker) +
		string(make([]byte, maxPromptFrameSize+1)) +
		EndMarker(marker)
	valid := promptFrame(marker, "base", "/repo")

	visible, prompts := scanner.Push([]byte(oversized + valid))
	if len(visible) != 0 {
		t.Fatalf("protocol data became visible: %q", visible)
	}
	if len(prompts) != 1 || prompts[0].Directory != "/repo" {
		t.Fatalf("scanner did not recover after oversized frame: %#v", prompts)
	}
}

func TestScannerRescansOutputAfterOversizedUnterminatedFrame(t *testing.T) {
	marker := "__READY__"
	scanner := NewScanner(marker)
	data := BeginMarker(marker) +
		string(make([]byte, maxPromptFrameSize+1)) +
		"visible output"

	visible, prompts := scanner.Push([]byte(data))
	if string(visible) != "visible output" {
		t.Fatalf("visible output = %q", visible)
	}
	if len(prompts) != 0 {
		t.Fatalf("unexpected prompts: %#v", prompts)
	}
}

func TestScannerDiscardsSplitEndMarkerAfterOversizedFrame(t *testing.T) {
	marker := "__READY__"
	scanner := NewScanner(marker)
	end := EndMarker(marker)
	split := len(end) - 3

	visible, prompts := scanner.Push([]byte(
		BeginMarker(marker) +
			string(make([]byte, maxPromptFrameSize+1)) +
			end[:split],
	))
	if len(visible) != 0 || len(prompts) != 0 {
		t.Fatalf("first Push() = %q, %#v", visible, prompts)
	}

	visible, prompts = scanner.Push([]byte(end[split:] + "visible output"))
	if string(visible) != "visible output" {
		t.Fatalf("visible output = %q", visible)
	}
	if len(prompts) != 0 {
		t.Fatalf("unexpected prompts: %#v", prompts)
	}
}

func TestRecoveryNameIsStableAndShellSafe(t *testing.T) {
	first := RecoveryName("marker with unsafe-$-characters")
	second := RecoveryName("marker with unsafe-$-characters")
	if first != second {
		t.Fatalf("recovery name is not stable: %q != %q", first, second)
	}
	for _, value := range first {
		if (value < 'a' || value > 'z') && (value < '0' || value > '9') && value != '_' {
			t.Fatalf("recovery name contains unsafe character %q: %q", value, first)
		}
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
