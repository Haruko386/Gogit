package terminal

import (
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

	for _, expected := range []string{"git branch --sh", "--show-current", "Print the current branch."} {
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
