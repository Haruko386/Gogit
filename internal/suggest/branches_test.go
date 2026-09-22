package suggest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseBranches(t *testing.T) {
	output := []byte(
		" \trefs/heads/dev\tdev\n" +
			"*\trefs/heads/main\tmain\n" +
			" \trefs/remotes/origin/HEAD\torigin/HEAD\n" +
			" \trefs/remotes/origin/dev\torigin/dev\n",
	)

	got := parseBranches(output)
	if len(got) != 3 {
		t.Fatalf("parseBranches() returned %d branches: %#v", len(got), got)
	}

	tests := []struct {
		value       string
		description string
	}{
		{value: "dev", description: "Local branch."},
		{value: "main", description: "Current local branch."},
		{value: "origin/dev", description: "Remote-tracking branch."},
	}

	for index, test := range tests {
		if got[index].Value != test.value {
			t.Fatalf("branch %d value = %q, want %q", index, got[index].Value, test.value)
		}
		if got[index].Description != test.description {
			t.Fatalf(
				"branch %d description = %q, want %q",
				index,
				got[index].Description,
				test.description,
			)
		}
		if got[index].Kind != KindBranch {
			t.Fatalf("branch %d kind = %q, want %q", index, got[index].Kind, KindBranch)
		}
	}
}

func TestParseRemoteSuggestions(t *testing.T) {
	got := parseNamedSuggestions(
		[]byte("origin\nupstream\n"),
		KindRemote,
		"Remote repository.",
	)
	want := []string{"origin", "upstream"}

	if len(got) != len(want) {
		t.Fatalf("parseNamedSuggestions() returned %d values: %#v", len(got), got)
	}
	for index, value := range want {
		if got[index].Value != value {
			t.Fatalf("remote %d = %q, want %q", index, got[index].Value, value)
		}
		if got[index].Kind != KindRemote {
			t.Fatalf("remote %d kind = %q, want %q", index, got[index].Kind, KindRemote)
		}
	}
}

func TestParseTagSuggestions(t *testing.T) {
	got := parseNamedSuggestions(
		[]byte("v1.0.0\n\nv1.1.0\n"),
		KindTag,
		"Repository tag.",
	)
	want := []string{"v1.0.0", "v1.1.0"}

	if len(got) != len(want) {
		t.Fatalf("parseNamedSuggestions() returned %d values: %#v", len(got), got)
	}
	for index, value := range want {
		if got[index].Value != value {
			t.Fatalf("tag %d = %q, want %q", index, got[index].Value, value)
		}
		if got[index].Kind != KindTag {
			t.Fatalf("tag %d kind = %q, want %q", index, got[index].Kind, KindTag)
		}
	}
}

func TestParseNullSeparatedSuggestions(t *testing.T) {
	got := parseNullSeparatedSuggestions(
		[]byte("README.md\x00docs/release notes.md\x00README.md\x00bad\x1b[31m\x00"),
		KindFile,
		"Repository file.",
	)
	want := []string{"README.md", "docs/release notes.md"}

	if len(got) != len(want) {
		t.Fatalf("parseNullSeparatedSuggestions() = %#v, want %q", got, want)
	}
	for index, value := range want {
		if got[index].Value != value || got[index].Kind != KindFile {
			t.Fatalf("file %d = %#v, want %q with kind %q", index, got[index], value, KindFile)
		}
	}
}

func TestLoadRepositoryCandidates(t *testing.T) {
	directory := t.TempDir()
	runGit := func(arguments ...string) {
		t.Helper()

		commandArguments := append([]string{"-C", directory}, arguments...)
		command := exec.Command("git", commandArguments...)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", arguments, err, output)
		}
	}

	runGit("init")
	modifiedPath := filepath.Join(directory, "modified file.txt")
	deletedPath := filepath.Join(directory, "deleted.txt")
	if err := os.WriteFile(modifiedPath, []byte("original\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deletedPath, []byte("delete me\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit("add", ".")
	runGit(
		"-c", "user.name=Gogit Test",
		"-c", "user.email=gogit@example.invalid",
		"commit", "-m", "initial",
	)
	runGit("remote", "add", "origin", "https://example.invalid/repository.git")
	runGit("tag", "v1.0.0")
	if err := os.WriteFile(modifiedPath, []byte("changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(deletedPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "untracked.txt"), []byte("new\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := LoadRepositoryCandidates(context.Background(), directory)
	if err != nil {
		t.Fatal(err)
	}

	assertCandidate := func(candidates []Suggestion, value string, kind Kind) {
		t.Helper()
		for _, candidate := range candidates {
			if candidate.Value == value && candidate.Kind == kind {
				return
			}
		}
		t.Fatalf("candidate %q with kind %q not found in %#v", value, kind, candidates)
	}

	assertCandidate(got.Remotes, "origin", KindRemote)
	assertCandidate(got.Tags, "v1.0.0", KindTag)
	assertCandidate(got.Files, "modified file.txt", KindFile)
	assertCandidate(got.Files, "deleted.txt", KindFile)
	assertCandidate(got.Files, "untracked.txt", KindFile)
	assertCandidate(got.RestorableFiles, "modified file.txt", KindFile)
	assertCandidate(got.RestorableFiles, "deleted.txt", KindFile)
	for _, candidate := range got.RestorableFiles {
		if candidate.Value == "untracked.txt" {
			t.Fatalf("untracked file was marked restorable: %#v", got.RestorableFiles)
		}
	}
	if len(got.Branches) == 0 {
		t.Fatal("repository branch was not loaded")
	}
}
