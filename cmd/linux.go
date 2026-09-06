//go:build !windows

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/Haruko386/Gogit/internal/protocol"
)

func systemShell(marker string) *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	begin := protocol.BeginMarker(marker)
	end := protocol.EndMarker(marker)
	separator := string(rune(0x1f))

	var (
		arguments []string
		prompt    string
	)

	switch filepath.Base(shell) {
	case "bash":
		arguments = []string{"--norc", "-i"}
		prompt = begin +
			`$(if [ -n "$CONDA_DEFAULT_ENV" ]; then ` +
			`printf "%s" "$CONDA_DEFAULT_ENV"; ` +
			`elif [ -n "$VIRTUAL_ENV" ]; then ` +
			`basename "$VIRTUAL_ENV"; fi)` +
			separator +
			`\w` +
			end

	case "zsh":
		arguments = []string{
			"-f",
			"-o",
			"PROMPT_SUBST",
			"-i",
		}
		prompt = begin +
			`${CONDA_DEFAULT_ENV:-${VIRTUAL_ENV:t}}` +
			separator +
			`%~` +
			end

	default:
		arguments = []string{"-i"}
		prompt = begin +
			`$(if [ -n "$CONDA_DEFAULT_ENV" ]; then ` +
			`printf "%s" "$CONDA_DEFAULT_ENV"; ` +
			`elif [ -n "$VIRTUAL_ENV" ]; then ` +
			`basename "$VIRTUAL_ENV"; fi)` +
			separator +
			`$(pwd)` +
			end
	}

	command := exec.Command(shell, arguments...)
	// The shell runs on xpty's slave terminal. Give it its own session and
	// controlling terminal so an interactive shell cannot apply job-control
	// signals (notably SIGTTOU) to Gogit itself. This matters when Gogit is
	// launched from another PTY owner such as VHS, tmux, or an SSH recorder.
	command.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    0,
	}
	command.Env = append(
		os.Environ(),
		"PS1="+prompt,
		"PROMPT="+prompt,
		"CONDA_CHANGEPS1=false",
		"VIRTUAL_ENV_DISABLE_PROMPT=1",
	)

	return command
}
