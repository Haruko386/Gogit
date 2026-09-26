package history

import "testing"

func TestHistoryPreviousAndNext(t *testing.T) {
	var history History
	history.Add("git status")
	history.Add("go test ./...")

	command, ok := history.Previous("")
	if !ok || command != "go test ./..." {
		t.Fatalf("first Previous() = %q, %v", command, ok)
	}

	command, ok = history.Previous(command)
	if !ok || command != "git status" {
		t.Fatalf("second Previous() = %q, %v", command, ok)
	}

	command, ok = history.Next()
	if !ok || command != "go test ./..." {
		t.Fatalf("first Next() = %q, %v", command, ok)
	}

	command, ok = history.Next()
	if !ok || command != "" {
		t.Fatalf("second Next() = %q, %v", command, ok)
	}
}

func TestHistoryRestoresDraft(t *testing.T) {
	var history History
	history.Add("git status")

	command, ok := history.Previous("git br")
	if !ok || command != "git status" {
		t.Fatalf("Previous() = %q, %v", command, ok)
	}

	command, ok = history.Next()
	if !ok || command != "git br" {
		t.Fatalf("Next() = %q, %v", command, ok)
	}
}

func TestHistorySkipsConsecutiveDuplicates(t *testing.T) {
	var history History
	history.Add("git status")
	history.Add("git status")

	command, ok := history.Previous("")
	if !ok || command != "git status" {
		t.Fatalf("Previous() = %q, %v", command, ok)
	}

	if _, ok := history.Previous(command); ok {
		t.Fatal("duplicate command was stored")
	}
}

func TestHistoryLimitsEntries(t *testing.T) {
	history := New(nil, 2)

	history.Add("git status")
	history.Add("git add .")
	history.Add("git commit")

	command, ok := history.Previous("")
	if !ok || command != "git commit" {
		t.Fatalf("latest command = %q, %t", command, ok)
	}

	command, ok = history.Previous(command)
	if !ok || command != "git add ." {
		t.Fatalf("oldest retained command = %q, %t", command, ok)
	}

	if _, ok := history.Previous(command); ok {
		t.Fatal("history retained more entries than its limit")
	}
}
