package cmd

import (
	"fmt"
	"testing"

	"github.com/Haruko386/Gogit/internal/suggest"
)

func TestCompletionInsertionValueQuotesRepositoryCandidates(t *testing.T) {
	for _, kind := range []suggest.Kind{
		suggest.KindBranch,
		suggest.KindRemote,
		suggest.KindTag,
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

func TestCompletionInsertionValueUsesLiteralPathspecForFiles(t *testing.T) {
	candidate := suggest.Suggestion{
		Value:       "docs/release*.md",
		Description: "Repository file.",
		Kind:        suggest.KindFile,
	}

	want := quoteCommandArgument(":(literal)" + candidate.Value)
	if got := completionInsertionValue(candidate); got != want {
		t.Fatalf("completionInsertionValue() = %q, want %q", got, want)
	}
	if candidate.Value != "docs/release*.md" || candidate.Description != "Repository file." {
		t.Fatalf("completion changed candidate metadata: %#v", candidate)
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

func TestCompletionReplacementValueAddsContextualSpace(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		cursor    int
		candidate suggest.Suggestion
		want      string
	}{
		{
			name:   "subcommand at line end",
			line:   "git com",
			cursor: len([]rune("git com")),
			candidate: suggest.Suggestion{
				Value: "commit",
				Kind:  suggest.KindSubcommand,
			},
			want: "commit ",
		},
		{
			name:   "option requiring value",
			line:   "git clone --dep",
			cursor: len([]rune("git clone --dep")),
			candidate: suggest.Suggestion{
				Value:      "--depth",
				Kind:       suggest.KindOption,
				TakesValue: true,
			},
			want: "--depth ",
		},
		{
			name:   "existing following argument",
			line:   "git com --amend",
			cursor: len([]rune("git com")),
			candidate: suggest.Suggestion{
				Value: "commit",
				Kind:  suggest.KindSubcommand,
			},
			want: "commit",
		},
		{
			name:   "inline option value",
			line:   "git log --for",
			cursor: len([]rune("git log --for")),
			candidate: suggest.Suggestion{
				Value:      "--format=",
				Kind:       suggest.KindOption,
				TakesValue: true,
			},
			want: "--format=",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, ok := suggest.ParseContext(test.line, test.cursor)
			if !ok {
				t.Fatal("ParseContext() returned false")
			}

			got := completionReplacementValue(
				test.line,
				context,
				test.candidate,
			)
			if got != test.want {
				t.Fatalf(
					"completionReplacementValue() = %q, want %q",
					got,
					test.want,
				)
			}
		})
	}
}

func TestSuggestionViewportFollowsSelection(t *testing.T) {
	suggestions := make([]suggest.Suggestion, 10)
	for index := range suggestions {
		suggestions[index] = suggest.Suggestion{
			Value: fmt.Sprintf("suggestion-%d", index),
		}
	}

	tests := []struct {
		name         string
		selected     int
		wantOffset   int
		wantSelected int
		wantFirst    string
		wantLast     string
	}{
		{
			name:         "no selection",
			selected:     -1,
			wantOffset:   0,
			wantSelected: -1,
			wantFirst:    "suggestion-0",
			wantLast:     "suggestion-5",
		},
		{
			name:         "last item in first window",
			selected:     5,
			wantOffset:   0,
			wantSelected: 5,
			wantFirst:    "suggestion-0",
			wantLast:     "suggestion-5",
		},
		{
			name:         "scrolls for seventh item",
			selected:     6,
			wantOffset:   1,
			wantSelected: 5,
			wantFirst:    "suggestion-1",
			wantLast:     "suggestion-6",
		},
		{
			name:         "shows final window",
			selected:     9,
			wantOffset:   4,
			wantSelected: 5,
			wantFirst:    "suggestion-4",
			wantLast:     "suggestion-9",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viewport := makeSuggestionViewport(
				suggestions,
				test.selected,
				6,
			)

			if len(viewport.suggestions) != 6 {
				t.Fatalf(
					"visible count = %d, want 6",
					len(viewport.suggestions),
				)
			}
			if viewport.offset != test.wantOffset {
				t.Fatalf(
					"offset = %d, want %d",
					viewport.offset,
					test.wantOffset,
				)
			}
			if viewport.selected != test.wantSelected {
				t.Fatalf(
					"selected = %d, want %d",
					viewport.selected,
					test.wantSelected,
				)
			}
			if viewport.suggestions[0].Value != test.wantFirst {
				t.Fatalf(
					"first suggestion = %q, want %q",
					viewport.suggestions[0].Value,
					test.wantFirst,
				)
			}
			if viewport.suggestions[5].Value != test.wantLast {
				t.Fatalf(
					"last suggestion = %q, want %q",
					viewport.suggestions[5].Value,
					test.wantLast,
				)
			}
		})
	}
}
