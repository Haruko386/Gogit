package editor

import "testing"

func TestEditorInsertMoveAndBackspace(t *testing.T) {
	var e Editor
	e.SetLine("git brnch")

	for range 3 {
		if !e.MoveLeft() {
			t.Fatal("MoveLeft unexpectedly reached the beginning")
		}
	}
	e.Insert('a')
	if !e.Backspace() {
		t.Fatal("Backspace did not remove the inserted rune")
	}
	e.Insert('a')

	if got, want := e.Line(), "git branch"; got != want {
		t.Fatalf("Line() = %q, want %q", got, want)
	}
	if got, want := e.Cursor(), 7; got != want {
		t.Fatalf("Cursor() = %d, want %d", got, want)
	}
}

func TestEditorReplaceUsesRuneIndexes(t *testing.T) {
	var e Editor
	e.SetLine("git 分支 --sh")

	if ok := e.Replace(7, 11, "--show-current"); !ok {
		t.Fatal("Replace rejected a valid rune range")
	}
	if got, want := e.Line(), "git 分支 --show-current"; got != want {
		t.Fatalf("Line() = %q, want %q", got, want)
	}
	if got, want := e.Cursor(), 21; got != want {
		t.Fatalf("Cursor() = %d, want %d", got, want)
	}
}

func TestEditorRejectsInvalidReplace(t *testing.T) {
	var e Editor
	e.SetLine("git status")
	before := e.Line()

	if e.Replace(5, 3, "x") {
		t.Fatal("Replace accepted an invalid range")
	}
	if got := e.Line(); got != before {
		t.Fatalf("invalid Replace changed line to %q", got)
	}
}

func TestEditorDelete(t *testing.T) {
	var e Editor
	e.SetLine("git brxanch")

	for range 5 {
		e.MoveLeft()
	}

	if !e.Delete() {
		t.Fatal("Delete did not remove the rune under the cursor")
	}

	if got, want := e.Line(), "git branch"; got != want {
		t.Fatalf("Line() = %q, want %q", got, want)
	}
	if got, want := e.Cursor(), 6; got != want {
		t.Fatalf("Cursor() = %d, want %d", got, want)
	}
}

func TestEditorMoveHomeAndEnd(t *testing.T) {
	var e Editor
	e.SetLine("git status")

	if !e.MoveHome() {
		t.Fatal("MoveHome did not move the cursor")
	}
	if got := e.Cursor(); got != 0 {
		t.Fatalf("cursor after MoveHome = %d, want 0", got)
	}

	if !e.MoveEnd() {
		t.Fatal("MoveEnd did not move the cursor")
	}
	if got, want := e.Cursor(), len([]rune("git status")); got != want {
		t.Fatalf("cursor after MoveEnd = %d, want %d", got, want)
	}
}

func TestEditorDeletePreviousWord(t *testing.T) {
	var e Editor
	e.SetLine("git add 文件 name")

	if !e.DeletePreviousWord() {
		t.Fatal("DeletePreviousWord did not delete the last word")
	}
	if got, want := e.Line(), "git add 文件 "; got != want {
		t.Fatalf("Line() = %q, want %q", got, want)
	}

	if !e.DeletePreviousWord() {
		t.Fatal("DeletePreviousWord did not delete the Unicode word")
	}
	if got, want := e.Line(), "git add "; got != want {
		t.Fatalf("Line() = %q, want %q", got, want)
	}
	if got, want := e.Cursor(), len([]rune("git add ")); got != want {
		t.Fatalf("Cursor() = %d, want %d", got, want)
	}
}

func TestEditorDeleteToStartAndEnd(t *testing.T) {
	var e Editor
	e.SetLine("git status --short")
	for range len([]rune("--short")) {
		e.MoveLeft()
	}

	if !e.DeleteToEnd() {
		t.Fatal("DeleteToEnd did not delete the line suffix")
	}
	if got, want := e.Line(), "git status "; got != want {
		t.Fatalf("Line() = %q, want %q", got, want)
	}

	e.SetLine("git status --short")
	for range len([]rune("--short")) {
		e.MoveLeft()
	}
	if !e.DeleteToStart() {
		t.Fatal("DeleteToStart did not delete the line prefix")
	}
	if got, want := e.Line(), "--short"; got != want {
		t.Fatalf("Line() = %q, want %q", got, want)
	}
	if got := e.Cursor(); got != 0 {
		t.Fatalf("Cursor() = %d, want 0", got)
	}
}

func TestEditorWordAndLineDeletionAtBoundaries(t *testing.T) {
	var e Editor
	if e.DeletePreviousWord() || e.DeleteToStart() || e.DeleteToEnd() {
		t.Fatal("empty editor reported a deletion")
	}
}
