package terminal

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Haruko386/Gogit/internal/suggest"
)

const (
	hideCursor  = "\x1b[?25l"
	showCursor  = "\x1b[?25h"
	clearLine   = "\r\x1b[2K"
	colorCyan   = "\x1b[36m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorGray   = "\x1b[90m"
	colorReset  = "\x1b[0m"
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
	Hint        *suggest.ValueHint
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
	} else if view.Hint != nil {
		currentRows = 1
	}
	rowsToClear := max(currentRows, r.previousSuggestionRows)

	var output strings.Builder
	output.WriteString(hideCursor)
	output.WriteString(clearLine)
	output.WriteString(view.Prompt)
	output.WriteString(highlightInput(view.Line))

	for index := range rowsToClear {
		output.WriteString("\r\n\x1b[2K")
		if len(view.Suggestions) == 0 && view.Hint != nil && index == 0 {
			output.WriteString("    ")
			output.WriteString(colorGray)
			output.WriteByte('<')
			output.WriteString(view.Hint.Name)
			output.WriteString(">  ")
			output.WriteString(view.Hint.Description)
			output.WriteString(colorReset)
			continue
		}
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

// highlightInput adds display-only syntax colors without changing the text or
// its rune indexes. The editor and cursor calculations continue to use the
// original, unstyled line.
func highlightInput(line string) string {
	runes := []rune(line)
	commandStart := 0
	for commandStart < len(runes) && unicode.IsSpace(runes[commandStart]) {
		commandStart++
	}

	commandEnd := commandStart
	for commandEnd < len(runes) && !unicode.IsSpace(runes[commandEnd]) {
		commandEnd++
	}
	highlightGit := string(runes[commandStart:commandEnd]) == "git"

	var output strings.Builder
	for index := 0; index < len(runes); {
		if highlightGit && index == commandStart {
			output.WriteString(colorCyan)
			output.WriteString("git")
			output.WriteString(colorReset)
			index = commandEnd
			continue
		}

		if runes[index] == '\'' || runes[index] == '"' {
			quote := runes[index]
			output.WriteString(colorYellow)
			output.WriteRune(quote)
			index++

			for index < len(runes) {
				value := runes[index]
				output.WriteRune(value)
				index++

				if (value == '\\' || value == '`') && index < len(runes) {
					output.WriteRune(runes[index])
					index++
					continue
				}
				if value == quote {
					break
				}
			}

			output.WriteString(colorReset)
			continue
		}

		output.WriteRune(runes[index])
		index++
	}

	return output.String()
}
