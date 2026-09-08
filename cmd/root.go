package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Haruko386/Gogit/internal/protocol"
	"github.com/Haruko386/Gogit/internal/session"
	"github.com/charmbracelet/x/term"
)

const (
	fallbackTerminalWidth  = 80
	fallbackTerminalHeight = 24
)

func Run() error {
	return runPersistentShell()
}

func runPersistentShell() (resultErr error) {
	inputFD := os.Stdin.Fd()
	outputFD := os.Stdout.Fd()

	if !term.IsTerminal(inputFD) {
		return errors.New("standard input is not a terminal")
	}

	width, height, err := term.GetSize(outputFD)
	if err != nil {
		width = fallbackTerminalWidth
		height = fallbackTerminalHeight
	}

	marker, err := protocol.NewMarker()
	if err != nil {
		return err
	}

	shellCommand, cleanupShell, err := systemShell(marker)
	if err != nil {
		return err
	}
	defer cleanupShell()

	shellSession := session.New(shellCommand, width, height)

	if err := shellSession.Start(); err != nil {
		return fmt.Errorf("start shell: %w", err)
	}

	defer func() {
		resultErr = errors.Join(resultErr, shellSession.Close())
	}()

	oldState, err := term.MakeRaw(inputFD)
	if err != nil {
		return fmt.Errorf("enable terminal raw mode: %w", err)
	}

	defer func() {
		resultErr = errors.Join(resultErr, term.Restore(inputFD, oldState))
	}()

	resizeCtx, stopResize := context.WithCancel(context.Background())
	resizeDone := watchTerminalResize(resizeCtx, outputFD, shellSession, width, height)

	resizeFinished, runErr := runShellUI(
		shellSession,
		marker,
		resizeDone,
	)

	stopResize()

	if !resizeFinished {
		if resizeErr := <-resizeDone; resizeErr != nil {
			runErr = errors.Join(runErr, fmt.Errorf("watch terminal resize: %w", resizeErr))
		}
	}

	return runErr
}
