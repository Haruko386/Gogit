package history

import (
	"bufio"
	"encoding/json"
	"fmt"
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

func TestStoreCompactsPersistedHistoryPastThreshold(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	store := NewStore(path, 3)

	for index := range 6 {
		if err := store.Append(fmt.Sprintf("command-%d", index)); err != nil {
			t.Fatal(err)
		}
	}

	if got := countHistoryRecords(t, path); got != 6 {
		t.Fatalf("record count at threshold = %d, want 6", got)
	}

	if err := store.Append("command-6"); err != nil {
		t.Fatal(err)
	}

	if got := countHistoryRecords(t, path); got != 3 {
		t.Fatalf("record count after compaction = %d, want 3", got)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"command-4", "command-5", "command-6"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() after compaction = %#v, want %#v", got, want)
	}
}

func countHistoryRecords(t *testing.T, path string) int {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var command string
		if err := json.Unmarshal(scanner.Bytes(), &command); err != nil {
			t.Fatalf("decode history record: %v", err)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	return count
}
