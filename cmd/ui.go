package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Haruko386/Gogit/internal/editor"
	"github.com/Haruko386/Gogit/internal/history"
	"github.com/Haruko386/Gogit/internal/protocol"
	"github.com/Haruko386/Gogit/internal/session"
	"github.com/Haruko386/Gogit/internal/suggest"
	"github.com/Haruko386/Gogit/internal/terminal"
	"github.com/charmbracelet/x/ansi"
)

const (
	colorCyan   = "\x1b[36m"
	colorYellow = "\x1b[33m"
	colorReset  = "\x1b[0m"
)

type streamEvent struct {
	data []byte
	err  error
}

type branchLoadResult struct {
	generation uint64
	branches   []suggest.Suggestion
}

var (
	bracketedPasteStart = []byte("\x1b[200~")
	bracketedPasteEnd   = []byte("\x1b[201~")
)

// bracketedPasteState keeps enough trailing input to recognize paste markers
// even when the terminal splits an escape sequence across multiple reads.
type bracketedPasteState struct {
	active bool
	tail   []byte
}

func (s *bracketedPasteState) observe(data []byte) {
	markerSize := max(len(bracketedPasteStart), len(bracketedPasteEnd))

	for _, value := range data {
		s.tail = append(s.tail, value)

		switch {
		case bytes.HasSuffix(s.tail, bracketedPasteStart):
			s.active = true
			s.tail = nil
		case bytes.HasSuffix(s.tail, bracketedPasteEnd):
			s.active = false
			s.tail = nil
		case len(s.tail) >= markerSize:
			s.tail = append(s.tail[:0], s.tail[len(s.tail)-markerSize+1:]...)
		}
	}
}

