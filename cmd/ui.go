package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Haruko386/Gogit/internal/editor"
	"github.com/Haruko386/Gogit/internal/history"
	"github.com/Haruko386/Gogit/internal/protocol"
	"github.com/Haruko386/Gogit/internal/session"
	"github.com/Haruko386/Gogit/internal/suggest"
	"github.com/Haruko386/Gogit/internal/terminal"
)

const (
	gogitPrompt      = "\x1b[36m(Gogit)\x1b[0m "
	gogitPromptWidth = 8
)

type streamEvent struct {
	data []byte
	err  error
}

func runShellUI(shellSession session.ShellSession, marker string, resizeDone <-chan error) (resizeFinished bool, resultErr error) {
	inputEvents := readStream(os.Stdin)
	outputEvents := readStream(shellSession)
	shellDone := make(chan error, 1)

	go func() {
		shellDone <- shellSession.Wait()
	}()

	var (
		decoder        terminal.Decoder
		lineEditor     editor.Editor
		commandHistory history.History
		renderer       terminal.Renderer
		markerScan     = protocol.NewScanner(marker)
		selected       = -1
		suggestionMode bool
		editing        bool
	)

	render := func() error {
		suggestions := suggest.Suggest(
			lineEditor.Line(),
			lineEditor.Cursor(),
		)

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
			Prompt:      gogitPrompt,
			PromptWidth: gogitPromptWidth,
			Line:        lineEditor.Line(),
			Cursor:      lineEditor.Cursor(),
			Suggestions: suggestions,
			Selected:    selected,
		}))
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

			if !editing {
				if err := writeAll(shellSession, event.data); err != nil {
					return false, fmt.Errorf(
						"write PTY input: %w",
						err,
					)
				}
				continue
			}

			keys := decoder.Feed(event.data)
			changed := false

		keyLoop:
			for _, key := range keys {
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
					suggestions := suggest.Suggest(
						lineEditor.Line(),
						lineEditor.Cursor(),
					)

					changed = navigateUp(&lineEditor, &commandHistory, suggestions, &selected, &suggestionMode) || changed
				case terminal.KeyDown:
					suggestions := suggest.Suggest(
						lineEditor.Line(),
						lineEditor.Cursor(),
					)

					changed = navigateDown(&lineEditor, &commandHistory, suggestions, &selected, &suggestionMode) || changed
				case terminal.KeyTab:
					suggestions := suggest.Suggest(
						lineEditor.Line(),
						lineEditor.Cursor(),
					)
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
						return false, err
					}
					if err := writeOutput("^C\r\n"); err != nil {
						return false, err
					}

					lineEditor.Clear()
					commandHistory.Reset()
					suggestionMode = false
					selected = -1
					changed = true

				case terminal.KeyCtrlD:
					if lineEditor.Line() == "" {
						if err := writeOutput(renderer.Clear()); err != nil {
							return false, err
						}
						return false, nil
					}

					changed = lineEditor.Delete() || changed

					suggestionMode = false
					selected = -1
				case terminal.KeyEnter:
					if err := writeOutput(renderer.Clear()); err != nil {
						return false, err
					}

					command := lineEditor.Line()
					// save command to history
					commandHistory.Add(command)
					command += "\r"

					if err := writeAll(
						shellSession,
						[]byte(command),
					); err != nil {
						return false, fmt.Errorf(
							"submit command: %w",
							err,
						)
					}

					lineEditor.Clear()
					commandHistory.Reset()
					suggestionMode = false
					selected = -1
					editing = false
					changed = false
					break keyLoop
				}
			}

			if editing && changed {
				if err := render(); err != nil {
					return false, err
				}
			}

		case event := <-outputEvents:
			if len(event.data) > 0 {
				visible, readyCount := markerScan.Push(event.data)

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

				if readyCount > 0 {
					editing = true
					lineEditor.Clear()
					commandHistory.Reset()
					suggestionMode = false
					selected = -1
				}

				if editing && (len(visible) > 0 || readyCount > 0) {
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
