package terminal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Haruko386/Gogit/internal/suggest"
)

func TestRendererDrawsSuggestionAndDescription(t *testing.T) {
	var renderer Renderer
	frame := renderer.Render(View{
		Prompt:      "\x1b[36m(Gogit)\x1b[0m > ",
		PromptWidth: 10,
		Line:        "git branch --sh",
		Cursor:      16,
		Suggestions: []suggest.Suggestion{
			{Value: "--show-current", Description: "Print the current branch."},
		},
	})

	for _, expected := range []string{" branch --sh", "--show-current", "Print the current branch."} {
		if !strings.Contains(frame, expected) {
			t.Fatalf("rendered frame does not contain %q: %q", expected, frame)
		}
	}
}

func TestRendererClearsPreviousSuggestionRows(t *testing.T) {
	var renderer Renderer
	renderer.Render(View{
		Prompt:      "> ",
		PromptWidth: 2,
		Suggestions: []suggest.Suggestion{
			{Value: "--merged"},
			{Value: "--no-merged"},
		},
	})

	frame := renderer.Render(View{Prompt: "> ", PromptWidth: 2})
	if got, want := strings.Count(frame, "\r\n\x1b[2K"), 3; got != want {
		t.Fatalf("clear row count = %d, want %d", got, want)
	}
	if !strings.Contains(frame, "\x1b[3A") {
		t.Fatalf("frame does not move back over stale rows: %q", frame)
	}
}

func TestRendererClearRemovesSuggestionRows(t *testing.T) {
	var renderer Renderer

	renderer.Render(View{
		Prompt:      "(Gogit) ",
		PromptWidth: 8,
		Line:        "git br",
		Cursor:      6,
		Suggestions: []suggest.Suggestion{
			{
				Value:       "branch",
				Description: "List, create, or delete branches",
				Kind:        suggest.KindSubcommand,
			},
		},
	})

	output := renderer.Clear()

	if !strings.Contains(output, "\x1b[2A") {
		t.Fatalf("Clear() did not move over two suggestion rows: %q", output)
	}
	if renderer.previousSuggestionRows != 0 {
		t.Fatalf(
			"previousSuggestionRows = %d, want 0",
			renderer.previousSuggestionRows,
		)
	}
}

func TestRendererDrawsValueHint(t *testing.T) {
	var renderer Renderer

	output := renderer.Render(View{
		Prompt:      "(Gogit) ",
		PromptWidth: 8,
		Line:        "git commit --message ",
		Cursor:      len([]rune("git commit --message ")),
		Hint: &suggest.ValueHint{
			Name:        "message",
			Description: "Enter the commit message.",
		},
	})

	if !strings.Contains(output, "<message>") {
		t.Fatalf("Render() did not draw the value name: %q", output)
	}
	if !strings.Contains(output, "Enter the commit message.") {
		t.Fatalf("Render() did not draw the value description: %q", output)
	}
}

func TestRendererHighlightsGitAndQuotedValue(t *testing.T) {
	var renderer Renderer

	output := renderer.Render(View{
		Prompt:      "> ",
		PromptWidth: 2,
		Line:        `git commit -m "message:xx"`,
		Cursor:      len([]rune(`git commit -m "message:xx"`)),
	})

	if !strings.Contains(output, colorCyan+"git"+colorReset) {
		t.Fatalf("Render() did not highlight git: %q", output)
	}
	if !strings.Contains(output, colorYellow+`"message:xx"`+colorReset) {
		t.Fatalf("Render() did not highlight quoted value: %q", output)
	}
}

func TestHighlightInputRequiresCompleteGitToken(t *testing.T) {
	for _, line := range []string{"g", "gitty status", `"git" status`} {
		if output := highlightInput(line); strings.Contains(output, colorCyan) {
			t.Fatalf("highlightInput(%q) highlighted a non-git token: %q", line, output)
		}
	}
}

func TestHighlightInputHandlesLeadingSpaceAndUnclosedQuote(t *testing.T) {
	line := `  git commit -m "unfinished`
	output := highlightInput(line)

	if !strings.Contains(output, `  `+colorCyan+"git"+colorReset) {
		t.Fatalf("leading-space git token was not highlighted: %q", output)
	}
	if !strings.Contains(output, colorYellow+`"unfinished`+colorReset) {
		t.Fatalf("unclosed quoted value was not highlighted: %q", output)
	}
}

func TestRendererPositionsCursorByTerminalCellWidth(t *testing.T) {
	tests := []struct {
		name string
		line string
		want int
	}{
		{name: "ASCII", line: "abc", want: 5},
		{name: "Chinese", line: "a中", want: 5},
		{name: "combining character", line: "e\u0301", want: 3},
		{name: "emoji", line: "🚀", want: 4},
		{name: "emoji sequence", line: "👨‍👩‍👧‍👦", want: 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var renderer Renderer
			output := renderer.Render(View{
				Prompt:      "> ",
				PromptWidth: 2,
				Line:        test.line,
				Cursor:      len([]rune(test.line)),
			})

			want := fmt.Sprintf("\r\x1b[%dC%s", test.want, showCursor)
			if !strings.Contains(output, want) {
				t.Fatalf("cursor sequence not found in %q; want %q", output, want)
			}
		})
	}
}