func runShellUI(shellSession session.ShellSession, marker string, resizeDone <-chan error) (resizeFinished bool, resultErr error) {
	inputEvents := readStream(os.Stdin)
	outputEvents := readStream(shellSession)
	shellDone := make(chan error, 1)
	branchResults := make(chan branchLoadResult, 1)

	go func() {
		shellDone <- shellSession.Wait()
	}()

	var (
		decoder          terminal.Decoder
		lineEditor       editor.Editor
		commandHistory   history.History
		renderer         terminal.Renderer
		markerScan       = protocol.NewScanner(marker)
		selected         = -1
		suggestionMode   bool
		editing          bool
		prompt           = colorCyan + "(Gogit)" + colorReset + " "
		promptWidth      = 8
		branches         []suggest.Suggestion
		branchCancel     context.CancelFunc
		branchGeneration uint64
		pendingKeys      []terminal.Key
		pasteState       bracketedPasteState
	)
	defer func() {
		if branchCancel != nil {
			branchCancel()
		}
	}()

	loadBranches := func(directory string) {
		if branchCancel != nil {
			branchCancel()
		}

		branchGeneration++
		generation := branchGeneration
		branches = nil

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		branchCancel = cancel

		go func() {
			defer cancel()

			loaded, _ := suggest.LoadBranches(ctx, directory)
			select {
			case branchResults <- branchLoadResult{
				generation: generation,
				branches:   loaded,
			}:
			case <-ctx.Done():
			}
		}()
	}

	analyze := func() suggest.Result {
		return suggest.AnalyzeWithBranches(
			lineEditor.Line(),
			lineEditor.Cursor(),
			branches,
		)
	}

	render := func() error {
		result := analyze()
		suggestions := result.Suggestions

		if len(suggestions) == 0 {
			suggestionMode = false
			selected = -1
		} else if suggestionMode {
			if selected < 0 || selected >= len(suggestions) {
				selected = 0
			}
		} else {
			selected = -1
		}

		return writeOutput(renderer.Render(terminal.View{
			Prompt:      prompt,
			PromptWidth: promptWidth,
			Line:        lineEditor.Line(),
			Cursor:      lineEditor.Cursor(),
			Suggestions: suggestions,
			Selected:    selected,
			Hint:        result.Hint,
		}))
	}

	// handleEditingKeys consumes one batch of decoded input. If the batch
	// submits a command, the caller keeps every key after Enter in pendingKeys
	// and replays it when the shell emits its next prompt. Input arriving in a
	// later terminal read while the command runs is passed through to the PTY
	// unless it belongs to the same bracketed paste.
	handleEditingKeys := func(keys []terminal.Key) (
		changed bool,
		consumed int,
		err error,
	) {
		for index, key := range keys {
			consumed = index + 1

			switch key.Type {
			case terminal.KeyRune:
				lineEditor.Insert(key.Rune)
				suggestionMode = false
				selected = -1
				changed = true
			case terminal.KeyBackspace:
				changed = lineEditor.Backspace() || changed
				suggestionMode = false
				selected = -1
			case terminal.KeyDelete:
				changed = lineEditor.Delete() || changed
				suggestionMode = false
				selected = -1
			case terminal.KeyLeft:
				changed = lineEditor.MoveLeft() || changed
				suggestionMode = false
				selected = -1
			case terminal.KeyRight:
				changed = lineEditor.MoveRight() || changed
				suggestionMode = false
				selected = -1
			case terminal.KeyHome:
				changed = lineEditor.MoveHome() || changed
				suggestionMode = false
				selected = -1
			case terminal.KeyEnd:
				changed = lineEditor.MoveEnd() || changed
				suggestionMode = false
				selected = -1
			case terminal.KeyUp:
				suggestions := analyze().Suggestions

				changed = navigateUp(&lineEditor, &commandHistory, suggestions, &selected, &suggestionMode) || changed
			case terminal.KeyDown:
				suggestions := analyze().Suggestions

				changed = navigateDown(&lineEditor, &commandHistory, suggestions, &selected, &suggestionMode) || changed
			case terminal.KeyTab:
				suggestions := analyze().Suggestions
				if len(suggestions) == 0 {
					continue
				}
				if selected < 0 || selected >= len(suggestions) {
					selected = 0
				}

				context, ok := suggest.ParseContext(
					lineEditor.Line(),
					lineEditor.Cursor(),
				)
				if !ok {
					continue
				}

				lineEditor.Replace(
					context.TokenStart,
					context.TokenEnd,
					suggestions[selected].Value,
				)

				suggestionMode = false
				selected = -1
				changed = true

			case terminal.KeyCtrlC:
				if err := writeOutput(renderer.Clear()); err != nil {
					return changed, consumed, err
				}
				if err := writeOutput(prompt + "^C\r\n"); err != nil {
					return changed, consumed, err
				}

				lineEditor.Clear()
				commandHistory.Reset()
				suggestionMode = false
				selected = -1
				changed = true

			case terminal.KeyCtrlD:
				if lineEditor.Line() == "" {
					if err := writeOutput(renderer.Clear()); err != nil {
						return changed, consumed, err
					}
					return changed, consumed, io.EOF
				}

				changed = lineEditor.Delete() || changed
				suggestionMode = false
				selected = -1

			case terminal.KeyEnter:
				if err := writeOutput(renderer.Clear()); err != nil {
					return changed, consumed, err
				}
				if err := writeOutput(prompt); err != nil {
					return changed, consumed, err
				}

				command := lineEditor.Line()
				commandHistory.Add(command)
				command += "\r"

				if err := writeAll(shellSession, []byte(command)); err != nil {
					return changed, consumed, fmt.Errorf(
						"submit command: %w",
						err,
					)
				}

				lineEditor.Clear()
				commandHistory.Reset()
				suggestionMode = false
				selected = -1
				editing = false
				return false, consumed, nil
			}
		}

		return changed, consumed, nil
	}

	replayPendingKeys := func() (bool, error) {
		if len(pendingKeys) == 0 || pasteState.active || !editing {
			return false, nil
		}

		keys := pendingKeys
		pendingKeys = nil

		changed, consumed, err := handleEditingKeys(keys)
		if consumed < len(keys) {
			pendingKeys = append(pendingKeys, keys[consumed:]...)
		}
		return changed, err
	}

	for {
		select {
		case event := <-inputEvents:
			if event.err != nil {
				if errors.Is(event.err, io.EOF) {
					return false, nil
				}
				return false, fmt.Errorf(
					"read terminal input: %w",
					event.err,
				)
			}

			wasPasting := pasteState.active
			pasteState.observe(event.data)

			if !editing && !wasPasting {
				if err := writeAll(shellSession, event.data); err != nil {
					return false, fmt.Errorf(
						"write PTY input: %w",
						err,
					)
				}
				continue
			}
			if wasPasting {
				pendingKeys = append(pendingKeys, decoder.Feed(event.data)...)
				changed, err := replayPendingKeys()
				if errors.Is(err, io.EOF) {
					return false, nil
				}
				if err != nil {
					return false, err
				}
				if editing && changed {
					if err := render(); err != nil {
						return false, err
					}
				}
				continue
			}

			keys := decoder.Feed(event.data)
			changed, consumed, err := handleEditingKeys(keys)
			if errors.Is(err, io.EOF) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			if consumed < len(keys) {
				pendingKeys = append(pendingKeys, keys[consumed:]...)
			}

			if editing && changed {
				if err := render(); err != nil {
					return false, err
				}
			}

		case event := <-outputEvents:
			if len(event.data) > 0 {
				visible, prompts := markerScan.Push(event.data)

				if len(prompts) > 0 {
					prompt, promptWidth = formatPrompt(
						prompts[len(prompts)-1],
					)
					loadBranches(prompts[len(prompts)-1].Directory)
				}

				if len(visible) > 0 && editing {
					if err := writeOutput(renderer.Clear()); err != nil {
						return false, err
					}
				}

				if len(visible) > 0 {
					if err := writeOutput(string(visible)); err != nil {
						return false, err
					}
				}

				if len(prompts) > 0 {
					editing = true
					lineEditor.Clear()
					commandHistory.Reset()
					suggestionMode = false
					selected = -1

					if len(pendingKeys) > 0 && !pasteState.active {
						_, err := replayPendingKeys()
						if errors.Is(err, io.EOF) {
							return false, nil
						}
						if err != nil {
							return false, err
						}
					}
				}

				if editing && (len(visible) > 0 || len(prompts) > 0) {
					if err := render(); err != nil {
						return false, err
					}
				}
			}

			if event.err != nil {
				remaining := markerScan.Flush()
				if len(remaining) > 0 {
					if err := writeOutput(string(remaining)); err != nil {
						return false, err
					}
				}

				if errors.Is(event.err, io.EOF) {
					outputEvents = nil
					continue
				}

				return false, fmt.Errorf(
					"read PTY output: %w",
					event.err,
				)
			}

		case err := <-shellDone:
			return false, err

		case result := <-branchResults:
			if result.generation != branchGeneration {
				continue
			}
			branches = result.branches
			if editing && lineEditor.Line() != "" {
				if err := render(); err != nil {
					return false, err
				}
			}

		case err := <-resizeDone:
			if err != nil {
				return true, fmt.Errorf(
					"watch terminal resize: %w",
					err,
				)
			}
			return true, nil
		}
	}
}

