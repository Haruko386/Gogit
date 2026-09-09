//go:build !windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func systemShell(marker string) (*exec.Cmd, func(), error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	temporaryDirectory, err := os.MkdirTemp("", "gogit-shell-")
	if err != nil {
		return nil, nil, fmt.Errorf("create shell configuration: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(temporaryDirectory)
	}

	var arguments []string
	environment := append([]string(nil), os.Environ()...)

	switch filepath.Base(shell) {
	case "bash":
		initFile := filepath.Join(temporaryDirectory, "bashrc")
		if err := os.WriteFile(initFile, []byte(bashInitScript(marker)), 0o600); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("write Bash configuration: %w", err)
		}
		arguments = []string{"--rcfile", initFile, "-i"}

	case "zsh":
		zshEnvFile := filepath.Join(temporaryDirectory, ".zshenv")
		if err := os.WriteFile(zshEnvFile, []byte(zshEnvInitScript()), 0o600); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("write Zsh environment bootstrap: %w", err)
		}

		zshRCFile := filepath.Join(temporaryDirectory, ".zshrc")
		if err := os.WriteFile(zshRCFile, []byte(zshInitScript(marker)), 0o600); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("write Zsh configuration: %w", err)
		}

		userZDOTDIR, zdotdirWasSet := os.LookupEnv("ZDOTDIR")
		if !zdotdirWasSet {
			userZDOTDIR = os.Getenv("HOME")
		}
		environment = append(
			environment,
			"GOGIT_USER_ZDOTDIR="+userZDOTDIR,
			fmt.Sprintf("GOGIT_USER_ZDOTDIR_WAS_SET=%t", zdotdirWasSet),
			"ZDOTDIR="+temporaryDirectory,
		)
		arguments = []string{"-i"}

	default:
		initFile := filepath.Join(temporaryDirectory, "profile")
		if err := os.WriteFile(initFile, []byte(posixInitScript(marker)), 0o600); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("write POSIX shell configuration: %w", err)
		}

		environment = append(
			environment,
			"GOGIT_USER_ENV="+os.Getenv("ENV"),
			"ENV="+initFile,
		)
		arguments = []string{"-i"}
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
		environment,
		"CONDA_CHANGEPS1=false",
		"VIRTUAL_ENV_DISABLE_PROMPT=1",
	)

	return command, cleanup, nil
}
