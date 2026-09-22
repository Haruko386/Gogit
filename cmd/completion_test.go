package cmd

import (
	"testing"

	"github.com/Haruko386/Gogit/internal/suggest"
)

func TestCompletionInsertionValueQuotesRepositoryCandidates(t *testing.T) {
	for _, kind := range []suggest.Kind{
		suggest.KindBranch,
		suggest.KindRemote,
		suggest.KindTag,
		suggest.KindFile,
	} {
		candidate := suggest.Suggestion{
			Value:       "release; echo unsafe",
			Description: "unchanged metadata",
			Kind:        kind,
		}

		if got, want := completionInsertionValue(candidate), quoteCommandArgument(candidate.Value); got != want {
			t.Errorf("completionInsertionValue() = %q, want %q", got, want)
		}
		if candidate.Value != "release; echo unsafe" || candidate.Description != "unchanged metadata" {
			t.Fatalf("completion changed candidate metadata: %#v", candidate)
		}
	}
}

func TestCompletionInsertionValueLeavesSafeRepositoryNamesUnquoted(t *testing.T) {
	for _, value := range []string{
		"main",
		"origin",
		"origin/feature-1.2",
	} {
		candidate := suggest.Suggestion{
			Value: value,
			Kind:  suggest.KindBranch,
		}

		if got := completionInsertionValue(candidate); got != value {
			t.Errorf("completionInsertionValue(%q) = %q", value, got)
		}
	}
}

func TestCompletionInsertionValueLeavesStaticCandidatesUnchanged(t *testing.T) {
	candidate := suggest.Suggestion{
		Value: "--show-current",
		Kind:  suggest.KindOption,
	}

	if got := completionInsertionValue(candidate); got != candidate.Value {
		t.Fatalf("completionInsertionValue() = %q, want %q", got, candidate.Value)
	}
}
