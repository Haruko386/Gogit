package terminal

import (
	"fmt"
	"strings"

	"github.com/Haruko386/Gogit/internal/suggest"
)

const (
	hideCursor = "\x1b[?25l"
	showCursor = "\x1b[?25h"
	clearLine  = "\r\x1b[2K"
	colorCyan  = "\x1b[36m"
	colorGreen = "\x1b[32m"
	colorGray  = "\x1b[90m"
	colorReset = "\x1b[0m"
)

// View contains everything needed to draw one editor frame. PromptWidth is
// the visible cell width and excludes ANSI color bytes in Prompt.
type View struct {
	Prompt      string
	PromptWidth int
	Line        string
	Cursor      int
	Suggestions []suggest.Suggestion
	Selected    int
}

// Renderer redraws one input line and its suggestion area.
type Renderer struct {
	previousSuggestionRows int
}

// Clear removes the current input line and all suggestion rows.
func (r *Renderer) Clear() string {
	var output strings.Builder

	output.WriteString(hideCursor)
	output.WriteString(clearLine)

	for range r.previousSuggestionRows {
		output.WriteString("\r\n\x1b[2K")
	}

	if r.previousSuggestionRows > 0 {
		fmt.Fprintf(&output, "\x1b[%dA", r.previousSuggestionRows)
	}

	output.WriteByte('\r')
	output.WriteString(showCursor)

	r.previousSuggestionRows = 0
	return output.String()
}

// Render returns the ANSI sequence for the next complete frame.
func (r *Renderer) Render(view View) string {
	selected := view.Selected
	hasSelection := selected >= 0 && selected < len(view.Suggestions)

	descriptionIndex := selected
	if !hasSelection && len(view.Suggestions) > 0 {
		descriptionIndex = 0
	}

	currentRows := 0
	if len(view.Suggestions) > 0 {
		currentRows = len(view.Suggestions) + 1
	}
	rowsToClear := max(currentRows, r.previousSuggestionRows)

	var output strings.Builder
	output.WriteString(hideCursor)
	output.WriteString(clearLine)
	output.WriteString(view.Prompt)
	output.WriteString(view.Line)

	for index := range rowsToClear {
		output.WriteString("\r\n\x1b[2K")
		if index == len(view.Suggestions) && len(view.Suggestions) > 0 {
			description := view.Suggestions[descriptionIndex].Description
			output.WriteString("    ")
			output.WriteString(colorGray)
			output.WriteString(description)
			output.WriteString(colorReset)
			continue
		}
		if index >= len(view.Suggestions) {
			continue
		}

		candidate := view.Suggestions[index]
		if hasSelection && index == selected {
			output.WriteString(colorGreen)
			output.WriteString("  > ")
		} else {
			output.WriteString("    ")
		}
		output.WriteString(colorCyan)
		output.WriteString(candidate.Value)
		output.WriteString(colorReset)
	}

	if rowsToClear > 0 {
		fmt.Fprintf(&output, "\x1b[%dA", rowsToClear)
	}
	output.WriteByte('\r')

	cursor := min(max(view.Cursor, 0), len([]rune(view.Line)))
	cursorColumn := view.PromptWidth + cursor
	if cursorColumn > 0 {
		fmt.Fprintf(&output, "\x1b[%dC", cursorColumn)
	}
	output.WriteString(showCursor)

	r.previousSuggestionRows = currentRows
	return output.String()
}
