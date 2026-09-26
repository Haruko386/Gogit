package history

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStoreRoundTripAndLimit(t *testing.T) {
	store := NewStore(
		filepath.Join(t.TempDir(), "history.jsonl"),
		3,
	)

	commands := []string{
		"git status",
		"git add .",
		"git commit -m \"line one\nline two\"",
		"git push",
	}

	for _, command := range commands {
		if err := store.Append(command); err != nil {
			t.Fatal(err)
		}
	}

	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}

	want := commands[1:]
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestStoreSkipsCorruptRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	data := []byte(
		"\"git status\"\n" +
			"not-json\n" +
			"\"git log\"\n",
	)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	store := NewStore(path, 10)
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"git status", "git log"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestShouldPersistFiltersSensitiveCommands(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{command: "git status", want: true},
		{command: " git push", want: false},
		{command: "tool --password hunter2", want: false},
		{command: "export API_TOKEN=value", want: false},
		{command: "curl -H 'Authorization: secret'", want: false},
	}

	for _, test := range tests {
		if got := ShouldPersist(test.command); got != test.want {
			t.Errorf(
				"ShouldPersist(%q) = %t, want %t",
				test.command,
				got,
				test.want,
			)
		}
	}
}

func TestStoreLoadReturnsOpenError(t *testing.T) {
	store := NewStore("invalid\x00path", 10)

	if _, err := store.Load(); err == nil {
		t.Fatal("Load() returned nil error for an invalid path")
	}
}
