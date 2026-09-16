package suggest

import (
	"context"
	"os/exec"
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
	runGit(
		"-c", "user.name=Gogit Test",
		"-c", "user.email=gogit@example.invalid",
		"commit", "--allow-empty", "-m", "initial",
	)
	runGit("remote", "add", "origin", "https://example.invalid/repository.git")
	runGit("tag", "v1.0.0")

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
	if len(got.Branches) == 0 {
		t.Fatal("repository branch was not loaded")
	}
}