func formatPrompt(state protocol.Prompt) (string, int) {
	var output strings.Builder

	output.WriteString(colorCyan)
	output.WriteString("(Gogit)")
	output.WriteString(colorReset)
	output.WriteByte(' ')

	if state.Environment != "" {
		output.WriteString(colorYellow)
		output.WriteByte('(')
		output.WriteString(state.Environment)
		output.WriteByte(')')
		output.WriteString(colorReset)
		output.WriteByte(' ')
	}

	output.WriteString(state.Directory)
	output.WriteString("> ")

	prompt := output.String()
	return prompt, ansi.StringWidth(prompt)
}

func navigateUp(lineEditor *editor.Editor, commandHistory *history.History, suggestions []suggest.Suggestion, selected *int, suggestionMode *bool) bool {
	if *suggestionMode && len(suggestions) > 0 {
		*selected--
		if *selected < 0 {
			*selected = len(suggestions) - 1
		}
		return true
	}

	// Up never starts suggestion navigation. Outside suggestion mode it
	// behaves like a regular shell and navigates to older history.
	*suggestionMode = false
	*selected = -1

	command, ok := commandHistory.Previous(lineEditor.Line())
	if !ok {
		return false
	}

	lineEditor.SetLine(command)
	return true
}

func navigateDown(lineEditor *editor.Editor, commandHistory *history.History, suggestions []suggest.Suggestion, selected *int, suggestionMode *bool) bool {
	if *suggestionMode && len(suggestions) > 0 {
		*selected++
		if *selected >= len(suggestions) {
			*selected = 0
		}
		return true
	}

	// Once history browsing has started, Down must continue toward newer
	// history and eventually restore the user's original draft.
	if commandHistory.Browsing() {
		command, ok := commandHistory.Next()
		if !ok {
			return false
		}

		lineEditor.SetLine(command)
		*suggestionMode = false
		*selected = -1
		return true
	}

	// Down explicitly enters suggestion navigation.
	if len(suggestions) > 0 {
		*suggestionMode = true
		*selected = 0
		return true
	}

	return false
}

func readStream(reader io.Reader) <-chan streamEvent {
	events := make(chan streamEvent, 1)

	go func() {
		buffer := make([]byte, 4096)

		for {
			count, err := reader.Read(buffer)

			if count > 0 {
				data := append([]byte(nil), buffer[:count]...)
				events <- streamEvent{data: data}
			}

			if err != nil {
				events <- streamEvent{err: err}
				return
			}
		}
	}()

	return events
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		count, err := writer.Write(data)
		if err != nil {
			return err
		}
		if count == 0 {
			return io.ErrShortWrite
		}

		data = data[count:]
	}

	return nil
}

func writeOutput(value string) error {
	if err := writeAll(os.Stdout, []byte(value)); err != nil {
		return fmt.Errorf("write terminal output: %w", err)
	}
	return nil
}
